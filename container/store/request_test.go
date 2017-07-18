package store

import (
	"reflect"
	"testing"

	"memhashd/container/hash"
)

func TestRequestKeys(t *testing.T) {
	s := newStore(&Config{16})
	s.Store("1", hash.Record{Data: 1})
	s.Store("2", hash.Record{Data: 2})
	s.Store("3", hash.Record{Data: 3})
	s.Store("4", hash.Record{Data: 4})

	req := &RequestKeys{}
	if req.Action() != ActionKeys {
		t.Fatalf("invalid request action")
	}
	if req.Hash() != "" {
		t.Fatalf("invalid hash returned: %s", req.Hash())
	}
	rec, err := req.Process(s)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	keys := rec.Data.([]string)
	if !reflect.DeepEqual(keys, []string{"1", "2", "3", "4"}) {
		t.Fatalf("invalid keys are returned: %v", keys)
	}
}

func TestRequestStore(t *testing.T) {
	s := newStore(&Config{16})
	req := &RequestStore{Key: "1", Data: 1}
	if req.Action() != ActionStore {
		t.Fatalf("invalid request action")
	}
	if req.Hash() != "1" {
		t.Fatalf("invalid hash returned: %s", req.Hash())
	}
	if _, err := req.Process(s); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if rec, _ := s.Load("1"); rec.Data.(int) != 1 {
		t.Fatalf("invalid data stored: %v", rec.Data)
	}
}

func TestRequestLoad(t *testing.T) {
	s := newStore(&Config{16})
	s.Store("2", hash.Record{Data: 2})

	req := &RequestLoad{Key: "2"}
	if req.Action() != ActionLoad {
		t.Fatalf("invalid request action")
	}
	if req.Hash() != "2" {
		t.Fatalf("invalid hash returned: %s", req.Hash())
	}
	rec, err := req.Process(s)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if rec.Data.(int) != 2 {
		t.Fatalf("invalid data returned: %v", rec.Data)
	}
	s.Delete("2")
	_, err = req.Process(s)
	if err == nil || err.Error() != "2 does not exist" {
		t.Fatalf("expected an error, %v", err)
	}
}

func TestRequestDelete(t *testing.T) {
}

func TestRequestListIndex(t *testing.T) {
	s := newStore(&Config{0})
	req := &RequestListIndex{Key: "1", Index: 2}
	_, err := req.Process(s)
	if err == nil || err.Error() != "1 does not exist" {
		t.Fatalf("expected an error, %v", err)
	}

	s.Store("1", hash.Record{Data: []int{1, 2, 3}})
	rec, err := req.Process(s)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if rec.Data.(int) != 3 {
		t.Fatalf("invalid value returned: %v", rec.Data)
	}

	req = &RequestListIndex{Key: "1", Index: 4}
	rec, err = req.Process(s)
	if err == nil || err.Error() != "position 4 is out of range" {
		t.Fatalf("expected an error, %v", err)
	}

	s.Store("1", hash.Record{Data: 3})
	_, err = req.Process(s)
	if err == nil || err.Error() != "1 is not a list" {
		t.Fatalf("expected an error, %v", err)
	}
}

func TestRequestDictItem(t *testing.T) {
	s := newStore(&Config{0})
	req := &RequestDictItem{Key: "2", Item: 3}
	_, err := req.Process(s)
	if err == nil || err.Error() != "2 does not exist" {
		t.Fatalf("expected an error, %v", err)
	}

	s.Store("2", hash.Record{Data: map[int]int{3: 4}})
	rec, err := req.Process(s)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if rec.Data.(int) != 4 {
		t.Fatalf("invalid value returned: %v", rec.Data)
	}

	req = &RequestDictItem{Key: "2", Item: "a"}
	_, err = req.Process(s)
	if err == nil || err.Error() != "item a is invalid" {
		t.Fatalf("expected an error, %v", err)
	}

	req = &RequestDictItem{Key: "2", Item: 5}
	_, err = req.Process(s)
	if err == nil || err.Error() != "unexpected value at key 5" {
		t.Fatalf("expected an error, %v", err)
	}
	s.Store("2", hash.Record{Data: 0})
	_, err = req.Process(s)
	if err == nil || err.Error() != "2 is not a dictionary" {
		t.Fatalf("expected an error, %v", err)
	}
}
