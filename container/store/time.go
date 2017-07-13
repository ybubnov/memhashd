package store

import (
	"time"
)

type timeHeapElement struct {
	Time time.Time
	Data interface{}
}

type timeHeap struct {
	arr []*timeHeapElement
}

func newTimeHeap(cap int) *timeHeap {
	return &timeHeap{make([]*timeHeapElement, 0, cap)}
}

func (h *timeHeap) Len() int {
	return len(h.arr)
}

func (h *timeHeap) Less(i, j int) bool {
	return h.arr[i].Time.Before(h.arr[j].Time)
}

func (h *timeHeap) Swap(i, j int) {
	h.arr[i], h.arr[j] = h.arr[j], h.arr[i]
}

func (h *timeHeap) Peek() interface{} {
	if h.Len() != 0 {
		return h.arr[0]
	}
	return nil
}

func (h *timeHeap) Push(v interface{}) {
	h.arr = append(h.arr, v.(*timeHeapElement))
}

func (h *timeHeap) Pop() interface{} {
	n := len(h.arr)
	val := h.arr[n-1]
	h.arr = h.arr[:n-1]
	return val
}

type refreshTimer struct {
	timer  *time.Timer
	cutoff time.Time
}

func (rt *refreshTimer) AfterFunc(t time.Time, fn func()) {
	// When the planned cutoff is later than a given time, timer has
	// to be stopped and re-scheduled again.
	if !rt.cutoff.IsZero() && rt.cutoff.Before(t) {
		// There is no need to perform a re-scheduling.
		return
	}

	// Stop a timer. It does not matter whether it was stopped or not.
	// If stopped, it should be started again, if running - we have
	// a closer timepoint, thus can re-start it.
	if rt.timer != nil && !rt.timer.Stop() {
		<-rt.timer.C
	}

	rt.cutoff = t
	rt.timer = time.AfterFunc(t.Sub(time.Now()), fn)
}
