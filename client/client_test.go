package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync"
	"testing"
)

func newTest(handler http.HandlerFunc) (*httptest.Server, Client) {
	s := httptest.NewServer(handler)

	u, _ := url.Parse(s.URL)
	host := u.Host

	c := NewClient(&Config{Host: host})
	return s, c
}

func TestClientKeys(t *testing.T) {
	var (
		host string
		wg   sync.WaitGroup
	)

	handler := func(rw http.ResponseWriter, r *http.Request) {
		wg.Wait()
		enc := json.NewEncoder(rw)
		switch r.RequestURI {
		case "/v1/nodes":
			enc.Encode([]Node{{Addr: host}})
		case "/v1/keys":
			enc.Encode([]string{"1", "2"})
		}
	}

	wg.Add(1)
	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()
	wg.Done()

	u, _ := url.Parse(s.URL)
	host = u.Host

	c := NewClient(&Config{Host: u.Host})
	keys, err := c.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error returned: %s", err)
	}
	if !reflect.DeepEqual(keys, []string{"1", "2"}) {
		t.Fatalf("invalid list of keys returned: %v", keys)
	}
}

func TestClientLoad(t *testing.T) {
	handler := func(rw http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/v1/keys/1" {
			enc := json.NewEncoder(rw)
			enc.Encode(Response{Data: 42})
		}
	}

	s, c := newTest(handler)
	defer s.Close()

	opts := &LoadOptions{Key: "1"}
	rec, err := c.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error returned: %s", err)
	}
	if rec.Data.(float64) != 42 {
		t.Fatalf("invalid data returned: %v", rec.Data)
	}
}

func TestClientStore(t *testing.T) {
}

func TestClientDelete(t *testing.T) {
}

func TestClientDictItem(t *testing.T) {
}

func TestClientListIndex(t *testing.T) {
}
