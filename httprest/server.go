package httprest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"memhashd/client"
	"memhashd/container/store"
	"memhashd/httprest/httputil"
	"memhashd/server"
	"memhashd/system/log"
	"memhashd/system/uuid"
)

// Server is an HTTP API server, it provides an access to the key-value
// storage.
type Server struct {
	laddr  net.Addr
	mux    *httputil.ServeMux
	server server.Server
	ctx    context.Context
}

// Config is a configuration of the HTTP API server.
type Config struct {
	// Server is an instance of the key-value storage.
	Server server.Server

	// LocalAddr is an address to listen for incoming requests.
	LocalAddr net.Addr

	// Context is a context of the server, it defines a lifetime
	// of each request handled by the server.
	Context context.Context
}

func (c *Config) context() context.Context {
	if c.Context != nil {
		return c.Context
	}
	return context.Background()
}

// NewServer creates a new instance of the Server.
func NewServer(config *Config) *Server {
	s := &Server{
		laddr:  config.LocalAddr,
		mux:    httputil.NewServeMux(),
		server: config.Server,
		ctx:    config.context(),
	}

	s.mux.HandleFunc("GET", "/v1/keys", s.keysHandler)
	s.mux.HandleFunc("GET", "/v1/keys/{key}", s.loadHandler)
	s.mux.HandleFunc("GET", "/v1/keys/{key}/index", s.indexHandler)
	s.mux.HandleFunc("GET", "/v1/keys/{key}/item", s.itemHandler)
	s.mux.HandleFunc("PUT", "/v1/keys/{key}", s.storeHandler)
	s.mux.HandleFunc("DELETE", "/v1/keys/{key}", s.deleteHandler)
	s.mux.HandleFunc("GET", "/v1/nodes", s.nodesHandler)
	return s
}

func (s *Server) readReq(rw http.ResponseWriter, r *http.Request,
	val interface{}) error {

	rf, wf, err := httputil.Format(rw, r)
	if err != nil {
		return err
	}
	if err := rf.Read(r, val); err != nil {
		const text = "failed to read request body, %s"
		log.ErrorLogf("server/READ_REQUEST", text, err)

		body := client.Error{fmt.Sprintf(text, err)}
		wf.Write(rw, body, http.StatusBadRequest)
		return err
	}
	return nil
}

// metaOf returns a record metadata in a client format.
func (s *Server) metaOf(resp *server.Response) client.Meta {
	return client.Meta{
		Index:      resp.Record.Meta.Index,
		ExpireTime: client.Duration(resp.Record.Meta.ExpireTime),
		AccessedAt: resp.Record.Meta.AccessedAt,
		CreatedAt:  resp.Record.Meta.CreatedAt,
		UpdatedAt:  resp.Record.Meta.UpdatedAt,
	}
}

// nodeOf return a node information in a client format.
func (s *Server) nodeOf(resp *server.Response) client.Node {
	return client.Node{ID: s.server.ID(), Addr: resp.Node.Addr.String()}
}

// keysHandler returns a list of keys stored on all nodes in a cluster.
func (s *Server) keysHandler(rw http.ResponseWriter, r *http.Request) {
	_, wf, err := httputil.Format(rw, r)
	if err != nil {
		return
	}

	req := &store.RequestKeys{ID: uuid.New()}
	resp := s.server.Do(s.ctx, req)
	if resp.Err() != nil {
		const text = "unable to load keys, %s"
		body := client.Error{fmt.Sprintf(text, resp.Err())}

		log.ErrorLogf("server/KEYS_HANDLER",
			"%s failed, %s", req.ID, resp.Err())
		wf.Write(rw, body, resp.Status)
		return
	}

	keys, ok := resp.Record.Data.([]string)
	if !ok {
		const text = "invalid type of keys"
		log.ErrorLogf("server/KEYS_HANDLER", text)

		body := client.Error{text}
		wf.Write(rw, body, http.StatusInternalServerError)
		return
	}

	// Force the server return empty list instead of nil.
	if keys == nil {
		keys = make([]string, 0)
	}
	wf.Write(rw, keys, http.StatusOK)
}

// loadHandler loads a requested data from the store (depending on
// requested action, it can return a partial data, like item in a list
// or dictionary).
func (s *Server) loadHandler(rw http.ResponseWriter, r *http.Request) {
	key := httputil.Param(r, "key")
	_, wf, err := httputil.Format(rw, r)
	if err != nil {
		return
	}

	// Create a new load request, assign an identifier to it, for easy
	// tracking in the logs of the application.
	req := &store.RequestLoad{ID: uuid.New(), Key: key}
	resp := s.server.Do(s.ctx, req)
	if resp.Err() != nil {
		const text = "unable to load %s key, %s"
		body := client.Error{fmt.Sprintf(text, req.Key, resp.Err())}

		log.ErrorLogf("server/LOAD_HANDLER",
			"%s failed, %s", req.ID, resp.Err())
		wf.Write(rw, body, resp.Status)
		return
	}

	// Construct a response of a successful request processing and
	// return it back to the client.
	cresp := client.Response{
		Action: "load",
		Data:   resp.Record.Data,
		Node:   s.nodeOf(&resp),
		Meta:   s.metaOf(&resp),
	}
	wf.Write(rw, cresp, http.StatusOK)
}

