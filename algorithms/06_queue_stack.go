package main

// Queue - First in first out
type Queue struct {
	Length	int
	Head	*Node
	Tail	*Node
}

// Stack - Last in first out
type Stack struct {
	Length	int
	Head	*Node
}

// Queue FIFO - all operations O(1)

// Enqueue - add element to the Queue
func (queue *Queue) Enqueue(val int) {

	newNode := &Node{ Value: val }

	if queue.Length == 0 {

		queue.Head = newNode
		queue.Tail = queue.Head
		queue.Length++
		return
	}

	queue.Tail.Next = newNode
	queue.Tail = newNode

	queue.Length++
}

// Deque - remove element from Queue
func (queue *Queue) Deque() int {

	if queue.Length == 0 {
		panic("queue is empty")
	}

	headValue := queue.Head.Value
	queue.Head = queue.Head.Next
	queue.Length--

	// If we reach the end of the queue set Tail to null
	if queue.Length == 0 {
		queue.Tail = nil
	}

	return headValue
}

// Peek - look at the element of the Queue that will be "next"
func (queue *Queue) Peek() int {

	if queue.Length == 0 {
		panic("queue is empty")
	}

	return queue.Head.Value
}

// Stack LIFO - all operations O(1)

// Push - add element to stack
func (stack *Stack) Push(val int) {

	newNode := &Node{Value: val}

	newNode.Next = stack.Head
	stack.Head = newNode

	stack.Length++
}

// Pop - remove "top" element, most recent added added (Head)
func (stack *Stack) Pop() int {

	if stack.Length == 0 {
		panic("stack is empty")
	}

	headValue := stack.Head.Value

	stack.Head = stack.Head.Next
	stack.Length--

	return headValue
}

// Peek - look at the "top" element (Head)
func (stack *Stack) Peek() int {

	if stack.Length == 0 {
		panic("stack is empty")
	}

	return stack.Head.Value
}
