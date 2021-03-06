package httputil

import (
	"net/http"
	"regexp"
	"strings"
	"sync"
)

var (
	paramRegexp = regexp.MustCompile(`\{([^\}]+)\}`)

	defaultMux = NewServeMux()
)

// Handle registers a handler for the given method and pattern in a
// default HTTP multiplexer.
func Handle(method, pattern string, handler http.Handler) {
	err := defaultMux.Handle(method, pattern, handler)
	if err != nil {
		panic(err)
	}
}

// HandleFunc registers a handler function for the given method and
// pattern in a default HTTP multiplexer.
func HandleFunc(method, pattern string, handler http.HandlerFunc) {
	err := defaultMux.HandleFunc(method, pattern, handler)
	if err != nil {
		panic(err)
	}
}

// muxEntry is an entry of the multiplexer.
type muxEntry struct {
	h       http.Handler
	pattern *regexp.Regexp
	params  []string
}

// responseWriter is an implementation of the http.ResponseWriter
// interface that marks the persists the fact of writing a header.
type responseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

// WriteHeader implements http.ResponseWriter interface.
func (rw *responseWriter) WriteHeader(code int) {
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// ServeMux is a HTTP request multiplexer. It matches incoming requests
// by method and target URL and executes a registered handler for them.
type ServeMux struct {
	m        map[string][]muxEntry
	f        []http.Handler
	mu       sync.RWMutex
	NotFound http.Handler
}

// NewServeMux creates a new instance of the ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{m: make(map[string][]muxEntry)}
}

// HandleFilter registers a filter that will be executed before each
// registered handler. When filter writes a response header, processing
// of the requests stops.
func (mux *ServeMux) HandleFilter(handler http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	mux.f = append(mux.f, handler)
}

// HandleFilterFunc registers a filter function that will be executed
// before each registered handler.
func (mux *ServeMux) HandleFilterFunc(handler http.HandlerFunc) {
	mux.HandleFilter(handler)
}

// Handle registers a handler for the given method and URL pattern.
func (mux *ServeMux) Handle(method, pattern string, handler http.Handler) error {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	path := strings.Split(pattern, "/")
	var params []string

	for index, p := range path {
		match := paramRegexp.FindStringSubmatch(p)
		if match == nil {
			continue
		}

		params = append(params, match[1])
		path[index] = "([^/]+)"
	}

	pattern = strings.Join(path, "/")
	r, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	entry := muxEntry{handler, r, params}
	mux.m[method] = append(mux.m[method], entry)
	return nil
}

// HandleFunc registers a handler function for the given method and URL
// pattern.
func (mux *ServeMux) HandleFunc(method, pattern string, handler http.HandlerFunc) error {
	return mux.Handle(method, pattern, handler)
}

// ServeHTTP dispatches the requests to the handler whose method
// and pattern most closely matches the request URL.
func (mux *ServeMux) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	mux.mu.RLock()
	w := &responseWriter{rw, false}

	for _, f := range mux.f {
		f.ServeHTTP(w, r)
		if w.wroteHeader {
			mux.mu.RUnlock()
			return
		}
	}

	entries, ok := mux.m[r.Method]
	mux.mu.RUnlock()
	if !ok {
		mux.notAllowed(rw, r)
		return
	}

	for _, entry := range entries {
		match := entry.pattern.FindStringSubmatch(r.URL.Path)
		if match == nil {
			continue
		}
		if match[0] != r.URL.Path {
			continue
		}
		match = match[1:]
		if len(match) != len(entry.params) {
			continue
		}
		for i := range match {
			param := entry.params[i] + "=" + match[i]
			r.URL.RawQuery = param + "&" + r.URL.RawQuery
		}

		entry.h.ServeHTTP(rw, r)
		return
	}

	mux.notFound(rw, r)
}

func (mux *ServeMux) notFound(rw http.ResponseWriter, r *http.Request) {
	if mux.NotFound != nil {
		mux.NotFound.ServeHTTP(rw, r)
	} else {
		http.Error(rw, "page not found", http.StatusNotFound)
	}
}

func (mux *ServeMux) notAllowed(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
}

// Params returns a list of URL parameters.
func Params(r *http.Request, s ...string) (p []string) {
	for _, key := range s {
		p = append(p, r.URL.Query().Get(key))
	}
	return
}

// Param returns a single URL parameter.
func Param(r *http.Request, s string) string {
	return r.URL.Query().Get(s)
}
