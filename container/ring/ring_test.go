package ring

import (
	"reflect"
	"testing"
)

func TestRingInsert(t *testing.T) {
	r := newRing(8)
	r.Insert(&Element{Value: 1})
	r.Insert(&Element{Value: 2})

	if len(r.elements) != 2 {
		t.Fatalf("expected two elements in a ring")
	}

	mapping := []int{0, 1, 0, 1, 0, 1, 0, 1}
	if !reflect.DeepEqual(r.virtual, mapping) {
		t.Fatalf("invalid partition mapping: %v", r.virtual)
	}
}

func TestRingRemove(t *testing.T) {
	r := newRing(4)
	r.Insert(&Element{Value: 1})
	r.Insert(&Element{Value: 2})
	r.Insert(&Element{Value: 3})
	r.Insert(&Element{Value: 4})

	r.Remove(&Element{Value: 2})
	if len(r.elements) != 3 {
		t.Fatalf("expected to elements in a ring")
	}

	mapping := []int{0, 1, 2, 0}
	if !reflect.DeepEqual(r.virtual, mapping) {
		t.Fatalf("invalid partition mapping: %v", r.virtual)
	}
}

func TestRingFind(t *testing.T) {
	r := newRing(4)
	r.Insert(&Element{Value: 1})
	r.Insert(&Element{Value: 2})
	r.Insert(&Element{Value: 3})

	tests := []struct {
		Key   string
		Value int
	}{
		{"1", 3},
		{"3", 1},
		{"0", 1},
	}

	for _, tt := range tests {
		el := r.Find(StringHasher(tt.Key))
		if el.Value.(int) != tt.Value {
			t.Fatalf("invalid node returned for %s: %v", tt.Key, el.Value)
		}
	}
}
