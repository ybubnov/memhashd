package ring

import (
	"hash/fnv"
)

type Ring interface {
	Insert(*Element)
	Remove(*Element)
	Find(Hasher) *Element
}

type Element struct {
	Value interface{}
}

type Hasher interface {
	Hash() uint32
}

func fnvSum32(b []byte) uint32 {
	ha := fnv.New32()
	ha.Write(b)
	return ha.Sum32()
}

type StringHasher string

func (h StringHasher) Hash() uint32 {
	return fnvSum32([]byte(h))
}

type ring struct {
	ratio    uint32
	virtual  []int
	elements []*Element
}

func newRing(ratio int) *ring {
	if ratio <= 0 {
		panic("ring: Ring ration is less than zero")
	}
	return &ring{
		ratio:   uint32(ratio),
		virtual: make([]int, ratio),
	}
}

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

func (r *ring) Insert(e *Element) {
	r.elements = append(r.elements, e)
	r.repartition()
}

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

func (r *ring) Find(h Hasher) *Element {
	index := h.Hash() % r.ratio
	element := r.virtual[index]
	return r.elements[element]
}
