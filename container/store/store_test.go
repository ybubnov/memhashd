package store

import (
	"reflect"
	"testing"
	"time"

	"memhashd/container/hash"
)

func TestStoreStore(t *testing.T) {
	s := newStore(&Config{16})

	now := time.Now()
	expire := 60 * time.Minute

	meta := hash.Meta{ExpireTime: expire}
	s.Store("1", hash.Record{Data: 1, Meta: meta})

	if s.expireTimer.cutoff.Before(now.Add(expire)) {
		t.Fatalf("timer scheduled incorrectly")
	}

	rec, ok := s.hashMap.Load("1")
	if !ok || rec.Data.(int) != 1 {
		t.Fatalf("invalid record returned")
	}
}

func TestStoreKeys(t *testing.T) {
	s := newStore(&Config{16})
	s.Store("1", hash.Record{Data: 1})
	s.Store("2", hash.Record{Data: 2})
	s.Store("3", hash.Record{Data: 3})
	s.Store("4", hash.Record{Data: 4})

	keys := s.Keys()
	if !reflect.DeepEqual(keys, []string{"1", "2", "3", "4"}) {
		t.Fatalf("invalid set of keys returned: %v", keys)
	}
}

func TestStoreLoad(t *testing.T) {
	s := newStore(&Config{16})

	s.Store("2", hash.Record{Data: 2})
	s.Store("1", hash.Record{Data: 1, Meta: hash.Meta{
		ExpireTime: 1 * time.Nanosecond}})

	time.Sleep(100 * time.Millisecond)
	if _, ok := s.Load("1"); ok {
		t.Fatalf("record should not be in store")
	}
	if _, ok := s.Load("2"); !ok {
		t.Fatalf("record should be in store")
	}
}

func TestDeleteExpiredKeys(t *testing.T) {
	s := newStore(&Config{16})

	s.Store("1", hash.Record{Data: 1, Meta: hash.Meta{
		ExpireTime: 1 * time.Nanosecond}})
	s.Store("2", hash.Record{Data: 2, Meta: hash.Meta{
		ExpireTime: 10 * time.Nanosecond}})

	time.Sleep(100 * time.Millisecond)
	if _, ok := s.Load("1"); ok {
		t.Fatalf("record should not be in a store")
	}
	if _, ok := s.Load("2"); ok {
		t.Fatalf("record should not be in a store")
	}
}
