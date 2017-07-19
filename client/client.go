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

// DefaultTransport is a default configuration of the Transport.
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

// Config is a configuration of the client.
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

// Meta is a metadata about the record in a hash.
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

// Node is a node in a cluster.
type Node struct {
	// ID is a node unique identifier.
	ID string `json:"id"`

	// Addr is an endpoint of the node in a cluster (a port:host pair).
	Addr string `json:"addr"`
}

// Error is a server error, usually it is returned when the user
// provided incorrect parameters of the request or data for the
// operation is not valid (e.g. dict item for string).
type Error struct {
	// Text is a text of the error.
	Text string `json:"text"`
}

// Error implements error interface.
func (e *Error) Error() string {
	return e.Text
}

// Response defines a response being returned by the server in case
// of the successful handling of the request.
type Response struct {
	// Action specifies a name of the action being processed.
	Action string `json:"action"`

	// Meta defines a metadata about the returned record.
	Meta Meta `json:"meta,omitempty"`

	// Data is the data returned from the key-value storage.
	Data interface{} `json:"data,omitempty"`

	// Node is a node of the cluster that stores the requested data.
	//
	// For the sake of communication speed, users could use it to
	// create an additional client for communication with nodes
	// directly.
	Node Node `json:"node"`
}

// LoadOptions defines parameters of the load request.
type LoadOptions struct {
	// Key is a key to load.
	Key string `json:"-"`
}

// StoreOptions defines parameters of the store request.
type StoreOptions struct {
	// Key is a key to store.
	Key string `json:"-"`
	// Data defines a data to store.
	Data interface{} `json:"data"`
	// ExpireTime specifies an expiration of the data.
	ExpireTime Duration `json:"expire_time"`
}

// DeleteOptions defines parameters for the delete request.
type DeleteOptions struct {
	// Key is a key to delete.
	Key string `json:"-"`
}

// DictItemOptions defines parameters for the dict item request.
type DictItemOptions struct {
	// Key is a key to use to retrieve the data.
	Key string `json:"-"`
	// Item is an item in a dictionary used to access to.
	Item interface{} `json:"item"`
}

// ListIndexOptions defines parameters for the list index request.
type ListIndexOptions struct {
	// Key is a key to use to retrieve the data.
	Key string `json:"-"`
	// Index is an index in a list used to access to.
	Index uint64 `json:"index"`
}

// Client describes types to communicate with a key-value storage.
type Client interface {
	// Keys returns a list of keys.
	Keys(context.Context) ([]string, error)

	// Load returns a record persisted under the given key.
	Load(context.Context, *LoadOptions) (*Response, error)

	// Store persists the record under the given key.
	Store(context.Context, *StoreOptions) (*Response, error)

	// Delete removes the record persisted under the given key.
	Delete(context.Context, *DeleteOptions) (*Response, error)

	// DictItem returns an element of the dictionary persisted under the
	// given key and item.
	DictItem(context.Context, *DictItemOptions) (*Response, error)

	// ListIndex returns an element of the dictionary persisted under the
	// given key and index.
	ListIndex(context.Context, *ListIndexOptions) (*Response, error)
}

// client is a key-value storage client.
type client struct {
	host       string
	tlsConfig  *tls.Config
	httpClient *http.Client
}

// NewClient creates a new instance of the Client.
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
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	// Decode the list of nodes from the body of the response.
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	// If server returned non-zero status, the response body is treated
	// as a error message, which will be returned to the user.
	if resp.StatusCode != http.StatusOK {
		// Server could return a response without a body (in case of
		// unexpected errors or dramatic failures), therefore double
		// check that there is something in a response body.
		if resp.ContentLength == 0 {
			return &Error{http.StatusText(resp.StatusCode)}
		}

		// Decode the error returned by the server and simply forward it
		// to the client without any modification.
		var re Error
		if err := decoder.Decode(&re); err != nil {
			return err
		}

		return &re
	}

	if out == nil {
		return nil
	}
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

// Keys implements Client interface. It retrieve a list of the nodes
// from the configured server and then polls each one in parallel to
// retrieve the list of keys. Result will be consolidated into a single
// list.
//
// This operation is extremly fragile as error on single node causes
// an error of the whole operation.
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

	// Retreive a list of keys from the given node.
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
	// Call each node in parallel and then aggregate the results
	// into a single list of keys.
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

// Load implements Client interface.
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

// Store implements Client interface.
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

// Delete implements Client interface.
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

// DictItem implements Client interface.
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

// ListIndex implement Client interface.
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
