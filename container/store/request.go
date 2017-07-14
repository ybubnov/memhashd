package store

import (
	"fmt"
	"reflect"

	"memhashd/container/hash"
)

type Request interface {
	// ID returns an identifier of the request for easier tracking
	// of the request lifetime.
	ID() string

	Process(hash.Hash) (hash.Record, error)
}

// RequestKeys defines a request to a storage to retrieve a list of
// all stored keys.
type RequestKeys struct {
	// ID is a request identifier.
	ID string
}

// Request implements Requester interface, it returns a list of keys.
func (r *RequestKeys) Process(h hash.Hash) (hash.Record, error) {
	panic("IS NOT A RECORD!!!")
	//return h.Keys(), nil
	return hash.RecordZero, nil
}

// RequestStore defines a request to a storage to store a value by
// the given key. Result should be overridden with the new value despite
// of the type of existing record.
type RequestStore struct {
	// ID is a request identifier.
	ID string
	// Key is a key used to store an element in a store.
	Key string
	// Value is a for the given key.
	Value interface{}
}

// Request implements Requester interface, it stores a value into the
// given hash-map. Hash should not be concurrently changed during this
// operation.
func (r *RequestStore) Process(h hash.Hash) (hash.Record, error) {
	h.Store(r.Key, hash.Record{Data: r.Value})
	return hash.RecordZero, nil
}

// RequestLoad defines a request to a storage to load an element from
// the storage. When the requested key is missing, an error is returned.
type RequestLoad struct {
	// ID is a request identifier.
	ID string
	// Key is a name of the key.
	Key string
}

// Request implements Requester interface, it returns a record value
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

// Request implements Requester interface, it deletes a record from the
// store.
func (r *RequestDelete) Process(h hash.Hash) (hash.Record, error) {
	h.Delete(r.Key)
	return hash.RecordZero, nil
}

// RequestListItem defines a request to a store to retrieve an item from
// the list. When a given value is not a list or position exceeds an
// amount of items in a list, an error is returned.
type RequestListItem struct {
	// ID is a request identifier.
	ID string
	// Key is a name of the key.
	Key string
	// Index is a position in a list.
	Index uint64
}

// Request implements Requester interface, it returns an element of the
// list.
func (r *RequestListItem) Process(h hash.Hash) (hash.Record, error) {
	rec, ok := h.Load(r.Key)
	if !ok {
		text := fmt.Sprintf("%s does not exist", r.Key)
		return hash.RecordZero, &ErrMissing{text}
	}

	switch reflect.TypeOf(rec.Data).Kind() {
	case reflect.Slice:
		slice := reflect.ValueOf(rec.Data)
		if slice.Len() <= int(r.Index) {
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

// Request implements Requester interface, it returns an item from
// the dictionary.
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

		val := hashmap.MapIndex(key)
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
