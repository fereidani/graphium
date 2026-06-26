package graphium

// heapItem pairs a tentative key (a path length or f-score) with a node index.
type heapItem[T Number] struct {
	key  T
	node int
}

// minHeap is a binary min-heap of heapItem ordered by key, breaking ties by node
// index so that traversals are deterministic.
//
// It is specialized for the shortest-path algorithms instead of container/heap
// to avoid interface conversion and boxing on every push and pop.
type minHeap[T Number] struct {
	data []heapItem[T]
}

func newMinHeap[T Number](cap int) *minHeap[T] {
	return &minHeap[T]{data: make([]heapItem[T], 0, cap)}
}

func (h *minHeap[T]) Len() int { return len(h.data) }

func (h *minHeap[T]) less(i, j int) bool {
	if h.data[i].key != h.data[j].key {
		return h.data[i].key < h.data[j].key
	}
	return h.data[i].node < h.data[j].node
}

// Push inserts the item and restores the heap invariant.
func (h *minHeap[T]) Push(key T, node int) {
	h.data = append(h.data, heapItem[T]{key: key, node: node})
	h.up(len(h.data) - 1)
}

// Pop removes and returns the minimum item. It panics if the heap is empty; the
// shortest-path algorithms only call Pop after checking Len.
func (h *minHeap[T]) Pop() heapItem[T] {
	n := len(h.data)
	top := h.data[0]
	h.data[0] = h.data[n-1]
	h.data = h.data[:n-1]
	if n-1 > 0 {
		h.down(0)
	}
	return top
}

// up moves the element at i up until the heap order is restored.
func (h *minHeap[T]) up(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !h.less(i, parent) {
			return
		}
		h.data[i], h.data[parent] = h.data[parent], h.data[i]
		i = parent
	}
}

// down moves the element at i down until the heap order is restored.
func (h *minHeap[T]) down(i int) {
	n := len(h.data)
	for {
		l := 2*i + 1
		if l >= n {
			return
		}
		smallest := l
		if r := l + 1; r < n && h.less(r, l) {
			smallest = r
		}
		if !h.less(smallest, i) {
			return
		}
		h.data[i], h.data[smallest] = h.data[smallest], h.data[i]
		i = smallest
	}
}
