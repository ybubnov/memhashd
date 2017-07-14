package format

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"memhashd/httprest/httputil"
)

func TestJSONFormatter(t *testing.T) {
	f := JSONFormatter{}
	rw := httptest.NewRecorder()

	err := f.Write(rw, map[string]string{"status": "alive"}, http.StatusOK)
	if err != nil {
		t.Fatal("Failed to write data in JSON format:", err)
	}

	header := rw.Header().Get(httputil.HeaderContentType)
	if header != httputil.TypeApplicationJSON {
		t.Fatal("Expected Content-Type header in a response")
	}

	if rw.Body.Len() == 0 {
		t.Fatal("Invalid data in body")
	}
}
