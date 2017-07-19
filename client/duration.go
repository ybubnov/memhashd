package client

import (
	"bytes"
	"fmt"
	"time"
)

// Duration represents elapsed time between two instants. This type is
// used for encoding and decoding the duration into/from the string.
type Duration time.Duration

// UnmarshalJSON implements json.Unmarshaler interface.
func (d *Duration) UnmarshalJSON(b []byte) error {
	s := string(bytes.Trim(b, `"`))
	duration, err := time.ParseDuration(s)
	*d = Duration(duration)
	return err
}

// MarshalJSON implements json.Marshaler interface.
func (d Duration) MarshalJSON() ([]byte, error) {
	duration := time.Duration(d)
	return []byte(fmt.Sprintf(`"%s"`, duration)), nil
}
