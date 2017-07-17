package store

import (
	"container/heap"
	"sync"
	"time"

	"memhashd/container/hash"
	"memhashd/system/log"
)

type Store interface {
	hash.Hash
	Serve(r Request) (hash.Record, error)
}

type Config struct {
	Capacity int
}

type store struct {
	hashMap hash.Hash

	expireHeap  *timeHeap
	expireTimer *refreshTimer

	worldMu sync.Mutex
}

func New(config *Config) Store {
	return newStore(config)
}

func newStore(config *Config) *store {
	return &store{
		hashMap:     hash.NewUnsafeHash(config.Capacity),
		expireHeap:  newTimeHeap(config.Capacity),
		expireTimer: new(refreshTimer),
	}
}

func (s *store) Keys() []string {
	s.worldMu.Lock()
	defer s.worldMu.Unlock()
	return s.hashMap.Keys()
}

func (s *store) Load(key string) (rec hash.Record, ok bool) {
	s.worldMu.Lock()
	defer s.worldMu.Unlock()

	rec, ok = s.hashMap.Load(key)
	if !ok {
		return rec, ok
	}

	// Remove an expired key to guarantee consistency of the storage.
	if rec.IsExpired() {
		log.DebugLogf("store/LOAD", "key %s is expired, deleting", key)
		s.hashMap.Delete(key)
		return hash.Record{}, false
	}

	return rec, ok
}

func (s *store) Store(key string, rec hash.Record) hash.Record {
	s.worldMu.Lock()
	defer s.worldMu.Unlock()

	// Store a new record into a storage.
	rec = s.hashMap.Store(key, rec)
	if rec.IsPermanent() {
		return rec
	}

	// For non-permanent records, calculate expiration time and schedule
	// an timer, that will purge all records with lower lifetime.
	cutoff := rec.ExpiresAt()
	heap.Push(s.expireHeap, &timeHeapElement{cutoff, key})

	log.DebugLogf("store/STORE",
		"scheduling next run of timer in %s", cutoff)
	s.expireTimer.AfterFunc(cutoff, func() { s.deleteAfter(cutoff) })
	return rec
}

func (s *store) deleteAfter(cutoff time.Time) {
	s.DeleteExpiredKeys(cutoff)
	s.worldMu.Lock()
	defer s.worldMu.Unlock()

	// When the length of the heap is zero, there are no more temporary
	// keys in it, therefore timer won't be started until a new record
	// will be added to a heap.
	if s.expireHeap.Len() == 0 {
		return
	}

	// Peek next timer form the heap and schedule an expiration timer.
	elem := s.expireHeap.Peek().(*timeHeapElement)
	log.DebugLogf("store/DELETE_AFTER",
		"re-scheduling next run of timer in %s", elem.Time)
	s.expireTimer.AfterFunc(elem.Time, func() {
		s.deleteAfter(elem.Time)
	})
}

// DeleteExpiredKeys removes expired keys from the storage and extracts
// all timers that are less than a specified cutoff.
func (s *store) DeleteExpiredKeys(cutoff time.Time) {
	s.worldMu.Lock()
	defer s.worldMu.Unlock()

	log.DebugLogf("store/DELETE_EXPIRED_KEYS",
		"starting deletion of expired keys")
	for {
		// Peek the next element and check if the saved record
		// is already expired, so it has to be removed.
		next, ok := s.expireHeap.Peek().(*timeHeapElement)
		if !ok || next == nil || next.Time.After(cutoff) {
			break
		}

		// Remove a keys from the storage and remove time from the heap of
		// expiration times.
		key := next.Data.(string)
		log.DebugLogf("store/DELETE_EXPIRED_KEYS",
			"deleted expired key `%s`", key)

		s.hashMap.Delete(key)
		s.expireHeap.Pop()
	}
	log.DebugLogf("store/DELETE_EXPIRED_KEYS",
		"stopped deletion of expired keys")
}

func (s *store) Delete(key string) {
	s.worldMu.Lock()
	defer s.worldMu.Unlock()
	s.hashMap.Delete(key)
}

func (s *store) Serve(r Request) (hash.Record, error) {
	return r.Process(s)
}
