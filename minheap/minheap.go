package minheap

type Heap struct {
	heap []Element
}

type Element struct {
	priority int
	payload interface{}
}

func (e Element) Less(i Element) bool {
	return e.priority < i.priority
}

func (h *Heap) Len() int {
	return len(h.heap)
}

func (h *Heap) Push(priority int, payload interface{}) {
	h.heap = append(h.heap, Element{priority, payload})
	
	var i int
	// While the newly inserted element is less than it's parent, move it up the tree
	for i = h.Len(); i > 1 && h.heap[i - 1].Less(h.heap[i >> 1 - 1]); i >>= 1 {
		h.heap[i - 1], h.heap[i >> 1 - 1] = h.heap[i >> 1 - 1], h.heap[i - 1]
	}
}

func (h *Heap) Pop() (v interface{}) {
	if h.Len() == 0 {
		return nil
	}
	
	// Get the payload of the first element
	v = h.heap[0].payload
	// Move the last child to the root
	h.heap[0] = h.heap[h.Len() - 1]
	// Cut the last element
	h.heap = h.heap[1:]
	
	l := h.Len()
	for i := 1; i <= l; {
		j := i - 1
		left, right := (i << 1) - 1, i << 1
		// if there are two children
		if right < l {
			// if one of the chilren is less than the current node
			if h.heap[left].Less(h.heap[j]) || h.heap[right].Less(h.heap[j]) {
				// if left is smaller than right
				if h.heap[left].Less(h.heap[right]) {
					// swap left with current
					i = left
					h.heap[j], h.heap[left] = h.heap[left], h.heap[j]
					continue
				} else {
					// swap right with current
					i = right
					h.heap[j], h.heap[right] = h.heap[right], h.heap[j]
					continue
				}
			}
		}
		// if there is one child
		if left < l {
			// if left is less than the current node
			if h.heap[left].Less(h.heap[j]) {
				// swap left with current
				i = left
				h.heap[j], h.heap[left] = h.heap[left], h.heap[j]
				continue
			}
		}
		return
	}
	return
}
