package store

import (
	"container/heap"
	"sync"
	"testing"
	"time"
)

func TestTimeHeap(t *testing.T) {
	h := newTimeHeap(10)
	if cap(h.arr) != 10 {
		t.Fatalf("invalid capacity of the heap")
	}

	now := time.Now()
	heap.Push(h, &timeHeapElement{now, 1})
	heap.Push(h, &timeHeapElement{now.Add(10 * time.Hour), 2})
	heap.Push(h, &timeHeapElement{now.Add(10 * time.Second), 3})

	tests := []struct {
		Data int
	}{
		{1}, {3}, {2},
	}

	for _, test := range tests {
		e := heap.Pop(h).(*timeHeapElement)
		if e.Data.(int) != test.Data {
			t.Fatalf("invalid data returned: %v", e.Data)
		}
	}
}

func TestRefreshTimer(t *testing.T) {
	rt := new(refreshTimer)
	now := time.Now()

	// Hopefully, an execution of this test take less
	// than 12 hours.
	rt.AfterFunc(now.Add(12*time.Hour), func() {
		t.Fatalf("should be re-scheduled")
	})

	var wg sync.WaitGroup
	wg.Add(1)

	rt.AfterFunc(now, func() { wg.Done() })
	wg.Wait()
}
