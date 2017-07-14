package store

type ErrInternal struct {
	Text string
}

func (e *ErrInternal) Error() string {
	return e.Text
}

type ErrConflict struct {
	Text string
}

func (e *ErrConflict) Error() string {
	return e.Text
}

type ErrMissing struct {
	Text string
}

func (e *ErrMissing) Error() string {
	return e.Text
}
