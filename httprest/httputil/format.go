package httputil

import (
	"errors"
	"net/http"

	"memhashd/system/log"
)

var (
	// ErrNotSupported returnes when value provided in a Content-Type header is
	// not supported by any formatter.
	ErrNotSuppoted = errors.New("format: requested format not supported")

	formatters = make(map[string]ReadWriteFormatter)
)

// Marshaler is the interface implemented by an object
// that can marshal itself into a some form.
type WriteFormatter interface {
	// Write writes data in a specific format to response writer.
	Write(http.ResponseWriter, interface{}, int) error
}

// Unmarshaler is the interface implemented by an object
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

// FormatNameList returns list of registered formatters names.
func FormatNameList() (names []string) {
	for name := range formatters {
		names = append(names, name)
	}
	return
}

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
		return &JSONFormatter{}, ErrNotSuppoted
	}
	return formatter, nil
}

func ReadFormat(r *http.Request) (ReadFormatter, error) {
	f, err := formatOf(r.Header.Get(HeaderContentType))
	if err != nil {
		log.ErrorLogf("format/READ_FORMAT",
			"Failed to select read formatter for request: %s", err)
	}
	return f, err
}

func WriteFormat(r *http.Request) (WriteFormatter, error) {
	f, err := formatOf(r.Header.Get(HeaderAccept))
	if err != nil {
		log.ErrorLogf("format/WRITE_FORMAT",
			"Failed to select write formatter for request: %s", err)
	}
	return f, err
}
