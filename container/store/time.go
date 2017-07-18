package store

import (
	"time"
)

// timeHeapElement an element of the time-ordered heap.
type timeHeapElement struct {
	// Time is a key of the element, it will be used for sorting
	// elements in a heap in increasing order.
	Time time.Time

	// Data is a placeholder for arbitrary data.
	Data interface{}
}

// timeHeap is a heap where time type is used for ordering of the
// elements.
type timeHeap struct {
	arr []*timeHeapElement
}

// newTimeHeap creates a new container. Capacity defines an initial
// size of the heap container.
func newTimeHeap(cap int) *timeHeap {
	return &timeHeap{make([]*timeHeapElement, 0, cap)}
}

// Len implements sort.Interface, it returns a length of the heap.
func (h *timeHeap) Len() int {
	return len(h.arr)
}

// Less implements sort.Interface, it return true, when element at
// position i is less than at position j and false otherwise.
func (h *timeHeap) Less(i, j int) bool {
	return h.arr[i].Time.Before(h.arr[j].Time)
}

// Swap implements sort.Interface, it swaps elements at i-th and j-th
// positions.
func (h *timeHeap) Swap(i, j int) {
	h.arr[i], h.arr[j] = h.arr[j], h.arr[i]
}

// Peek returns an element on the top of the heap and nil, when the
// heap is empty.
func (h *timeHeap) Peek() interface{} {
	if h.Len() != 0 {
		return h.arr[len(h.arr)-1]
	}
	return nil
}

// Push inserts a new element into a time-ordered heap.
func (h *timeHeap) Push(v interface{}) {
	h.arr = append(h.arr, v.(*timeHeapElement))
}

// Pop extracts an elements from the heap.
func (h *timeHeap) Pop() interface{} {
	n := len(h.arr)
	val := h.arr[n-1]
	h.arr = h.arr[:n-1]
	return val
}

// refreshTimer is timer that can be re-started.
type refreshTimer struct {
	// An internal timer.
	timer *time.Timer
	// cutoff defines a point in time when the timer should be stopped.
	cutoff time.Time
}

// AfterFunc schedules a run of a given function at the specified time.
//
// When timer is already started and the new time is closer, the timer
// will be stopped and a new function will be applied instead.
func (rt *refreshTimer) AfterFunc(t time.Time, fn func()) {
	// When the planned cutoff is later than a given time, timer has
	// to be stopped and re-scheduled again.
	now := time.Now()
	if !rt.cutoff.IsZero() && rt.cutoff.Before(t) && now.Before(rt.cutoff) {
		// There is no need to perform a re-scheduling.
		return
	}

	// Stop a timer. It does not matter whether it was stopped or not.
	// If stopped, it should be started again, if running - we have
	// a closer time point, thus can re-start it.
	if rt.timer != nil {
		rt.timer.Stop()
	}

	rt.cutoff = t
	rt.timer = time.AfterFunc(t.Sub(now), fn)
}
