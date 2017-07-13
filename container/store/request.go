package store

import (
	"reflect"

	"memhashd/container/hash"
)

type Requester interface {
	Request(hash.Hash) (Record, error)
}

type Response interface {
	//WriteStatus(Status)
	WriteBody(interface{})
}

// RequestKeys defines a request to a storage to retrieve a list of
// all stored keys.
type RequestKeys struct {
	// ID is a request identifier.
	ID string
}

// Request implements Requester interface, it returns a list of keys.
func (r *RequestKeys) Request(h hash.Hash) (interface{}, error) {
	return h.Keys(), nil
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
func (r *RequestStore) Request(h hash.Hash) (interface{}, error) {
	h.Store(r.Key, Record{Data: r.Value})
	return nil, nil
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
func (r *RequestLoad) Request(h hash.Hash) (interface{}, error) {
	rec, ok := h.Load(r.Key)
	if !ok {
		// error!
	}

	return rec.Data, nil
}

// RequestMeta defines a request to a storage to retrieve metadata
// (last access time, creation time, etc.) about the record by the
// given key.
type RequestMeta struct {
	// ID is a request identifier.
	ID string
	// Key is a name of the key.
	Key string
}

// Request implements Requester interface, it returns a record metadata
// stored in hash map.
func (r *RequestMeta) Request(h hash.Hash) (interface{}, error) {
	rec, ok := h.Load(r.Key)
	if !ok {
		// error!
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
func (r *RequestDelete) Request(h hash.Hash) (interface{}, error) {
	h.Delete(r.Key)
	return nil, nil
}

// RequestListItem defines a request to a store to retrieve an item from
// the list. When a given value is not a list or position exceeds an
// amount of items in a list, an error is returned.
type RequestListItem struct {
	// ID is a request identifier.
	ID string
	// Key is a name of the key.
	Key string
	// Pos is a position in a list.
	Pos int
}

// Request implements Requester interface, it returns an element of the
// list.
func (r *RequestListItem) Request(h hash.Hash) (interface{}, error) {
	rec, ok := h.Load(r.Key)
	if !ok {
		// error!
	}

	switch reflect.TypeOf(rec.Data).Kind() {
	case reflect.Slice:
		slice := reflect.ValueOf(rec.Data)
		if slice.Len() <= r.Pos {
			// error!
		}

		// Return an item at the requested position.
		val := slice.Index(r.Pos)
		if !val.IsValid() {
			// error!
		}

		return val.Interface(), error
	default:
		// error!
	}
}

// RequestDictItem defines a request to a store to retrieve an item
// from the dictionary. When a given value is not a dictionary type
// or requested item is not in a dictionary, an error is returned.
type RequestDictItem struct {
	// ID is a request identifier.
	ID     string
	Key    string
	SubKey interface{}
}

// Request implements Requester interface, it returns an item from
// the dictionary.
func (r *RequestDictItem) Request(h hash.Hash) (Record, error) {
	rec, ok := h.Load(r.Key)
	if !ok {
		// error!
	}

	switch reflect.TypeOf(rec.Data).Kind() {
	case reflect.Map:
		hashmap := reflect.ValueOf(rec.Data)
		key := reflect.ValueOf(r.SubKey)

		val := hashmap.MapIndex(val)
		if !val.IsValid() {
			// error!
		}

		return val.Interface(), error
	default:
		// error!
	}
}
