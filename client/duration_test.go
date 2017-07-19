package client

import (
	"testing"
	"time"
)

func TestDurationJSON(t *testing.T) {
	tests := []struct {
		Duration Duration
		Text     string
	}{
		{Duration(1 * time.Second), "1s"},
		{Duration(4 * time.Minute), "5m"},
		{Duration(10 * time.Hour), "10h"},
	}

	for _, tt := range tests {
		b, err := tt.Duration.MarshalJSON()
		if err != nil {
			t.Fatalf("unexpected error returned: %s", err)
		}

		var duration Duration
		err = duration.UnmarshalJSON(b)
		if err != nil {
			t.Fatalf("unexpected error returned: %s", err)
		}
	}
}
