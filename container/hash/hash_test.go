package hash

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

var (
	benchmarkKeys   []string
	benchmarkRecord *Record
)

func TestUnsafeHashStore(t *testing.T) {
	h := newUnsafeHash(0)
	uh := h.(*unsafeHash)

	tests := []struct {
		Key    string
		Before interface{}
		After  interface{}
	}{
		{"1", 1, 2},
		{"2", "hello", "goodbye"},
		{"3", []int{1, 2, 3}, []int{3, 4}},
		{"4", map[string]int{"a": 2, "b": 3}, map[string]int{}},
	}

	for _, tt := range tests {
		h.Store(tt.Key, &Record{Data: tt.Before})
		rec, ok := uh.records[tt.Key]

		if !ok {
			t.Fatalf("key `%s` is expected to be in hash table", tt.Key)
		}

		if !reflect.DeepEqual(tt.Before, rec.Data) {
			t.Fatalf("value for `%s` must be `%v`, got `%v`",
				tt.Key, tt.Before, rec.Data)
		}

		if rec.CreatedAt.IsZero() {
			t.Fatalf("created at time for `%s` must be non-zero", tt.Key)
		}

		if rec.Index != 1 {
			t.Fatalf("index for `%s` must be equal to 1", tt.Key)
		}

		h.Store(tt.Key, &Record{Data: tt.After})
		rec, _ = uh.records[tt.Key]

		if !reflect.DeepEqual(tt.After, rec.Data) {
			t.Fatalf("value for `%s` must be `%v`, got `%v`",
				tt.Key, tt.After, rec.Data)
		}

		if rec.UpdatedAt.IsZero() {
			t.Fatalf("updated at time for `%s` must be non-zero", tt.Key)
		}
	}
}

func TestUnsafeHashLoad(t *testing.T) {
	h := newUnsafeHash(0)

	tests := []struct {
		Key  string
		Data interface{}
	}{
		{"a", 1},
		{"b", 2},
	}

	for _, tt := range tests {
		h.Store(tt.Key, &Record{Data: tt.Data})
		rec, ok := h.Load(tt.Key)

		if !ok {
			t.Fatalf("key `%s` expected to be in hash", tt.Key)
		}

		if rec.AccessedAt.IsZero() {
			t.Fatalf("accessed at time for `%s` must be non-zero", tt.Key)
		}
	}
}

func TestUnsafeHashKeys(t *testing.T) {
	h := newUnsafeHash(0)
	h.Store("1", &Record{Data: 42})
	h.Store("2", &Record{Data: 43})

	tests := []struct {
		Key  string
		Data interface{}
		Keys []string
	}{
		{"3", 44, []string{"1", "2", "3"}},
		{"4", 45, []string{"1", "2", "3", "4"}},
	}

	for _, tt := range tests {
		h.Store(tt.Key, &Record{Data: tt.Data})
		if keys := h.Keys(); !reflect.DeepEqual(keys, tt.Keys) {
			t.Fatalf("invalid list of keys returned: `%v`", keys)
		}
	}
}

func TestUnsafeHashDelete(t *testing.T) {
	h := newUnsafeHash(0)

	h.Store("1", &Record{Data: 0})
	h.Store("2", &Record{Data: 1})
	h.Store("3", &Record{Data: 2})
	h.Store("4", &Record{Data: 3})

	tests := []struct {
		Key  string
		Keys []string
	}{
		{"2", []string{"1", "3", "4"}},
		{"2", []string{"1", "3", "4"}},
		{"4", []string{"1", "3"}},
		{"1", []string{"3"}},
		{"3", []string{}},
		{"3", []string{}},
	}

	for _, tt := range tests {
		h.Delete(tt.Key)
		keys := sort.StringSlice(h.Keys())
		sort.Sort(keys)

		if !reflect.DeepEqual(h.Keys(), tt.Keys) {
			t.Fatalf("invalid list of keys returned: `%v`", keys)
		}
	}
}

func BenchmarkUnsafeHashKeys(b *testing.B) {
	h := newUnsafeHash(0)
	var keys []string

	for i := 0; i < b.N; i++ {
		h.Store(fmt.Sprintf("%d", i), &Record{Data: nil})
		keys = h.Keys()
	}

	// Force the compiler to preserve the assignment instructions
	// from the previous loop to calculate the actual performance
	// of the keys retrieval.
	benchmarkKeys = keys
}

func BenchmarkUnsafeHashLoad(b *testing.B) {
	h := newUnsafeHash(0)
	var rec *Record

	for i := 0; i < b.N; i++ {
		h.Store(fmt.Sprintf("%d", i), &Record{Data: nil})
		rec, _ = h.Load(fmt.Sprintf("%d", i))
	}

	benchmarkRecord = rec
}

func BenchmarkUnsafeHashStore(b *testing.B) {
	h := newUnsafeHash(0)
	for i := 0; i < b.N; i++ {
		h.Store(fmt.Sprintf("%d", i), &Record{Data: nil})
	}
}

func BenchmarkUnsafeHashDelete(b *testing.B) {
	h := newUnsafeHash(0)
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("%d", i)
		h.Store(key, &Record{Data: nil})
		h.Delete(key)
	}
}
