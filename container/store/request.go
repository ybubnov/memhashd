package store

import (
	"fmt"
	"reflect"
	"time"

	"memhashd/container/hash"
)

type Request interface {
	fmt.Stringer

	Hash() string

	Process(hash.Hash) (hash.Record, error)
}

// RequestKeys defines a request to a storage to retrieve a list of
// all stored keys.
type RequestKeys struct {
	// ID is a request identifier.
	ID string
}

// Hash implements Request interface. Hash for keys request is always
// an empty string, which means this request can be processed by a
// local shard.
func (r *RequestKeys) Hash() string {
	return ""
}

// String implements fmt.Stringer interface.
func (r *RequestKeys) String() string {
	return fmt.Sprintf("id: %s, type: keys", r.ID)
}

// Process implements Request interface, it returns a list of keys.
func (r *RequestKeys) Process(h hash.Hash) (hash.Record, error) {
	// Create a fake record, which does not represent an actual
	// record in a storage.
	return hash.Record{Data: h.Keys()}, nil
}

// RequestStore defines a request to a storage to store a value by
// the given key. Result should be overridden with the new value despite
// of the type of existing record.
type RequestStore struct {
	// ID is a request identifier.
	ID string
	// ExpireTime defines a record expiration time.
	ExpireTime time.Duration
	// Key is a key used to store an element in a store.
	Key string
	// Data is a for the given key.
	Data interface{} `json:"data"`
}

// Hash implements Request interface.
func (r *RequestStore) Hash() string {
	return r.Key
}

// String implements fmt.Stringer interface.
func (r *RequestStore) String() string {
	return fmt.Sprintf("id: %s, type: store, key: %s"+
		", data: %v, expire_time: %s",
		r.ID, r.Key, r.Data, r.ExpireTime)
}

// Process implements Request interface, it stores a value into the
// given hash-map. Hash should not be concurrently changed during this
// operation.
func (r *RequestStore) Process(h hash.Hash) (hash.Record, error) {
	rec := h.Store(r.Key, hash.Record{
		Data: r.Data, Meta: hash.Meta{ExpireTime: r.ExpireTime},
	})
	return rec, nil
}

// RequestLoad defines a request to a storage to load an element from
// the storage. When the requested key is missing, an error is returned.
type RequestLoad struct {
	// ID is a request identifier.
	ID string
	// Key is a name of the key.
	Key string
}

// Hash implements Request interface.
func (r *RequestLoad) Hash() string {
	return r.Key
}

// String implements fmt.Stringer interface.
func (r *RequestLoad) String() string {
	return fmt.Sprintf("id: %s, type: load, key: %s", r.ID, r.Key)
}

// Process implements Request interface, it returns a record value
// stored in a hash map.
func (r *RequestLoad) Process(h hash.Hash) (hash.Record, error) {
	rec, ok := h.Load(r.Key)
	if !ok {
		text := fmt.Sprintf("%s does not exist", r.Key)
		return hash.RecordZero, &ErrMissing{text}
	}
	return rec, nil
}

// RequestDelete defines a request to a storage to delete a record from
// the hash map stored by a given key. An error is returned if the given
// key is not in a hash map.
type RequestDelete struct {
	// ID is a request identifier.
	ID string
	// Key is a name of the key.
	Key string
}

// Hash implements Request interface.
func (r *RequestDelete) Hash() string {
	return r.Key
}

// String implements fmt.Stringer interface.
func (r *RequestDelete) String() string {
	return fmt.Sprintf("id: %s, type: delete, key: %s", r.ID, r.Key)
}

// Process implements Request interface, it deletes a record from the
// store.
func (r *RequestDelete) Process(h hash.Hash) (hash.Record, error) {
	h.Delete(r.Key)
	return hash.RecordZero, nil
}

// RequestListIndex defines a request to a store to retrieve an item from
// the list. When a given value is not a list or position exceeds an
// amount of items in a list, an error is returned.
type RequestListIndex struct {
	// ID is a request identifier.
	ID string
	// Key is a name of the key.
	Key string
	// Index is a position in a list.
	Index uint64
}

// Hash implements Request interface.
func (r *RequestListIndex) Hash() string {
	return r.Key
}

// String implements fmt.Stringer interface.
func (r *RequestListIndex) String() string {
	return fmt.Sprintf("id: %s, type: list index, key: %s"+
		", index: %d", r.ID, r.Key, r.Index)
}

// Process implements Request interface, it returns an element of the
// list.
func (r *RequestListIndex) Process(h hash.Hash) (hash.Record, error) {
	rec, ok := h.Load(r.Key)
	if !ok {
		text := fmt.Sprintf("%s does not exist", r.Key)
		return hash.RecordZero, &ErrMissing{text}
	}

	switch reflect.TypeOf(rec.Data).Kind() {
	case reflect.Slice:
		slice := reflect.ValueOf(rec.Data)
		if uint64(slice.Len()) <= r.Index {
			text := fmt.Sprintf("position %d is out of range", r.Index)
			return hash.RecordZero, &ErrConflict{text}
		}

		// Return an item at the requested position.
		val := slice.Index(int(r.Index))
		if !val.IsValid() {
			text := fmt.Sprintf("unexpected value at position %d", r.Index)
			return hash.RecordZero, &ErrInternal{text}
		}

		rec.Data = val.Interface()
		return rec, nil
	default:
		text := fmt.Sprintf("%s is not a list", r.Key)
		return hash.RecordZero, &ErrConflict{text}
	}
}

// RequestDictItem defines a request to a store to retrieve an item
// from the dictionary. When a given value is not a dictionary type
// or requested item is not in a dictionary, an error is returned.
type RequestDictItem struct {
	// ID is a request identifier.
	ID string
	// Key is a name of a key.
	Key string
	// A key of the dictionary to request.
	Item interface{}
}

// Hash implements Request interface.
func (r *RequestDictItem) Hash() string {
	return r.Key
}

// String implement fmt.Stringer interface.
func (r *RequestDictItem) String() string {
	return fmt.Sprintf("id: %s, type: dict item, key: %s"+
		", item: %v", r.ID, r.Key, r.Item)
}

// mapIndex returns a value at the index in a given map, it returns
// an error if key type is different or key is not in a mapping.
func (r *RequestDictItem) mapIndex(m reflect.Value, key reflect.Value) (
	val reflect.Value, err error) {

	defer func() {
		t := recover()
		switch t := t.(type) {
		case error:
			err = t
		case string:
			err = fmt.Errorf(t)
		}
	}()
	return m.MapIndex(key), nil
}

// Process implements Request interface, it returns an item from the
// dictionary.
func (r *RequestDictItem) Process(h hash.Hash) (hash.Record, error) {
	rec, ok := h.Load(r.Key)
	if !ok {
		text := fmt.Sprintf("%s does not exist", r.Key)
		return hash.RecordZero, &ErrMissing{text}
	}

	switch reflect.TypeOf(rec.Data).Kind() {
	case reflect.Map:
		hashmap := reflect.ValueOf(rec.Data)
		key := reflect.ValueOf(r.Item)

		val, err := r.mapIndex(hashmap, key)
		if err != nil {
			text := fmt.Sprintf("item %v is invalid", r.Item)
			return hash.RecordZero, &ErrConflict{text}
		}

		if !val.IsValid() {
			text := fmt.Sprintf("unexpected value at key %v", r.Item)
			return hash.RecordZero, &ErrInternal{text}
		}

		rec.Data = val.Interface()
		return rec, nil
	default:
		text := fmt.Sprintf("%s is not a dictionary", r.Key)
		return hash.RecordZero, &ErrConflict{text}
	}
}
