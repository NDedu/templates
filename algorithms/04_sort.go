package main

// BubbleSort O(n^2)
func BubbleSort(array []int) {

	for i := range len(array) {
		for j := range (len(array) - 1 - i) {

			if array[j] > array[j + 1] {
				tmp := array[j]
				array[j] = array[j + 1]
				array[j + 1] = tmp
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
			array[j + 1] = array[j]
			j--
		}
		array[j + 1] = tmp
	}
}

// QuickSort O(n log n) average; O(n^2) worst
func QuickSort(array []int) {

	if len(array) < 2 {
		return
	}

	left := 0
	right := len(array) - 1
	pivot := len(array) / 2

	// Move the pivot to the far right
	tmp := array[pivot]
	array[pivot] = array[right]
	array[right] = tmp

	// Move elements smaller than pivot to the left
	for i := range array {
		if array[i] < array[right] {
			tmp = array[left]
			array[left] = array[i]
			array[i] = tmp
			left++
		}
	}

	// Move the pivot to its final place
	tmp = array[left]
	array[left] = array[right]
	array[right] = tmp

	// Recursively sort the left and right sides
	QuickSort(array[:left])
	QuickSort(array[left+1:])
}

// MergeSort O(n log n)
func MergeSort(array []int) []int{

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
	for i < len(left) && j < len(right) {
		if left[i] < right[j] {
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
