package main

// Map (Hash Table / Dictionary)
// Maps keys to values for efficient O(1) lookup. Put, Get, Delete - O(1) average, increases with collisions
// A Map is backed by an Array
// Hash Function to convert a Key into an integer (an index) for that Array
// If load factor gets too large, >0.75 usually, increase capacity and rehash
// Because an array has a fixed size, multiple different keys might hash to the same index
// Two common ways to handle these collisions:
//		Separate Chaining: Each array slot holds a pointer to a Linked List of entries. If two keys hash to index 5, they both get added to the linked list at index 5
//		Open Addressing: If a slot is taken, we scan the array for the next available empty slot and put it there

// Entry - single key-value pair, like a Linked List node
type Entry struct {
	key		string
	value	int
	next	*Entry // Pointer to the next entry in case of a collision
}

type HashMap struct {
	buckets		[]*Entry // Array of pointers to Entries (the buckets)
	capacity	int      // Buckets in the array
	size		int      // Total key-value pairs stored
}

func NewHashMap(capacity int) *HashMap {

	if capacity <= 0 {
		capacity = 16 // Standard default capacity
	}

	return &HashMap{
		buckets:	make([]*Entry, capacity), // An array filled with nil pointers
		capacity:	capacity,
		size:		0,
	}
}

// hash turns a string key into an array index
// Iterates over the characters of the string, calculates a large numeric value, and using modulo to fit it into the array
func (m *HashMap) hash(key string) int {

	var hashVal int
	for i := 0; i < len(key); i++ {
		// Multiply current hash by a prime number (31) and add the ASCII value of the character
		hashVal = (hashVal * 31) + int(key[i])
	}

	// Ensure the hash is a positive number
	if hashVal < 0 {
		hashVal = -hashVal
	}

	// Modulo operator (%) ensures the final index fits within the bucket array bounds (0 to capacity - 1)
	return hashVal % m.capacity
}

// Put - inserts a new key-value pair into the map, or updates an existing key
func (m *HashMap) Put(key string, value int) {

	index := m.hash(key)

	// Get the head of the linked list at this specific bucket
	head := m.buckets[index]

	// Walk the linked list to see if the key already exists
	current := head
	for current != nil {

		if current.key == key {
			current.value = value // Found it, update the value and return
			return
		}

		current = current.next
	}

	// The key does not exist, create new Entry
	// Insert it at the front of the linked list for this bucket (O(1) insertion time)
	newEntry := &Entry{
		key:	key,
		value:	value,
		next:	head, // Point new entry's 'next' to the old head
	}

	// Update the bucket array to point new entry as the new head
	m.buckets[index] = newEntry
	m.size++
}

// Get - retrieves the value for a given key. Returns the value and a boolean indicating if it was successfully found
func (m *HashMap) Get(key string) (int, bool) {

	index := m.hash(key)

	current := m.buckets[index]
	
	// Traverse the linked list at this bucket (usually only 1 or 2 items long if the hash function is good)
	for current != nil {

		if current.key == key {
			return current.value, true // Found it
		}

		current = current.next
	}

	return 0, false // Not found
}

// Delete - delete a key-value pair from the map
func (m *HashMap) Delete(key string) {

	index := m.hash(key)

	current := m.buckets[index]
	var prev *Entry

	// Traverse the linked list to find the exact key to delete
	for current != nil {

		if current.key == key {
			
			if prev == nil {
				// The entry to delete is the head of the bucket
				m.buckets[index] = current.next
			} else {
				// The entry is in the middle or end of the list.
				prev.next = current.next
			}
			
			m.size--
			return
		}
		
		prev = current
		current = current.next
	}
}

// Has - check if a key exists in the map
func (m *HashMap) Has(key string) bool {
	_, found := m.Get(key)
	return found
}

// Size - returns the total number of items stored in the map
func (m *HashMap) Size() int {
	return m.size
}

// Clear - removes all elements from the map
func (m *HashMap) Clear() {
	m.buckets = make([]*Entry, m.capacity)
	m.size = 0
}
