package main

// RingBuffer a fixed-size circular queue that maintains order
// Avoids shifting and Enqueue and Dequeue are 0(1) (if not full)
// Tail acts as the point where new elements are added and Head acts as the point where data leaves
type RingBuffer struct {
	data		[]int
	capacity	int
	head		int // The read pointer (oldest data)
	tail		int // The write pointer (next empty slot)
	length		int
}

func NewRingBuffer(capacity int) *RingBuffer {

	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	return &RingBuffer{
		data:		make([]int, capacity),
		capacity:	capacity,
		head:		0,
		tail:		0,
		length:		0,
	}
}

// Enqueue - O(1) adds an item to the buffer
func (rb *RingBuffer) Enqueue(val int) {

	if rb.length == rb.capacity {
		panic("ring buffer is full")
	}

	rb.data[rb.tail] = val

	// Move the tail forward
	// The % wraps it back to 0 if it hits the capacity (circular)
	rb.tail = (rb.tail + 1) % rb.capacity

	rb.length++
}

// Dequeue - O(1) reads and removes the oldest item
func (rb *RingBuffer) Dequeue() int {

	if rb.length == 0 {
		panic("ring buffer is empty")
	}

	val := rb.data[rb.head]

	// Move the head forward
	rb.head = (rb.head + 1) % rb.capacity

	rb.length--

	return val
}

// Peek - O(1) looks at the oldest item
func (rb *RingBuffer) Peek() int {

	if rb.length == 0 {
		panic("ring buffer is empty")
	}

	return rb.data[rb.head]
}
