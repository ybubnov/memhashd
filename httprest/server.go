package httprest

import (
	"net/http"

	"memhashd/httprest/httputil"
	"memhashd/server"
)

type Server struct {
	mux    *httputil.ServeMux
	server server.Server
}

type Config struct {
	Server server.Server
}

type Request struct {
	Action string
}

func NewServer(config *Config) *Server {
	s := &Server{
		mux:    httputil.NewServeMux(),
		server: config.Server,
	}

	s.mux.HandleFunc("GET", "/v1/keys", s.keysHandler)
	s.mux.HandleFunc("GET", "/v1/keys/{key}", s.keyHandler)
	s.mux.HandleFunc("PUT", "/v1/keys/{key}", s.storeHandler)
	s.mux.HandleFunc("DELETE", "/v1/keys/{key}", s.deleteHandler)

	return s
}

func (s *Server) keysHandler(rw http.ResponseWriter, r *http.Request) {
}

func (s *Server) keyHandler(rw http.ResponseWriter, r *http.Request) {
	//key := httputil.Param(r, "key")
}

func (s *Server) loadHandler(rw http.ResponseWriter, r *http.Request) {
	//key := httputil.Param(r, "key")
}

func (s *Server) storeHandler(rw http.ResponseWriter, r *http.Request) {
	//key := httputil.Param(r, "key")
}

func (s *Server) deleteHandler(rw http.ResponseWriter, r *http.Request) {
	//key := httputil.Param(r, "key")
}

func (s *Server) ListenAndServe() error {
	http.ListenAndServe(":8080", s.mux)
	return nil
}
