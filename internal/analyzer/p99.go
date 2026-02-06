package analyzer

import (
	"container/heap"
)

type LogEntry struct {
	Latency   float64
	Data      string
	Timestamp string
}

type MinHeap []LogEntry

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].Latency < h[j].Latency }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(LogEntry))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type P99Tracker struct {
	Heap    *MinHeap
	MaxSize int
}

func NewTracker(size int) *P99Tracker {
	h := &MinHeap{}
	heap.Init(h)
	return &P99Tracker{Heap: h, MaxSize: size}
}

func (t *P99Tracker) Process(entry LogEntry) {
	if t.Heap.Len() < t.MaxSize {
		heap.Push(t.Heap, entry)
	} else if entry.Latency > (*t.Heap)[0].Latency {
		heap.Pop(t.Heap)
		heap.Push(t.Heap, entry)
	}
}
