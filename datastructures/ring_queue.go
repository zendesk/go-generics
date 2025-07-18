package datastructures

import (
	"iter"
	"sync"
)

type RingQueue[T any] struct {
	data         []T
	start        int // start idx
	end          int // end idx
	size         int // size of the queue (non empty elements)
	capacity     int
	synchronized bool
	mu           sync.RWMutex
}

// NewRingQueue creates a new RingQueue with the given capacity and synchronization flag. If synchronized, operations require locks.
func NewRingQueue[T any](capacity int, synchronized bool) *RingQueue[T] {
	return &RingQueue[T]{
		data:         make([]T, capacity), // buffer allocation
		synchronized: synchronized,
		capacity:     capacity,
		size:         0,
		start:        0,
		end:          -1, // -1 is 1 behind 0 (start)
	}
}

// Push adds a new element to the queue. If the queue is full, it overwrites the oldest element.
func (r *RingQueue[T]) Push(elem T) {
	if r.synchronized {
		r.mu.Lock()
		defer r.mu.Unlock()
	}
	// if full, move start to the next element, overwrite start with new element
	if r.size == r.capacity {
		r.end = r.start
		r.data[r.start] = elem
		r.start = (r.start + 1) % r.capacity
		return
	}

	// If not full, add element to available space and shift end forward
	next := (r.end + 1) % r.capacity
	r.data[next] = elem
	r.end = next

	if r.size < r.capacity {
		r.size = r.size + 1
	}
}

// Pop removes the oldest element in the queue and returns it. If the queue is empty, it returns the zero value of the type.
func (r *RingQueue[T]) Pop() T {
	if r.synchronized {
		r.mu.Lock()
		defer r.mu.Unlock()
	}

	var res T // "zero" element (respective of the type)
	if r.size == 0 {
		return res
	}

	res = r.data[r.start]                // copy over the first element in the queue
	r.start = (r.start + 1) % r.capacity // move the start of the queue
	r.size = r.size - 1
	return res
}

// Items returns all elements in the queue. The order is oldest to newest.
func (r *RingQueue[T]) Items() []T {
	if r.synchronized {
		r.mu.RLock()
		defer r.mu.RUnlock()
	}
	return r.toSlice()
}

// Size returns the number of non-zero elements in the queue.
func (r *RingQueue[T]) Size() int {
	if r.synchronized {
		r.mu.RLock()
		defer r.mu.RUnlock()
	}
	return r.size
}

// Iter2 returns an iterator that yields all elements in the ring queue oldest to newest. Empty elements are not yielded
func (r *RingQueue[T]) Iter2() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		queueCopy := r.toSlice()
		r.mu.RLock()
		size := r.size
		r.mu.RUnlock()
		for i := 0; i < size; i++ {
			var val T
			if r.synchronized {
				r.mu.RLock()
				val = queueCopy[i]
				r.mu.RUnlock()
			} else {
				val = queueCopy[i]
			}
			if !yield(i, val) {
				return
			}
		}
	}
}

// Iter returns an iterator that yields all elements in the ring queue oldest to newest. Empty elements are not yielded
func (r *RingQueue[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		queueCopy := r.toSlice()
		r.mu.RLock()
		size := r.size
		r.mu.RUnlock()
		for i := 0; i < size; i++ {
			var val T
			if r.synchronized {
				r.mu.RLock()
				val = queueCopy[i]
				r.mu.RUnlock()
			} else {
				val = queueCopy[i]
			}
			if !yield(val) {
				return
			}
		}
	}
}

func (r *RingQueue[T]) toSlice() []T {
	if r.size == 0 {
		return []T{}
	}

	part1 := r.data[r.start:]
	if r.start+r.size < r.capacity {
		part1 = r.data[r.start : r.start+r.size]
	}
	var queueCopy = make([]T, 0)
	queueCopy = append(queueCopy, part1...)
	if r.end < r.start {
		part2 := r.data[:r.end+1]
		queueCopy = append(queueCopy, part2...)
	}
	return queueCopy
}
