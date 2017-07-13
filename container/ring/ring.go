package ring

import (
	"hash"
	"hash/fnv"
)

type Ring interface {
	//
}

type Element struct {
	Value interface{}
}

type ring struct {
	ha hash.Hash32

	ratio   int
	virtual []int
}

func NewRing(ratio int) Ring {
	if ratio <= 0 {
		panic("ring: Ring of less than zero ratio")
	}

	return &ring{
		ha:    fnv.New32(),
		ratio: ratio,
		virt:  make([]int, ratio),
	}
}

func (r *ring) Insert() {
}

func (r *ring) Remove() {
}

func (r *ring) Find() {
}
