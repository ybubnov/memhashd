package ring

import (
	"hash/fnv"
)

// Ring describes types that implement a ring container.
//
// The ring is an implementation of consistent hashing algorithm used
// for equal balancing of the data between multiple database shards.
type Ring interface {
	// Insert inserts a new element into a ring. After the insertion,
	// the partition assignments will be re-calculated.
	Insert(*Element)

	// Remove removes element from the ring. After the remove, the
	// partition assignments will be re-calculated.
	Remove(*Element)

	// Find searches for element, that is assigned to the given key.
	Find(Hasher) *Element
}

// Element defines an element of the ring.
type Element struct {
	Value interface{}
}

// Hasher describes hashable types. In other words, such types that can
// be used as keys of the hash-table.
type Hasher interface {
	// Hash returns a 32-bit checksum of the type.
	Hash() uint32
}

// fnvSum32 calculates Fowler-Noll-Vo hash of the given sequence of
// bytes.
func fnvSum32(b []byte) uint32 {
	ha := fnv.New32()
	ha.Write(b)
	return ha.Sum32()
}

// StringHasher is an implementation of the Hasher interface that
// calculates Fowler-Noll-Vo checksum of the string.
type StringHasher string

// Hash implement Hasher interface.
func (h StringHasher) Hash() uint32 {
	return fnvSum32([]byte(h))
}

// ring is a consistent-hashing ring with virtual partitions. The ring
// represents a space of the keys, which is divided to the equal parts.
//
// Each element can be assigned to multiple partitions of the ring.
type ring struct {
	ratio    uint32
	virtual  []int
	elements []*Element
}

// newRing creates a new instance of the ring with the given partition
// ratio. Ratio should be more than zero.
func newRing(ratio int) *ring {
	if ratio <= 0 {
		panic("ring: Ring ration is less than zero")
	}
	return &ring{
		ratio:   uint32(ratio),
		virtual: make([]int, ratio),
	}
}

// New creates a new instance of the consistent-hashing ring.
func New(ratio int) Ring {
	return newRing(ratio)
}

// repartition performs a remapping of virtual partitions to the
// elements. It is called after insertion or deletion of elements.
func (r *ring) repartition() {
	num := len(r.elements)
	for ii := range r.virtual {
		r.virtual[ii] = ii % num
	}
}

// Insert implements Ring interface.
func (r *ring) Insert(e *Element) {
	r.elements = append(r.elements, e)
	r.repartition()
}

// Remove implements consistent hashing interface.
func (r *ring) Remove(e *Element) {
	for ii := 0; ii < len(r.elements); ii++ {
		if r.elements[ii].Value != e.Value {
			continue
		}

		r.elements = append(r.elements[:ii], r.elements[ii+1:]...)
		break
	}
	r.repartition()
}

// Find implements Ring interface.
func (r *ring) Find(h Hasher) *Element {
	index := h.Hash() % r.ratio
	element := r.virtual[index]
	return r.elements[element]
}
