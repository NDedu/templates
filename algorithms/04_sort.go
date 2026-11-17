package main

// BubbleSort O(n^2)
func BubbleSort(array []int) {

	for i := range len(array) {
		for j := range len(array) - 1 - i {

			if array[j] > array[j+1] {
				tmp := array[j]
				array[j] = array[j+1]
				array[j+1] = tmp
			}
		}
	}
}

// InsertionSort O(n^2)
func InsertionSort(array []int) {

	for i := 1; i < len(array); i++ {

		tmp := array[i]
		j := i - 1

		for j >= 0 && array[j] > tmp {
			array[j+1] = array[j]
			j--
		}
		array[j+1] = tmp
	}
}

// QuickSort O(n log n) average; O(n^2) worst
// Hoare partition: both sides stop at values equal to the pivot and swap them, so duplicates are split evenly
// between the two halves (a partition that sends every equal value to the same side is O(n^2) on duplicates)
func QuickSort(array []int) {

	for len(array) > 1 {

		// The lower middle: with it, both sides of the partition below are never empty
		pivot := array[(len(array)-1)/2]

		// Walk inwards from both ends and swap each pair that is on the wrong side
		i := -1
		j := len(array)
		for {
			i++
			for array[i] < pivot {
				i++
			}

			j--
			for array[j] > pivot {
				j--
			}

			if i >= j {
				break
			}

			tmp := array[i]
			array[i] = array[j]
			array[j] = tmp
		}

		// [0, j] is <= pivot and [j+1, end] is >= pivot
		// Recurse into the smaller side and loop on the bigger one, so the stack depth stays O(log n)
		left := array[:j+1]
		right := array[j+1:]

		if len(left) < len(right) {
			QuickSort(left)
			array = right
		} else {
			QuickSort(right)
			array = left
		}
	}
}

// MergeSort O(n log n), stable: equal values keep their original order
func MergeSort(array []int) []int {

	if len(array) <= 1 {
		return array
	}
	mid := len(array) / 2
	left := MergeSort(array[:mid])
	right := MergeSort(array[mid:])

	return merge(left, right)
}

func merge(left, right []int) []int {

	result := make([]int, len(left)+len(right))

	i := 0 // Index for left array
	j := 0 // Index for right array
	k := 0 // Index for the new result array

	// Compare and copy into the result array
	// <= takes from the left on ties: the left half came first in the input, which keeps the sort stable
	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			result[k] = left[i]
			i++
		} else {
			result[k] = right[j]
			j++
		}
		k++
	}

	// Copy any remaining elements from left
	for i < len(left) {
		result[k] = left[i]
		i++
		k++
	}

	// Copy any remaining elements from right
	for j < len(right) {
		result[k] = right[j]
		j++
		k++
	}

	return result
}

// HeapSort O(n log n)
func HeapSort(array []int) {

	n := len(array)

	// Build max heap from the bottom up
	for i := n/2 - 1; i >= 0; i-- {
		heapify(array, n, i)
	}

	// Extract elements one by one from the heap
	for i := n - 1; i > 0; i-- {
		tmp := array[0]
		array[0] = array[i]
		array[i] = tmp

		heapify(array, i, 0)
	}
}

func heapify(arr []int, n, i int) {

	largest := i
	left := 2*i + 1
	right := 2*i + 2

	if left < n && arr[left] > arr[largest] {
		largest = left
	}
	if right < n && arr[right] > arr[largest] {
		largest = right
	}
	if largest != i {
		tmp := arr[i]
		arr[i] = arr[largest]
		arr[largest] = tmp

		heapify(arr, n, largest)
	}
}