// storeHandler stores a given record in a key-value storage. It
// returns a node where a record was created, creation and update time.
func (s *Server) storeHandler(rw http.ResponseWriter, r *http.Request) {
	key := httputil.Param(r, "key")
	_, wf, err := httputil.Format(rw, r)
	if err != nil {
		return
	}

	var opts client.StoreOptions
	if err := s.readReq(rw, r, &opts); err != nil {
		return
	}

	req := &store.RequestStore{
		ID:  uuid.New(),
		Key: key, Data: opts.Data,
		ExpireTime: time.Duration(opts.ExpireTime),
	}
	resp := s.server.Do(s.ctx, req)
	if resp.Err() != nil {
		const text = "unable to store %s key, %s"
		body := client.Error{fmt.Sprintf(text, req.Key, resp.Err())}

		log.ErrorLogf("server/STORE_HANDLER",
			"%s failed, %s", req.ID, resp.Err())
		wf.Write(rw, body, resp.Status)
		return
	}

	cresp := client.Response{
		Action: "store",
		Data:   resp.Record.Data,
		Node:   s.nodeOf(&resp),
		Meta:   s.metaOf(&resp),
	}
	wf.Write(rw, cresp, http.StatusOK)
}

// deleteHandler deletes the requested key from the storage. It returns
// a node where a record was removed. Return does not return an error if
// record does not exist.
func (s *Server) deleteHandler(rw http.ResponseWriter, r *http.Request) {
	key := httputil.Param(r, "key")
	_, wf, err := httputil.Format(rw, r)
	if err != nil {
		return
	}

	// Create a new delete request, assign an identifier to it for
	// tracking.
	req := &store.RequestDelete{ID: uuid.New(), Key: key}
	resp := s.server.Do(s.ctx, req)
	if resp.Err() != nil {
		const text = "unable to delete %s key, %s"
		body := client.Error{fmt.Sprintf(text, req.Key, resp.Err())}

		log.ErrorLogf("server/DELETE_HANDLER",
			"%s failed, %s", req.ID, resp.Err())
		wf.Write(rw, body, resp.Status)
		return
	}

	cresp := client.Response{
		Action: "delete",
		Node:   s.nodeOf(&resp),
		Meta:   s.metaOf(&resp),
	}
	wf.Write(rw, cresp, http.StatusOK)
}

// indexHandler returns a value at the given position. Method returns error
// when index is out of array bounds or requested key is not an array.
func (s *Server) indexHandler(rw http.ResponseWriter, r *http.Request) {
	key := httputil.Param(r, "key")
	_, wf, err := httputil.Format(rw, r)
	if err != nil {
		return
	}

	// Parse a requested parameters of the list index function.
	var opts client.ListIndexOptions
	if err := s.readReq(rw, r, &opts); err != nil {
		return
	}

	req := &store.RequestListIndex{
		ID: uuid.New(), Key: key, Index: opts.Index,
	}
	resp := s.server.Do(s.ctx, req)
	if resp.Err() != nil {
		const text = "unable to load value, %s"
		body := client.Error{fmt.Sprintf(text, resp.Err())}

		log.ErrorLogf("server/INDEX_HANDLER",
			"%s failed, %s", req.ID, resp.Err())
		wf.Write(rw, body, resp.Status)
		return
	}

	cresp := client.Response{
		Action: "index",
		Data:   resp.Record.Data,
		Node:   s.nodeOf(&resp),
		Meta:   s.metaOf(&resp),
	}
	wf.Write(rw, cresp, http.StatusOK)
}

// itemHandler returns an item in the dictionary stored at a specified
// key. Method returns an error, when the target type is not a dictionary.
func (s *Server) itemHandler(rw http.ResponseWriter, r *http.Request) {
	key := httputil.Param(r, "key")
	_, wf, err := httputil.Format(rw, r)
	if err != nil {
		return
	}

	var opts client.DictItemOptions
	if err := s.readReq(rw, r, &opts); err != nil {
		return
	}

	// Retrieve an item of the given dictionary.
	req := &store.RequestDictItem{
		ID: uuid.New(), Key: key, Item: opts.Item,
	}
	resp := s.server.Do(s.ctx, req)
	if resp.Err() != nil {
		const text = "unable to load value, %s"
		body := client.Error{fmt.Sprintf(text, resp.Err())}

		log.ErrorLogf("server/ITEM_HANDLER",
			"%s failed, %s", req.ID, resp.Err())
		wf.Write(rw, body, resp.Status)
		return
	}

	cresp := client.Response{
		Action: "item",
		Data:   resp.Record.Data,
		Node:   s.nodeOf(&resp),
		Meta:   s.metaOf(&resp),
	}
	wf.Write(rw, cresp, http.StatusOK)
}

// nodesHandler returns a list of nodes in a cluster, so the clients
// can easily communicate with each one.
func (s *Server) nodesHandler(rw http.ResponseWriter, r *http.Request) {
	var nodes []*client.Node
	for _, node := range s.server.Nodes() {
		nodes = append(nodes, &client.Node{
			ID:   node.ID,
			Addr: node.Addr.String(),
		})
	}

	_, wf, err := httputil.Format(rw, r)
	if err != nil {
		return
	}
	wf.Write(rw, nodes, http.StatusOK)
}

// ListenAndServe starts an HTTP server at the configured endpoint.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.laddr.String(), s.mux)
}
