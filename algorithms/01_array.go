package main

import "unsafe"

// Array - contiguous space in memory that can't grow
// Insertion - we set the value at an idex; Deletion - we set the sentinal value (0, null)
// Get, Insert (Set) or Push on Dynamic Array, Delete, Get - O(1) - we don't walk the array, and we can jump to the address with: TargetAddress = BaseAddress + (Index * ItemSize)
// Array is a Base Pointer (a pointer to the very first slot) and because the memory is contiguous (every slot is perfectly side-by-side), we just use math to get addresses (not next/prev like linked lists).
// Issue with Array, it needs to allocate the memory upfront. For Lists memory is allocated as values are added.

// Implementation only for int array

type Array struct {
	basePointer	unsafe.Pointer // The raw memory address of the very first slot
	itemSize	uintptr        // How many bytes a single item takes (e.g., 8 bytes for an int)
	capacity	int
}

func NewArray(capacity int) *Array {

	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	// We allocate a generic block of bytes, and grab the raw memory address and throw the slice away
	itemSize := unsafe.Sizeof(int(0))
	rawMemory := make([]byte, uintptr(capacity)*itemSize)

	return &Array{
		basePointer:	unsafe.Pointer(&rawMemory[0]),
		itemSize:		itemSize,
		capacity:		capacity,
	}
}

func (arr *Array) Len() int {
	return arr.capacity
}

func (arr *Array) Get(index int) int {

	if index < 0 || index >= arr.capacity {
		panic("index out of bounds")
	}

	targetAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(index) * arr.itemSize))
	return *(*int)(targetAddress)
}

// Set physically overwrites the memory address
func (arr *Array) Set(index int, val int) {

	if index < 0 || index >= arr.capacity {
		panic("index out of bounds")
	}

	// An integer takes up 8 bytes of space, finding index 5 is just 1000 + (5 * 8) = 1040. The CPU jumps directly to address 1040 in O(1) time.
	// TargetAddress=BaseAddress+(Index×ItemSize)
	
	// Calculate how far down the memory block we need to jump
	memoryOffset := uintptr(index) * arr.itemSize

	// Add the offset to our starting address
	targetAddress := unsafe.Pointer(uintptr(arr.basePointer) + memoryOffset)

	// Treat raw memory address as an integer, and add this value to it
	*(*int)(targetAddress) = val // typecast and dereference for int array
}

// Clear - O(n)
func (arr *Array) Clear() {

	// Start at the base address and overwrite every byte jump by jump
	for i := 0; i < arr.capacity; i++ {

		memoryOffset := uintptr(i) * arr.itemSize
		targetAddress := unsafe.Pointer(uintptr(arr.basePointer) + memoryOffset)
		
		// Overwrite with the zero-value for an int
		*(*int)(targetAddress) = 0
	}
}

// IndexOf - O(n)
func (arr *Array) IndexOf(targetValue int) int {

	for i := 0; i < arr.capacity; i++ {
		
		offset := uintptr(i) * arr.itemSize
		targetAddress := unsafe.Pointer(uintptr(arr.basePointer) + offset)
		
		if *(*int)(targetAddress) == targetValue {
			return i
		}
	}
	
	// Found nothing
	return -1
}

// Clone - O(n)
func (arr *Array) Clone() *Array {

	// Allocate a brand new array of the same size
	newArr := NewArray(arr.capacity)

	// Mathematically walk through both arrays and copy the values
	for i := 0; i < arr.capacity; i++ {
		
		// Calculate the jump
		offset := uintptr(i) * arr.itemSize
		
		// Find the addresses for both the old and new
		oldAddress := unsafe.Pointer(uintptr(arr.basePointer) + offset)
		newAddress := unsafe.Pointer(uintptr(newArr.basePointer) + offset)
		
		// Read from the old address, and write to the new address
		*(*int)(newAddress) = *(*int)(oldAddress)
	}

	return newArr
}

// Dynamic Array / ArrayList - can grow

type DynamicArray struct {
	basePointer	unsafe.Pointer
	itemSize	uintptr
	capacity	int
	length		int
}

