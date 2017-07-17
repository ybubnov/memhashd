package client

import (
	"context"
	"net"
	"net/http"
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
	Transport http.RoundTripper
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
	ListItem(context.Context, *ListIndexOptions) (*Response, error)
}

type client struct {
	httpClient *http.Client
}

func NewClient(cfg *Config) Client {
	return &client{
		httpClient: &http.Client{
			Transport: cfg.transport(),
		},
	}
}

func (c *client) Keys(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (c *client) Load(ctx context.Context, opts *LoadOptions) (*Response, error) {
	return nil, nil
}

func (c *client) Store(ctx context.Context, opts *StoreOptions) (*Response, error) {
	return nil, nil
}

func (c *client) Delete(ctx context.Context, opts *DeleteOptions) (*Response, error) {
	return nil, nil
}

func (c *client) DictItem(ctx context.Context, opts *DictItemOptions) (*Response, error) {
	return nil, nil
}

func (c *client) ListItem(ctx context.Context, opts *ListIndexOptions) (*Response, error) {
	return nil, nil
}
