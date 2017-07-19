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
		if r.Method != "GET" {
			return
		}

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
		if r.Method == "GET" && r.RequestURI == "/v1/keys/1" {
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
	ch := make(chan interface{}, 1)
	handler := func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" && r.RequestURI == "/v1/keys/2" {
			var opts StoreOptions
			dec := json.NewDecoder(r.Body)
			dec.Decode(&opts)

			ch <- opts.Data

			enc := json.NewEncoder(rw)
			enc.Encode(Response{Data: opts.Data})
		}
	}

	s, c := newTest(handler)
	defer s.Close()

	opts := &StoreOptions{Key: "2", Data: "hello"}
	rec, err := c.Store(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error returned: %s", err)
	}
	data, ok := <-ch
	if !ok || data.(string) != "hello" {
		t.Fatalf("invalid data returned: %v", rec.Data)
	}
}

func TestClientDelete(t *testing.T) {
	handler := func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.RequestURI == "/v1/keys/3" {
			enc := json.NewEncoder(rw)
			enc.Encode(Response{Meta: Meta{Index: 3}})
		}
	}

	s, c := newTest(handler)
	defer s.Close()

	opts := &DeleteOptions{Key: "3"}
	resp, err := c.Delete(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error returned: %s", err)
	}
	if resp.Meta.Index != 3 {
		t.Fatalf("wrong index value returned: %d", resp.Meta.Index)
	}
}

func TestClientDictItem(t *testing.T) {
	handler := func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.RequestURI == "/v1/keys/4/item" {
			var opts DictItemOptions
			dec := json.NewDecoder(r.Body)
			dec.Decode(&opts)

			enc := json.NewEncoder(rw)
			enc.Encode(Response{Data: opts.Item})
		}
	}

	s, c := newTest(handler)
	defer s.Close()

	opts := &DictItemOptions{Key: "4", Item: "username"}
	resp, err := c.DictItem(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error returned: %s", err)
	}
	if resp.Data.(string) != "username" {
		t.Fatalf("invalid data returned: %v", resp.Data)
	}
}

func TestClientListIndex(t *testing.T) {
	handler := func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.RequestURI == "/v1/keys/4/index" {
			var opts ListIndexOptions
			dec := json.NewDecoder(r.Body)
			dec.Decode(&opts)

			enc := json.NewEncoder(rw)
			enc.Encode(Response{Data: opts.Index})
		}
	}

	s, c := newTest(handler)
	defer s.Close()

	opts := &ListIndexOptions{Key: "4", Index: 4}
	resp, err := c.ListIndex(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error returned: %s", err)
	}
	if resp.Data.(float64) != 4 {
		t.Fatalf("invalid data returned: %v", resp.Data)
	}
}

func TestClientError(t *testing.T) {
	handler := func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusNotAcceptable)
	}

	s, c := newTest(handler)
	defer s.Close()

	_, err := c.Keys(context.Background())
	text := http.StatusText(http.StatusNotAcceptable)

	if err == nil || err.Error() != text {
		t.Fatalf("invalid error returned: %v", err)
	}
}
