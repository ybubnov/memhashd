package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONFormatter(t *testing.T) {
	f := JSONFormatter{}
	rw := httptest.NewRecorder()

	err := f.Write(rw, map[string]string{"status": "alive"}, http.StatusOK)
	if err != nil {
		t.Fatal("failed to write data in JSON format:", err)
	}

	header := rw.Header().Get(HeaderContentType)
	if header != TypeApplicationJSON {
		t.Fatal("expected Content-Type header in a response")
	}
	if rw.Body.Len() == 0 {
		t.Fatal("invalid data in body")
	}
}
