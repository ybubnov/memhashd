package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeMux(t *testing.T) {
	mux := NewServeMux()

	tenantsIndex := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("tenant index"))
	}

	tenantsShow := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("tenant show"))
	}

	usersIndex := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("user index"))
	}

	usersShow := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("user show"))
	}

	notFound := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("not found"))
	}

	mux.HandleFunc("GET", "/tenants", tenantsIndex)
	mux.HandleFunc("GET", "/tenants/{tenant}", tenantsShow)
	mux.HandleFunc("GET", "/tenants/{tenant}/users", usersIndex)
	mux.HandleFunc("GET", "/tenants/{tenant}/users/{user}", usersShow)
	mux.NotFound = http.HandlerFunc(notFound)

	test := func(url, body string) {
		r, _ := http.NewRequest("GET", url, nil)
		rw := httptest.NewRecorder()

		mux.ServeHTTP(rw, r)

		if rw.Body.String() != body {
			t.Fatal("Failed to send request to right handler:", url, rw.Body.String())
		}
	}

	test("/tenants", "tenant index")
	test("/tenants/123", "tenant show")
	test("/tenants/123/users", "user index")
	test("/tenants/123/users/456", "user show")
	test("/tenants/123/users/456/logout", "not found")
	test("/customers/123", "not found")
}

func TestServeMuxFilter(t *testing.T) {
	var filter1, filter2 bool
	mux := NewServeMux()

	mux.HandleFunc("GET", "/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("handler"))
	})

	mux.HandleFilterFunc(func(rw http.ResponseWriter, r *http.Request) {
		if filter1 {
			rw.WriteHeader(http.StatusNotAcceptable)
			rw.Write([]byte("filter handler 1"))
		}
	})

	mux.HandleFilterFunc(func(rw http.ResponseWriter, r *http.Request) {
		if filter2 {
			rw.WriteHeader(http.StatusUnsupportedMediaType)
			rw.Write([]byte("filter handler 2"))
		}
	})

	test := func(status int, body string) {
		r, _ := http.NewRequest("GET", "/", nil)
		rw := httptest.NewRecorder()

		mux.ServeHTTP(rw, r)

		if rw.Code != status {
			t.Fatal("Failed to return right status:", rw.Code)
		}

		if rw.Body.String() != body {
			t.Fatal("Failed to properly filter request:", rw.Body.String())
		}
	}

	test(http.StatusOK, "handler")

	filter1, filter2 = true, false
	test(http.StatusNotAcceptable, "filter handler 1")

	filter1, filter2 = false, true
	test(http.StatusUnsupportedMediaType, "filter handler 2")

	filter1, filter2 = false, false
	test(http.StatusOK, "handler")
}