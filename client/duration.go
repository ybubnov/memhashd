package client

import (
	"bytes"
	"fmt"
	"time"
)

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	s := string(bytes.Trim(b, `"`))
	duration, err := time.ParseDuration(s)
	*d = Duration(duration)
	return err
}

func (d Duration) MarshalJSON() ([]byte, error) {
	duration := time.Duration(d)
	return []byte(fmt.Sprintf(`"%s"`, duration)), nil
}
