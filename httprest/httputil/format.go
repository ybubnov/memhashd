package httputil

import (
	"errors"
	"net/http"

	"github.com/ybubnov/memhashd/system/log"
)

var (
	// ErrNotSupported returnes when value provided in a Content-Type header is
	// not supported by any formatter.
	ErrNotSupported = errors.New("format: requested format not supported")

	// formatters stores registered formatters.
	formatters = make(map[string]ReadWriteFormatter)
)

// WriteFormatter is the interface implemented by an object
// that can marshal itself into a some form.
type WriteFormatter interface {
	// Write writes data in a specific format to response writer.
	Write(http.ResponseWriter, interface{}, int) error
}

// ReadFormatter is the interface implemented by an object
// that can unmarshal a representation of itself.
type ReadFormatter interface {
	// Read reads data in a specific format from request.
	Read(*http.Request, interface{}) error
}

// ReadWriteFormatter is the interface implemented by an object that
// can marshal and unmarshal objects of specific format.
type ReadWriteFormatter interface {
	ReadFormatter
	WriteFormatter
}

// Register registers formatter for specified media type.
func Register(t string, f ReadWriteFormatter) {
	if f == nil {
		log.FatalLogf("format/REGISTER",
			"failed to register nil formatter for: %s", t)
	}
	if _, dup := formatters[t]; dup {
		log.FatalLogf("format/REGISTER",
			"failed to register duplicate formatter for: %s", t)
	}
	formatters[t] = f
}

// Format returns read and write formatters according to the headers
// of the HTTP request. When either content-type or accept media type
// is not supported, method terminates the request with respective
// status code.
func Format(rw http.ResponseWriter, r *http.Request) (ReadFormatter, WriteFormatter, error) {
	rf, err := ReadFormat(r)
	if err != nil {
		rw.WriteHeader(http.StatusUnsupportedMediaType)
		return nil, nil, err
	}

	wf, err := WriteFormat(r)
	if err != nil {
		rw.WriteHeader(http.StatusNotAcceptable)
		return nil, nil, err
	}
	return rf, wf, nil
}

// formatOf returns formatter for provided mime type, defaults to JSON
// formatter.
func formatOf(t string) (ReadWriteFormatter, error) {
	formatter, ok := formatters[t]
	if !ok {
		return &JSONFormatter{}, ErrNotSupported
	}
	return formatter, nil
}

// ReadFormat returns a read-formatter according to the content type
// specified in headers of the HTTP request.
func ReadFormat(r *http.Request) (ReadFormatter, error) {
	f, err := formatOf(r.Header.Get(HeaderContentType))
	if err != nil {
		log.ErrorLogf("format/READ_FORMAT",
			"Failed to select read formatter for request: %s", err)
	}
	return f, err
}

// WriteFormat returns a write-formatter according to the accept header
// specified in the HTTP request.
func WriteFormat(r *http.Request) (WriteFormatter, error) {
	f, err := formatOf(r.Header.Get(HeaderAccept))
	if err != nil {
		log.ErrorLogf("format/WRITE_FORMAT",
			"Failed to select write formatter for request: %s", err)
	}
	return f, err
}
