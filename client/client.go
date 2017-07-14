package client

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"
)

var DefaultTransport RoundTripper = &Transport{
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
	Transport http.Transport
}

func (cfg *Config) transport() http.Transport {
	if cfg.Transport != nil {
		return cfg.Transport
	}
	return DefaultTransport
}

type Meta struct {
	// Index defines a record serial number. Each time the record
	// data is updated an index value is incremented.
	Index int64

	// Time to live for the record. If time to live is less than or
	// equal to zero, record won't be ever evicted from the storage.
	ExpireTime time.Duration

	// AccessedAt defines a moment when the record was accessed last
	// time.
	AccessedAt time.Time

	// CreatedAt defines a moment when the record was stored into a
	// hash table.
	CreatedAt time.Time

	// UpdatedAt defines a moment when the record was updated. This
	// value is changed on update of the resource record.
	UpdatedAt time.Time
}

type Response struct {
	Meta Meta

	Data interface{}
}

// KeysOptions defines options for a keys request. It is empty
// now and reserved for future use.
type KeysOptions struct {
}

type LoadOptions struct {
	Key   string
	Value json.Marshaler
}

type StoreOptions struct {
	Key        string
	Value      json.Marshaler
	ExpireTime time.Duration
}

type DeleteOptions struct {
	Key string
}

type DictItemOptions struct {
	Key  string
	Item interface{}
}

type ListItemOptions struct {
	Key  string
	Item uint64
}

type Client interface {
	Keys(context.Context, *KeysOptions) (*Response, error)
	Load(context.Context, *LoadOptions) (*Response, error)
	Store(context.Context, *StoreOptions) (*Response, error)
	Delete(context.Context, *DeleteOptions) (*Response, error)
	DictItem(context.Context, *DictItemOptions) (*Response, error)
	ListItem(context.Context, *ListItemOptions) (*Response, error)
}

type client struct {
	httpClient *http.Client
}

func NewClient(cfg *Config) Client {
	return &client{
		httpClient: http.Client{
			Transport: cfg.transport(),
		},
	}
}

func (c *client) Keys(ctx context.Context, opts *KeysOptions) (*Response, error) {
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

func (c *client) DictItem(ctx context.Context, opts *DictItemOptions) (*Respsonse, error) {
	return nil, nil
}

func (c *client) ListItem(ctx context.Context, opts *ListItemOptions) error {
	return nil
}
