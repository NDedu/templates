package main

import "math"

// All consider inclusive low and exclusive high [low, high)

// LinearSearch O(n)
func LinearSearch(haystack []int, needle int) bool {

	for i := range len(haystack) {
		if haystack[i] == needle {
			return true
		}
	}
	return false
}

// BinarySearch O(log n)
func BinarySearch(haystack []int, needle int) bool {

	low := 0
	high := len(haystack)

	for low < high {
		midpoint := low + (high - low) / 2
		value := haystack[midpoint]

		if value == needle {
			return true
		} else if value < needle {
			low = midpoint + 1
		} else {
			high = midpoint
		}
	}
	return false
}

// BinarySearchRange O(log n)
func BinarySearchRange(haystack []int, low int, high int, needle int) bool {

    for low < high {
        mid := low + (high-low)/2
        if haystack[mid] == needle {
            return true
        }
        if haystack[mid] < needle {
            low = mid + 1
        } else {
            high = mid
        }
    }
    return false
}

// InterpolationSearch for sorted arrays where data is uniformly distributed (like IDs)
// Average 0(log(log n)) - worst O(n). Performs good for 10 20 30 and worst for 10 100 1000 (exponential distributed data)
func InterpolationSearch(haystack []int, needle int) bool {

	low := 0
	high := len(haystack)

	for low < high && needle >= haystack[low] && needle <= haystack[high-1] {
		
		// If the range has one element and needle is within range, it must be this one
		if haystack[low] == haystack[high - 1] {
			return haystack[low] == needle
		}

		// Formula: low + [(needle - min) * (max_idx - min_idx) / (max - min)]
		pos := low + int(float64(needle - haystack[low]) * float64((high - 1) - low) / (float64(haystack[high - 1]) - float64(haystack[low])))

		if haystack[pos] == needle {
			return true
		}
		
		if haystack[pos] < needle {
			low = pos + 1
		} else {
			high = pos
		}
	}
	return false
}

// ExponentialSearch best for sorted arrays where size unknown, or the needle is likely at the beginning
// Average O(log i) where i is the index of the needle - worst O(log n)
func ExponentialSearch(haystack []int, needle int) bool {

	n := len(haystack)
	if n == 0 {
		return false
	}
	if haystack[0] == needle {
		return true
	}

	// Find the range where the needle might exist
	high := 1
	for high < n && haystack[high] <= needle {
		high *= 2
	}

	return BinarySearchRange(haystack, high/2, min(high, n), needle)
}

// JumpSearch useful when the cost of "stepping back" is higher than "jumping forward", O(sqrt n)
func JumpSearch(haystack []int, needle int) bool {

	n := len(haystack)
	if n == 0 {
		return false
	}

	step := int(math.Sqrt(float64(n)))
	low := 0
	high := step

	// Jump blocks until the end of the block is greater than or equal to the needle or we reach the end of the haystack.
	for high < n && haystack[high - 1] < needle {
		low = high
		high += step
	}

	// Adjust high to n if it overshot the array length
	if high > n {
		high = n
	}

	// Linear search within the range [low, high)
	for i := low; i < high; i++ {
		if haystack[i] == needle {
			return true
		}
		// Since it's sorted, if value > needle, it's not here
		if haystack[i] > needle {
			return false
		}
	}
	return false
}

// TwoCrystalBalls - similar to JumpSearch
// Two Crystal Balls Problem - Having 2 crystal balls find at which point they break (basically an array[0, 0, 0, 0, ..., 1, 1, 1, ...])
// We jump sqrt n until it breaks, and then we linear search from the last good position
func TwoCrystalBalls(breaks []bool) int {

	jumpAmount := int(math.Sqrt(float64(len(breaks))))
	i := jumpAmount

	for ; i < len(breaks); i += jumpAmount {
		if breaks[i] {
			break
		}
	}

	i -= jumpAmount
	for j := 0; j < jumpAmount && i < len(breaks); j++ {
		if breaks[i] {
			return i
		}
		i++
	}

	return -1
}
