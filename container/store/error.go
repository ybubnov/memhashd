package store

// ErrInternal describes internal errors of the storage.
type ErrInternal struct {
	// Text is a text of the error.
	Text string
}

// Error implements error interface. It returns a string representation
// of the error.
func (e *ErrInternal) Error() string {
	return e.Text
}

// ErrConflict describes type conflict errors. Error if this type is
// returned when the request cannot be processed because type of the
// target value does not match to the expected one.
type ErrConflict struct {
	// Test is a text of the error.
	Text string
}

// Error implements error interface. It returns a string representation
// of the error.
func (e *ErrConflict) Error() string {
	return e.Text
}

// ErrMissing describes error generated when the requested key is
// missing in a store.
type ErrMissing struct {
	// Test is a text of the error.
	Text string
}

// Error implements error interface. It returns a string representation
// of the error.
func (e *ErrMissing) Error() string {
	return e.Text
}
