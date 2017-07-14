package format

import (
	"errors"
	"net/http"

	"memhashd/system/log"
)

var (
	// ErrNotSupported returnes when value provided in a Content-Type header is
	// not supported by any formatter.
	ErrNotSuppoted = errors.New("format: requested format not supported")
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

// Format returns formatter for provided mime type,
// defaults to JSON formatter.
func Format(t string) (ReadWriteFormatter, error) {
	formatter, ok := formatters[t]
	if !ok {
		return &JSONFormatter{}, ErrNotSuppoted
	}

	return formatter, nil
}

var formatters = make(map[string]ReadWriteFormatter)

// Register registers formatter for specified media type.
func Register(t string, f ReadWriteFormatter) {
	if f == nil {
		log.FatalLog("format/REGISTER",
			"failed to register nil formatter for: ", t)
	}

	if _, dup := formatters[t]; dup {
		log.FatalLog("format/REGISTER",
			"failed to register duplicate formatter for: ", t)
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
