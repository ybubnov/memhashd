package hash

import (
	"time"
)

// Record defines a record of a hash table.
type Record struct {
	// Index defines a record serial number. Each time the record
	// data is updated an index value is incremented.
	Index int64

	// Time to live for the record. If time to live is less than zero
	// means, record won't be ever evicted from the storage.
	TTL time.Time

	// AccessedAt defines a moment when the record was accessed last
	// time.
	AccessedAt time.Time

	// CreatedAt defines a moment when the record was stored into a
	// hash table.
	CreatedAt time.Time

	// UpdatedAt defines a moment when the record was updated. This
	// value is changed on update of the resource record.
	UpdatedAt time.Time

	// A value of the record (can be a primitive type or list, map).
	Data interface{}
}

type Hash interface {
	Keys() []string
	Load(key string) (rec *Record, ok bool)
	Store(key string, rec *Record)
	Delete(key string)
}

// unsafeHash is an extension of the builtin map type. It is unsafe to
// use this type by multiple go-routines.
type unsafeHash struct {
	records map[string]*Record

	// The list of keys to be able retrieve them in a constant time
	// in average.
	keys []string

	// Dirty flags instructs that an amount of keys has been changed
	// since last construction, therefore a list of keys require rebuild.
	//
	// List becomes dirty, when a key is delete from the hash table.
	dirty bool
}

// newUnsafeHash creates a thread-safe hash table, where cap defines
// an initial capacity of the hash-table.
func newUnsafeHash(cap int) Hash {
	return &unsafeHash{
		records: make(map[string]*Record, cap),
	}
}

// Keys implements Hash interface.
func (h *unsafeHash) Keys() []string {
	if !h.dirty {
		return h.keys
	}

	// Reduce the length of the slice to zero, but preserve capacity,
	// so it won't spend time on allocation of a new slice.
	h.keys = h.keys[:0]
	for key := range h.records {
		h.keys = append(h.keys, key)
	}

	h.dirty = false
	return h.keys
}

// Load implements Hash interface.
func (h *unsafeHash) Load(key string) (*Record, bool) {
	rec, ok := h.records[key]
	if !ok {
		return nil, false
	}

	// Update accessed time of the record in a hash table.
	rec.AccessedAt = time.Now()
	return rec, true
}

// Store implements Hash interface.
func (h *unsafeHash) Store(key string, rec *Record) {
	if rec == nil {
		panic("hash: Store record is nil")
	}

	prevrec, ok := h.records[key]
	if !ok {
		// Create a new record, when it is missing in the hash table.
		prevrec = &Record{CreatedAt: time.Now()}
		h.records[key] = prevrec
	}

	// Append a new key into the list of keys only when it is not
	// dirty, otherwise, it will re-constructed during access of keys.
	if !h.dirty {
		h.keys = append(h.keys, key)
	}

	prevrec.Index++
	prevrec.UpdatedAt = time.Now()
	prevrec.Data = rec.Data
	prevrec.TTL = rec.TTL
}

// Delete implements Hash interface.
func (h *unsafeHash) Delete(key string) {
	// Mark the keys as dirty only when the key to delete is presented
	// in the hash, otherwise there is no need to re-build keys list.
	if _, ok := h.records[key]; ok {
		h.dirty = true
	}
	delete(h.records, key)
}
