package store

import (
	"testing"
)

func TestErrorInternal(t *testing.T) {
	err := error(&ErrInternal{Text: "boom!"})
	if err.Error() != "boom!" {
		t.Fatalf("invalid error string returned")
	}
}

func TestErrorConflict(t *testing.T) {
	err := error(&ErrConflict{Text: "kaaboom!"})
	if err.Error() != "kaaboom!" {
		t.Fatalf("invalid error string returned")
	}
}

func TestErrorMissing(t *testing.T) {
	err := error(&ErrMissing{Text: "zappp"})
	if err.Error() != "zappp" {
		t.Fatalf("invalid error string returned")
	}
}