func NewDynamicArray() *DynamicArray {

	initialCapacity := 2 // because of resize math
	itemSize := unsafe.Sizeof(int(0))
	
	// Allocate the initial raw memory block
	rawMemory := make([]byte, uintptr(initialCapacity) * itemSize)

	return &DynamicArray{
		basePointer:	unsafe.Pointer(&rawMemory[0]),
		itemSize:		itemSize,
		capacity:		initialCapacity,
		length:			0,
	}
}

// O(n) rarely because of the double capacity, the average is O(1)
func (arr *DynamicArray) resize() {

	// Always double the capacity
	newCapacity := arr.capacity * 2

	// New pointer
	newRawMemory := make([]byte, uintptr(newCapacity) * arr.itemSize)
	newBasePointer := unsafe.Pointer(&newRawMemory[0])

	// Copy data from old pointer to new pointer
	for i := 0; i < arr.length; i++ {

		offset := uintptr(i) * arr.itemSize
		
		oldAddress := unsafe.Pointer(uintptr(arr.basePointer) + offset)
		newAddress := unsafe.Pointer(uintptr(newBasePointer) + offset)
		
		// Read from old, write to new
		*(*int)(newAddress) = *(*int)(oldAddress)
	}

	// The old memory block is abandoned
	arr.basePointer = newBasePointer
	arr.capacity = newCapacity
}

func (arr *DynamicArray) Len() int {
	return arr.length
}

func (arr *DynamicArray) Get(index int) int {

	if index < 0 || index >= arr.length {
		panic("index out of bounds")
	}

	targetAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(index) * arr.itemSize))
	return *(*int)(targetAddress)
}

// Push - O(1) if array not full
func (arr *DynamicArray) Push(val int) {

	// Array full check
	if arr.length == arr.capacity {
		arr.resize()
	}

	targetAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(arr.length) * arr.itemSize))

	// Write the value into raw memory
	*(*int)(targetAddress) = val
	
	arr.length++
}

// Pop - O(1)
func (arr *DynamicArray) Pop() int {
	
	if arr.length == 0 {
		panic("array is empty")
	}

	// Calculate the address of the last item
	targetAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(arr.length-1) * arr.itemSize))

	val := *(*int)(targetAddress)

	arr.length--

	return val
}

// Prepend - O(n)
func (arr *DynamicArray) Prepend(val int) {

	if arr.length == arr.capacity {
		arr.resize()
	}

	// Shift everything to the right by one to make index 0 empty then loop backwards
	for i := arr.length; i > 0; i-- {

		sourceAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(i-1) * arr.itemSize))
		destinationAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(i) * arr.itemSize))
		
		*(*int)(destinationAddress) = *(*int)(sourceAddress)
	}

	*(*int)(arr.basePointer) = val
	
	arr.length++
}

// Shift - O(n)
func (arr *DynamicArray) Shift() int {

	if arr.length == 0 {
		panic("array is empty")
	}

	val := *(*int)(arr.basePointer)

	// Shift everything to the left then loop forward
	for i := 0; i < arr.length-1; i++ {

		sourceAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(i+1) * arr.itemSize))
		destinationAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(i) * arr.itemSize))
		
		*(*int)(destinationAddress) = *(*int)(sourceAddress)
	}

	arr.length--

	return val
}

func (arr *DynamicArray) InsertAt(index int, val int) {

	if index < 0 || index > arr.length {
		panic("index out of bounds")
	}

	if arr.length == arr.capacity {
		arr.resize()
	}

	// Shift everything to the right and loop backwards
	for i := arr.length; i > index; i-- {
		
		// Address of the item we want to move
		sourceAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(i-1) * arr.itemSize))
		
		// Address of the new slot it is moving into
		destinationAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(i) * arr.itemSize))

		// Copy data forward
		*(*int)(destinationAddress) = *(*int)(sourceAddress)
	}

	targetAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(index) * arr.itemSize))
	*(*int)(targetAddress) = val

	arr.length++
}

func (arr *DynamicArray) RemoveAt(index int) {

	if index < 0 || index >= arr.length {
		panic("index out of bounds")
	}

	// Shift everything from the right of the gap one step to the left
	for i := index; i < arr.length-1; i++ {
		
		sourceAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(i+1) * arr.itemSize))
		
		destinationAddress := unsafe.Pointer(uintptr(arr.basePointer) + (uintptr(i) * arr.itemSize))

		// Copy the data backward
		*(*int)(destinationAddress) = *(*int)(sourceAddress)
	}

	arr.length--
}
