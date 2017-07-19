package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var DefaultTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

type Config struct {
	// Host is the address of the server.
	Host string

	// Transport is the Transport to use for the HTTP client.
	Transport http.RoundTripper

	// TLSConfig is a TLS configuration for the HTTP client.
	TLSConfig *tls.Config
}

func (cfg *Config) transport() http.RoundTripper {
	if cfg.Transport != nil {
		return cfg.Transport
	}
	return DefaultTransport
}

type Meta struct {
	// Index defines a record serial number. Each time the record
	// data is updated an index value is incremented.
	Index int64 `json:"index"`

	// Time to live for the record. If time to live is less than or
	// equal to zero, record won't be ever evicted from the storage.
	ExpireTime Duration `json:"expire_time"`

	// AccessedAt defines a moment when the record was accessed last
	// time.
	AccessedAt time.Time `json:"accessed_at"`

	// CreatedAt defines a moment when the record was stored into a
	// hash table.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt defines a moment when the record was updated. This
	// value is changed on update of the resource record.
	UpdatedAt time.Time `json:"updated_at"`
}

type Node struct {
	ID   string `json:"id"`
	Addr string `json:"addr"`
}

type Error struct {
	Text string `json:"text"`
}

type Response struct {
	Action string `json:"action"`

	Meta Meta `json:"meta,omitempty"`

	Data interface{} `json:"data,omitempty"`

	Node Node `json:"node"`
}

type LoadOptions struct {
	Key string `json:"-"`
}

type StoreOptions struct {
	Key        string      `json:"-"`
	Data       interface{} `json:"data"`
	ExpireTime Duration    `json:"expire_time"`
}

type DeleteOptions struct {
	Key string `json:"-"`
}

type DictItemOptions struct {
	Key  string      `json:"-"`
	Item interface{} `json:"item"`
}

type ListIndexOptions struct {
	Key   string `json:"-"`
	Index uint64 `json:"index"`
}

type Client interface {
	Keys(context.Context) ([]string, error)
	Load(context.Context, *LoadOptions) (*Response, error)
	Store(context.Context, *StoreOptions) (*Response, error)
	Delete(context.Context, *DeleteOptions) (*Response, error)
	DictItem(context.Context, *DictItemOptions) (*Response, error)
	ListIndex(context.Context, *ListIndexOptions) (*Response, error)
}

type client struct {
	host       string
	tlsConfig  *tls.Config
	httpClient *http.Client
}

func NewClient(config *Config) Client {
	return &client{
		host: config.Host,
		httpClient: &http.Client{
			Transport: config.transport(),
		},
	}
}

func (c *client) scheme() string {
	if c.tlsConfig != nil {
		return "https"
	}
	return "http"
}

func (c *client) do(ctx context.Context, method string,
	u *url.URL, in, out interface{}) error {

	var (
		b   []byte
		err error
	)

	if in != nil {
		b, err = json.Marshal(in)
		if err != nil {
			return err
		}
	}
	body := bytes.NewReader(b)
	req, err := http.NewRequest("GET", u.String(), body)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	// Decode the list of nodes from the body of the response.
	defer resp.Body.Close()
	if out == nil {
		return nil
	}

	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(out)
}

// nodes returns a list of the nodes in a cluster.
func (c *client) nodes(ctx context.Context) (nodes []Node, err error) {
	err = c.do(ctx, "GET", c.urlOf("/v1/nodes"), nil, &nodes)
	return nodes, err
}

func (c *client) joinerr(ch <-chan error) error {
	var text []string
	for err := range ch {
		if err != nil {
			text = append(text, err.Error())
		}
	}
	if text != nil {
		return fmt.Errorf(strings.Join(text, ", "))
	}
	return nil
}

func (c *client) urlOf(path string) *url.URL {
	return &url.URL{Scheme: c.scheme(), Host: c.host, Path: path}
}

func (c *client) Keys(ctx context.Context) ([]string, error) {
	nodes, err := c.nodes(ctx)
	if err != nil {
		return nil, err
	}

	var (
		wg   sync.WaitGroup
		keys = make(chan []string, len(nodes))
		errs = make(chan error, len(nodes))
	)

	retrieve := func(n *Node) {
		u := &url.URL{
			Scheme: c.scheme(),
			Host:   n.Addr,
			Path:   "/v1/keys",
		}
		defer wg.Done()
		var ks []string

		errs <- c.do(ctx, "GET", u, nil, &ks)
		keys <- ks
	}
	for _, node := range nodes {
		wg.Add(1)
		go retrieve(&node)
	}

	wg.Wait()
	close(keys)
	close(errs)

	if err := c.joinerr(errs); err != nil {
		return nil, err
	}
	var kys []string
	for ks := range keys {
		kys = append(kys, ks...)
	}
	return kys, nil
}

func (c *client) Load(ctx context.Context,
	opts *LoadOptions) (resp *Response, err error) {

	resp = new(Response)
	path := fmt.Sprintf("/v1/keys/%s", opts.Key)
	err = c.do(ctx, "GET", c.urlOf(path), nil, resp)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (c *client) Store(ctx context.Context,
	opts *StoreOptions) (resp *Response, err error) {

	resp = new(Response)
	path := fmt.Sprintf("/v1/keys/%s", opts.Key)
	err = c.do(ctx, "PUT", c.urlOf(path), opts, resp)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (c *client) Delete(ctx context.Context,
	opts *DeleteOptions) (resp *Response, err error) {

	resp = new(Response)
	path := fmt.Sprintf("/v1/keys/%s", opts.Key)
	err = c.do(ctx, "DELETE", c.urlOf(path), nil, resp)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (c *client) DictItem(ctx context.Context,
	opts *DictItemOptions) (resp *Response, err error) {

	resp = new(Response)
	path := fmt.Sprintf("/v1/keys/%s/item", opts.Key)
	err = c.do(ctx, "GET", c.urlOf(path), opts, resp)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (c *client) ListIndex(ctx context.Context,
	opts *ListIndexOptions) (resp *Response, err error) {

	resp = new(Response)
	path := fmt.Sprintf("/v1/keys/%s/index", opts.Key)
	err = c.do(ctx, "GET", c.urlOf(path), opts, resp)
	if err != nil {
		return nil, err
	}
	return resp, err
}
