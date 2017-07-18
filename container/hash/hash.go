package hash

import (
	"time"
)

var (
	// RecordZero is an empty instance of the hash record.
	RecordZero Record
)

// Meta is a metadata about the record in a hash.
type Meta struct {
	// Index defines a record serial number. Each time the record
	// data is updated an index value is incremented.
	Index int64

	// Time to live for the record. If time to live is less than or
	// equal to zero, record won't be ever evicted from the storage.
	ExpireTime time.Duration

	// AccessedAt defines a moment when the record was accessed last
	// time.
	AccessedAt time.Time

	// CreatedAt defines a moment when the record was stored into a
	// hash table.
	CreatedAt time.Time

	// UpdatedAt defines a moment when the record was updated. This
	// value is changed on update of the resource record.
	UpdatedAt time.Time
}

// Record defines a record of a hash table.
type Record struct {
	// Meta defines a record metadata.
	Meta Meta

	// A value of the record (can be a primitive type or list, map).
	Data interface{}
}

// IsExpired returns true when the ExpireTime is more than zero and record
// exist longer than value define in a ExpireTime.
func (r *Record) IsExpired() bool {
	if r.IsPermanent() {
		return false
	}

	now := time.Now()
	// Calculate how much data is presented in a storage.
	live := now.Sub(r.Meta.CreatedAt)
	return live > r.Meta.ExpireTime
}

// IsPermanent returns true, when expiration time is less than or equal
// to zero, and false otherwise.
func (r *Record) IsPermanent() bool {
	return r.Meta.ExpireTime <= 0
}

// ExpiresAt returns a moment in a future, when the record expires.
func (r *Record) ExpiresAt() time.Time {
	return r.Meta.CreatedAt.Add(r.Meta.ExpireTime)
}

// Hash describes types that implement hashing table.
type Hash interface {
	// Keys returns a list of keys.
	Keys() []string

	// Load returns a record persisted under the given key.
	Load(key string) (Record, bool)

	// Store persists the record under the given key.
	Store(key string, rec Record) Record

	// Delete removes the record stored under the given key.
	Delete(key string)
}

// unsafeHash is an extension of the builtin map type. It is unsafe to
// use this type by multiple go-routines.
type unsafeHash struct {
	records map[string]Record

	// The list of keys to be able retrieve them in a constant time
	// in average.
	keys []string

	// Dirty flags instructs that an amount of keys has been changed
	// since last construction, therefore a list of keys require rebuild.
	//
	// List becomes dirty, when a key is delete from the hash table.
	dirty bool
}

// NewUnsafeHash creates a new instance of the hash. This hash should
// not be used by multiple go-routines simultaneously.
func NewUnsafeHash(cap int) Hash {
	return newUnsafeHash(cap)
}

// newUnsafeHash creates a thread-safe hash table, where cap defines
// an initial capacity of the hash-table.
func newUnsafeHash(cap int) *unsafeHash {
	return &unsafeHash{
		records: make(map[string]Record, cap),
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
//
// When the record is expired it will be removed from the storage, until
// that moment it will occupy the memory.
func (h *unsafeHash) Load(key string) (rec Record, ok bool) {
	rec, ok = h.records[key]
	if !ok {
		return rec, false
	}

	// Update accessed time of the record in a hash table.
	rec.Meta.AccessedAt = time.Now()
	return rec, true
}

// Store implements Hash interface.
func (h *unsafeHash) Store(key string, rec Record) Record {
	prevrec, ok := h.records[key]
	if !ok {
		// Create a new record, when it is missing in the hash table.
		prevrec = Record{Meta: Meta{CreatedAt: time.Now()}}
	}

	// Append a new key into the list of keys only when it is not
	// dirty, otherwise, it will re-constructed during access of keys.
	if !h.dirty {
		h.keys = append(h.keys, key)
	}

	prevrec.Meta.Index++
	prevrec.Meta.UpdatedAt = time.Now()
	prevrec.Meta.ExpireTime = rec.Meta.ExpireTime
	prevrec.Data = rec.Data

	h.records[key] = prevrec
	return prevrec
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
