package main

// Heap - A priority Queue
// Every child and grand child is smaller (MaxHeap - root is the biggest) or larger (MinHeap - root smallest) than the current node
// With insert/delete the tree must be adjusted, there is no traversal
// Always add from left to right, no gaps
// There are no link/left/right, nodes have an index (starting from 0 root) and the relations are established with math equations
// Children are at 2 * i + 1 (i index of parent +1 or 2); parent is at floor((i - 1) / 2) (i index of child)

// MinHeap - root is the smallest
type MinHeap struct {
	array	[]int
}

// Insert - O(log n)
func (heap *MinHeap) Insert(val int) {

	heap.array = append(heap.array, val)
	heapifyUp(heap.array, len(heap.array) - 1, true)
}

// Delete - O(log n)
func (heap *MinHeap) Delete() int {

	if len(heap.array) == 0 {
		return 0
	}

	root := heap.array[0]
	last := len(heap.array) - 1

	heap.array[0] = heap.array[last]
	heap.array = heap.array[:last]

	if len(heap.array) > 0 {
		heapifyDown(heap.array, 0, true)
	}

	return root
}

// MaxHeap - root is the biggest
type MaxHeap struct {
	array	[]int
}

// Insert - O(log n)
func (heap *MaxHeap) Insert(val int) {

	heap.array = append(heap.array, val)
	heapifyUp(heap.array, len(heap.array) - 1, false)
}

// Delete - O(log n)
func (heap *MaxHeap) Delete() int {

	if len(heap.array) == 0 {
		return 0
	}

	root := heap.array[0]
	last := len(heap.array) - 1

	heap.array[0] = heap.array[last]
	heap.array = heap.array[:last]

	if len(heap.array) > 0 {
		heapifyDown(heap.array, 0, false)
	}

	return root
}

func parentIdx(idx int) int {
	return (idx - 1) / 2
}

func leftChildIdx(idx int) int {
	return 2 * idx + 1
}

func rightChildIdx(idx int) int {
	return 2 * idx + 2
}

// shouldMoveUp - MinHeap isMin true; MaxHeap isMin false
func shouldMoveUp(child, parent int, isMin bool) bool {

	if isMin {
		return child < parent
	}

	return child > parent
}

func heapifyUp(array []int, idx int, isMin bool) {

	for idx > 0 {

		parent := parentIdx(idx)

		if !shouldMoveUp(array[idx], array[parent], isMin) {
			return
		}

		tmp := array[idx]
		array[idx] = array[parent]
		array[parent] = tmp
		idx = parent
	}
}

func heapifyDown(array []int, idx int, isMin bool) {

	size := len(array)

	for {

		leftIdx := leftChildIdx(idx)
		rightIdx := rightChildIdx(idx)
		target := idx

		if leftIdx < size && shouldMoveUp(array[leftIdx], array[target], isMin) {
			target = leftIdx
		}

		if rightIdx < size && shouldMoveUp(array[rightIdx], array[target], isMin) {
			target = rightIdx
		}

		if target == idx {
			return
		}

		tmp := array[idx]
		array[idx] = array[target]
		array[target] = tmp
		idx = target
	}
}

/* recursive heapify
// heapifyUp - O(log n)
func heapifyUp(array []int, idx int, isMin bool) {

	if idx == 0 {
		return
	}

	parent := parentIdx(idx)

	if shouldMoveUp(array[idx], array[parent], isMin) {
		tmp := array[idx]
		array[idx] = array[parent]
		array[parent] = tmp
		heapifyUp(array, parent, isMin)
	}
}

// heapifyDown - O(log n)
func heapifyDown(array []int, idx int, isMin bool) {

	leftIdx := leftChildIdx(idx)
	rightIdx := rightChildIdx(idx)
	target := idx

	if leftIdx < len(array) && shouldMoveUp(array[leftIdx], array[target], isMin) {
		target = leftIdx
	}

	if rightIdx < len(array) && shouldMoveUp(array[rightIdx], array[target], isMin) {
		target = rightIdx
	}

	if target != idx {
		tmp := array[idx]
		array[idx] = array[target]
		array[target] = tmp
		heapifyDown(array, target, isMin)
	}
}
*/
