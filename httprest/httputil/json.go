package httputil

import (
	"encoding/json"
	"net/http"
)

func init() {
	// Register formatter for application/json and */* types.
	Register(TypeApplicationJSON, &JSONFormatter{})
	Register(TypeAny, &JSONFormatter{})
	Register("", &JSONFormatter{})
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
	w.Header().Set(HeaderContentType, TypeApplicationJSON)
	w.WriteHeader(status)

	if v == nil {
		return nil
	}

	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}
