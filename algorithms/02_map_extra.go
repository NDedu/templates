package main

import (
	"sync"
)

// TreeMap - Ordered Map
// Built using a Binary Search Tree (BST)
// This would be a Self-Balancing Tree to guarantee O(log n) time
// Maintains keys in sorted alphabetical order

type TreeMapNode struct {
	key		string
	value	int
	left	*TreeMapNode
	right	*TreeMapNode
}

type TreeMap struct {
	root	*TreeMapNode
	size	int
}

func NewTreeMap() *TreeMap {
	return &TreeMap{}
}

func (t *TreeMap) Put(key string, value int) {

	if t.root == nil {
		t.root = &TreeMapNode{key: key, value: value}
		t.size++
		return
	}

	t.root.insert(key, value, &t.size)
}

func (n *TreeMapNode) insert(key string, value int, size *int) {

	if key == n.key {
		n.value = value // Update existing
		return
	}

	if key < n.key {

		if n.left == nil {
			n.left = &TreeMapNode{key: key, value: value}
			*size++
		} else {
			n.left.insert(key, value, size)
		}

	} else {

		if n.right == nil {
			n.right = &TreeMapNode{key: key, value: value}
			*size++
		} else {
			n.right.insert(key, value, size)
		}
	}
}

func (t *TreeMap) Get(key string) (int, bool) {

	if t.root == nil {
		return 0, false
	}

	return t.root.get(key)
}

func (n *TreeMapNode) get(key string) (int, bool) {

	if n == nil {
		return 0, false
	}

	if key == n.key {
		return n.value, true
	}

	if key < n.key {
		return n.left.get(key)
	}

	return n.right.get(key)
}

// InOrderKeys - returns the keys in sorted alphabetical order
func (t *TreeMap) InOrderKeys() []string {

	var keys []string
	if t.root != nil {
		t.root.inOrder(&keys)
	}

	return keys
}

func (n *TreeMapNode) inOrder(keys *[]string) {

	if n.left != nil {
		n.left.inOrder(keys)
	}

	*keys = append(*keys, n.key)
	if n.right != nil {
		n.right.inOrder(keys)
	}
}


// LinkedHashMap - Insertion-Ordered Map
// Array of pointers for HashMap lookups (O(1)), and a Doubly-Linked List spanning all entries to remember the exact order they were inserted

type LinkedMapEntry struct {
	key				string
	value			int
	nextInBucket	*LinkedMapEntry // For the HashMap (Collision chaining)
	prevInList		*LinkedMapEntry // For the Doubly-Linked List Insertion order
	nextInList		*LinkedMapEntry
}

type LinkedHashMap struct {
	buckets		[]*LinkedMapEntry
	capacity	int
	size		int
	head		*LinkedMapEntry
	tail		*LinkedMapEntry
}

func NewLinkedHashMap(capacity int) *LinkedHashMap {

	if capacity <= 0 {
		capacity = 16
	}

	return &LinkedHashMap{
		buckets:	make([]*LinkedMapEntry, capacity),
		capacity:	capacity,
	}
}

func (m *LinkedHashMap) hash(key string) int {
	
	var hashVal int
	for i := 0; i < len(key); i++ {
		hashVal = (hashVal * 31) + int(key[i])
	}
	
	if hashVal < 0 {
		hashVal = -hashVal
	}
	
	return hashVal % m.capacity
}

func (m *LinkedHashMap) Put(key string, value int) {
	index := m.hash(key)
	current := m.buckets[index]

	// Check if key exists; if so, update
	for current != nil {

		if current.key == key {
			current.value = value
			return
		}

		current = current.nextInBucket
	}

	// Create new entry
	newEntry := &LinkedMapEntry{
		key:			key,
		value:			value,
		nextInBucket:	m.buckets[index], // Insert at front of bucket list
	}

	m.buckets[index] = newEntry

	// Append to the doubly-linked list to remember insertion order
	if m.tail == nil {
		m.head = newEntry
		m.tail = newEntry
	} else {
		newEntry.prevInList = m.tail
		m.tail.nextInList = newEntry
		m.tail = newEntry
	}

	m.size++
}

func (m *LinkedHashMap) Get(key string) (int, bool) {

	index := m.hash(key)
	current := m.buckets[index]
	
	for current != nil {
		
		if current.key == key {
			return current.value, true
		}
		
		current = current.nextInBucket
	}
	
	return 0, false
}

// GetKeysInOrder - returns the keys in order they were inserted
func (m *LinkedHashMap) GetKeysInOrder() []string {

	var keys []string
	current := m.head
	
	for current != nil {
		keys = append(keys, current.key)
		current = current.nextInList
	}

	return keys
}


// ConcurrentHashMap - Thread-Safe Map with Bucket-Level Locking
// Instead of locking the entire map during a write (which blocks everyone else), only lock the specific bucket, this allows multiple threads as long as they touch different buckets

type ConcurrentEntry struct {
	key		string
	value	int
	next	*ConcurrentEntry
}

// Bucket - wrap a regular bucket in a RWMutex so each bucket can be locked independently
type Bucket struct {
	head	*ConcurrentEntry
	lock	sync.RWMutex
}

type ConcurrentHashMap struct {
	buckets		[]*Bucket
	capacity	int
}

func NewConcurrentHashMap(capacity int) *ConcurrentHashMap {

	if capacity <= 0 {
		capacity = 16
	}
	
	buckets := make([]*Bucket, capacity)
	for i := 0; i < capacity; i++ {
		buckets[i] = &Bucket{}
	}
	
	return &ConcurrentHashMap{
		buckets:	buckets,
		capacity:	capacity,
	}
}

func (m *ConcurrentHashMap) hash(key string) int {
	
	var hashVal int
	for i := 0; i < len(key); i++ {
		hashVal = (hashVal * 31) + int(key[i])
	}
	
	if hashVal < 0 {
		hashVal = -hashVal
	}
	
	return hashVal % m.capacity
}

func (m *ConcurrentHashMap) Put(key string, value int) {
	
	index := m.hash(key)
	bucket := m.buckets[index]

	// Only lock this specific bucket. Other threads can write to other buckets simultaneously
	bucket.lock.Lock()
	defer bucket.lock.Unlock()

	current := bucket.head
	for current != nil {

		if current.key == key {
			current.value = value
			return
		}

		current = current.next
	}

	newEntry := &ConcurrentEntry{
		key:	key,
		value:	value,
		next:	bucket.head,
	}

	bucket.head = newEntry
}

func (m *ConcurrentHashMap) Get(key string) (int, bool) {

	index := m.hash(key)
	bucket := m.buckets[index]

	// Allows multiple simultaneous readers on this bucket, but no writers
	bucket.lock.RLock()
	defer bucket.lock.RUnlock()

	current := bucket.head
	for current != nil {

		if current.key == key {
			return current.value, true
		}

		current = current.next
	}

	return 0, false
}
