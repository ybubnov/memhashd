package store

import (
	"testing"

	"memhashd/container/hash"
)

func TestStoreServe(t *testing.T) {
	s := newStore(&Options{Capacity: 0})
	s.Store("1", hash.Record{Data: []int{1, 2, 3}})

	val, err := s.Serve(&RequestListItem{Key: "1", Pos: 2})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if val.(int) != 3 {
		t.Fatalf("invalid value returned: %v", val)
	}
}
