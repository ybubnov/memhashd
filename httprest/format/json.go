package format

import (
	"encoding/json"
	"net/http"

	"memhashd/httprest/httputil"
)

func init() {
	// Register formatter for application/json and */* types.
	Register(httputil.TypeApplicationJSON, &JSONFormatter{})
	Register(httputil.TypeAny, &JSONFormatter{})
}

// JSONFormatter formats data into JSON.
type JSONFormatter struct{}

// Read implements Formatter interface.
func (f *JSONFormatter) Read(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

// Write implements Formatter interface.
func (f *JSONFormatter) Write(w http.ResponseWriter, v interface{}, status int) error {
	w.Header().Set(httputil.HeaderContentType, httputil.TypeApplicationJSON)
	w.WriteHeader(status)

	if v == nil {
		return nil
	}

	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}
