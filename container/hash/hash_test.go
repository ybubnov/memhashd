package hash

import (
	"reflect"
	"testing"
)

func TestUnsafeHashStore(t *testing.T) {
	h := newUnsafeHash(0)
	uh := h.(*unsafeHash)

	assert := func(key string, val1 interface{}, val2 interface{}) {
		h.Store(key, &Record{Data: val1})
		rec, ok := uh.records[key]

		if !ok {
			t.Fatalf("key `%s` is expected to be in hash table", key)
		}

		if !reflect.DeepEqual(val1, rec.Data) {
			t.Fatalf("value for `%s` must be `%v`, got `%v`", key, val1, rec.Data)
		}

		if rec.CreatedAt.IsZero() {
			t.Fatalf("created at time for `%s` must be non-zero", key)
		}

		if rec.Index != 1 {
			t.Fatalf("index for `%s` must be equal to 1", key)
		}

		h.Store(key, &Record{Data: val2})
		rec, _ = uh.records[key]

		if !reflect.DeepEqual(val2, rec.Data) {
			t.Fatalf("value for `%s` must be `%v`, got `%v`", key, val2, rec.Data)
		}

		if rec.UpdatedAt.IsZero() {
			t.Fatalf("updated at time for `%s` must be non-zero", key)
		}
	}

	assert("1", 1, 2)
	assert("2", "hello", "goodbye")
	assert("3", []int{1, 2, 3}, []int{3, 4})
	assert("4", map[string]int{"a": 2, "b": 3}, map[string]int{"4": 0})
}

func TestUnsafeHashLoad(t *testing.T) {
	h := newUnsafeHash(0)

	assert := func(key string, val interface{}) {
		h.Store(key, &Record{Data: val})
		rec, ok := h.Load(key)

		if !ok {
			t.Fatalf("key `%s` expected to be in hash", key)
		}

		if rec.AccessedAt.IsZero() {
			t.Fatalf("accessed at time for `%s` must be non-zero", key)
		}
	}

	assert("a", 1)
	assert("b", 2)
}

func TestUnsafeHashKeys(t *testing.T) {
}

func TestUnsafeHashDelete(t *testing.T) {
}

func BenchmarkUnsafeHashKeys(b *testing.B) {
	h := newUnsafeHash(0)
	for i := 0; i < b.N; i++ {
		_ = h.Keys()
	}
}

func BenchmarkUnsafeHashLoad(b *testing.B) {
	//h := newUnsafeHash(0)
	for i := 0; i < b.N; i++ {
	}
}

func BenchmarkUnsafeHashStore(b *testing.B) {
	//h := newUnsafeHash(0)
	for i := 0; i < b.N; i++ {
	}
}

func BenchmarkUnsafeHashDelete(b *testing.B) {
	//h := newUnsafeHash(0)
	for i := 0; i < b.N; i++ {
	}
}
