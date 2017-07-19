package httprest

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"memhashd/client"
	"memhashd/container/hash"
	"memhashd/container/store"
	"memhashd/server"
)

type stubServer struct {
	Request  store.Request
	Response server.Response
}

func (s *stubServer) ID() string   { return "" }
func (s *stubServer) Start() error { return nil }
func (s *stubServer) Stop() error  { return nil }

func (s *stubServer) Nodes() server.Nodes {
	return server.Nodes{{Addr: &net.TCPAddr{
		IP: net.ParseIP("127.0.0.1"), Port: 2371,
	}}}
}

func (s *stubServer) Do(_ context.Context, req store.Request) server.Response {
	s.Request = req
	return s.Response
}

func assertResponse(t *testing.T, rw *httptest.ResponseRecorder,
	action string) *client.Response {

	var resp client.Response
	if rw.Code != http.StatusOK {
		t.Fatalf("wrong status code returned: %d", rw.Code)
	}

	json.Unmarshal(rw.Body.Bytes(), &resp)
	if resp.Action != action {
		t.Fatalf("invalid response action: %s", resp.Action)
	}
	return &resp
}

func assertError(t *testing.T, rw *httptest.ResponseRecorder,
	status int, body string) {

	if rw.Code != status {
		t.Fatalf("wrong status code returned: %d", rw.Code)
	}
	if strings.TrimRight(rw.Body.String(), "\n") != body {
		t.Fatalf("wrong body returned: %s", rw.Body.String())
	}
}

func TestKeysHandler(t *testing.T) {
	keys := []string{"1", "2", "3"}
	resp := server.Response{Record: hash.Record{Data: keys}}

	stub := &stubServer{Response: resp}
	s := NewServer(&Config{Server: stub})

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/keys", nil)

	s.keysHandler(rw, req)
	var res []string

	json.Unmarshal(rw.Body.Bytes(), &res)
	if !reflect.DeepEqual(keys, res) {
		t.Fatalf("invalid keys are returned: %s", res)
	}
	if rw.Code != http.StatusOK {
		t.Fatalf("wrong status code returned: %d", rw.Code)
	}

	stub.Response = server.Response{
		Error: "boom", Status: http.StatusConflict}
	rw = httptest.NewRecorder()
	s.keysHandler(rw, req)

	body := "{\"text\":\"unable to load keys, boom\"}"
	assertError(t, rw, stub.Response.Status, body)

	// Response on the keys request should be always a list of
	// strings, therefore internal server error is expected here.
	stub.Response = server.Response{Record: hash.Record{Data: 1}}
	rw = httptest.NewRecorder()
	s.keysHandler(rw, req)

	body = "{\"text\":\"invalid type of keys\"}"
	assertError(t, rw, http.StatusInternalServerError, body)
}

func TestLoadHandler(t *testing.T) {
	res := server.Response{Record: hash.Record{Data: 42}}
	stub := &stubServer{Response: res}
	s := NewServer(&Config{Server: stub})

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/keys?key=1", nil)

	s.loadHandler(rw, req)
	resp := assertResponse(t, rw, store.ActionLoad)

	if resp.Data.(float64) != 42 {
		t.Fatalf("invalid data returned: %v", resp.Data)
	}

	stub.Response = server.Response{
		Error: "zap!", Status: http.StatusTeapot}
	rw = httptest.NewRecorder()
	s.loadHandler(rw, req)

	body := "{\"text\":\"unable to load 1 key, zap!\"}"
	assertError(t, rw, stub.Response.Status, body)
}

func TestStoreHandler(t *testing.T) {
	res := server.Response{
		Record: hash.Record{
			Data: 42, Meta: hash.Meta{CreatedAt: time.Now()},
		},
	}
	stub := &stubServer{Response: res}
	s := NewServer(&Config{Server: stub})

	rw := httptest.NewRecorder()
	rd := strings.NewReader(`{"data": 42}`)
	req := httptest.NewRequest("PUT", "/v1/keys?key=1", rd)

	s.storeHandler(rw, req)
	resp := assertResponse(t, rw, store.ActionStore)

	if resp.Meta.CreatedAt.IsZero() {
		t.Fatalf("creation time of record should be non-zero")
	}
	if stub.Request.Hash() != "1" {
		t.Fatalf("invalid hash of the request: %s", stub.Request.Hash())
	}

	stub.Response = server.Response{
		Error: "oops", Status: http.StatusLocked}

	rw = httptest.NewRecorder()
	req.Body = ioutil.NopCloser(strings.NewReader(`{"data": 44}`))

	s.storeHandler(rw, req)
	body := "{\"text\":\"unable to store 1 key, oops\"}"
	assertError(t, rw, stub.Response.Status, body)
}

func TestDeleteHandler(t *testing.T) {
	stub := &stubServer{}
	s := NewServer(&Config{Server: stub})

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/v1/keys?key=3", nil)

	s.deleteHandler(rw, req)
	assertResponse(t, rw, store.ActionDelete)

	stub.Response = server.Response{
		Error: "bam", Status: http.StatusGone}

	rw = httptest.NewRecorder()
	s.deleteHandler(rw, req)

	body := "{\"text\":\"unable to delete 3 key, bam\"}"
	assertError(t, rw, stub.Response.Status, body)
}

func TestIndexHandler(t *testing.T) {
	res := server.Response{Record: hash.Record{Data: 3}}
	stub := &stubServer{Response: res}
	s := NewServer(&Config{Server: stub})

	body := strings.NewReader(`{"index": 3}`)
	req := httptest.NewRequest("GET", "/v1/keys?key=2", body)

	rw := httptest.NewRecorder()

	s.indexHandler(rw, req)
	assertResponse(t, rw, store.ActionListIndex)

	stub.Response = server.Response{
		Error: "wak", Status: http.StatusForbidden}

	body = strings.NewReader(`{"index": 3}`)
	req = httptest.NewRequest("GET", "/v1/keys?key=2", body)

	rw = httptest.NewRecorder()
	s.indexHandler(rw, req)

	bodyText := "{\"text\":\"unable to load value, wak\"}"
	assertError(t, rw, stub.Response.Status, bodyText)
}

func TestItemHandler(t *testing.T) {
	res := server.Response{Record: hash.Record{Data: 42}}
	stub := &stubServer{Response: res}
	s := NewServer(&Config{Server: stub})

	body := strings.NewReader(`{"item": "aa"}`)
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/keys?key=4", body)

	s.itemHandler(rw, req)
	assertResponse(t, rw, store.ActionDictItem)

	stub.Response = server.Response{
		Error: "bang", Status: http.StatusNotFound}

	rw = httptest.NewRecorder()
	body = strings.NewReader(`{"item": "no1"}`)
	req = httptest.NewRequest("GET", "/v1/keys?key=5", body)

	s.itemHandler(rw, req)
	bodyText := "{\"text\":\"unable to load value, bang\"}"
	assertError(t, rw, stub.Response.Status, bodyText)
}
