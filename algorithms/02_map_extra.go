package main

import (
	"sync"
	"sync/atomic"
)

// TreeMap - Ordered Map
// Built using a Self-Balancing Binary Search Tree (AVL) to guarantee O(log n) time
// A plain BST turns into a linked list (O(n)) when keys arrive in sorted order (timestamps, IDs), the AVL rotations prevent that
// Maintains keys in sorted alphabetical order

type TreeMapNode struct {
	key    string
	value  int
	height int // Height of the subtree rooted here (a leaf is 1), used to detect when a side got too tall
	left   *TreeMapNode
	right  *TreeMapNode
}

type TreeMap struct {
	root *TreeMapNode
	size int
}

func NewTreeMap() *TreeMap {
	return &TreeMap{}
}

// Put - O(log n)
func (t *TreeMap) Put(key string, value int) {

	// insert returns the new root of the subtree, because a rotation can change it
	t.root = t.root.insert(key, value, &t.size)
}

func (n *TreeMapNode) insert(key string, value int, size *int) *TreeMapNode {

	if n == nil {
		*size++
		return &TreeMapNode{key: key, value: value, height: 1}
	}

	if key == n.key {
		n.value = value // Update existing
		return n
	}

	if key < n.key {
		n.left = n.left.insert(key, value, size)
	} else {
		n.right = n.right.insert(key, value, size)
	}

	// On the way back up, every node on the insert path updates its height and rotates if one side got too tall
	return n.rebalance()
}

// nodeHeight - an empty subtree has height 0
func (n *TreeMapNode) nodeHeight() int {

	if n == nil {
		return 0
	}

	return n.height
}

func (n *TreeMapNode) updateHeight() {
	n.height = 1 + max(n.left.nodeHeight(), n.right.nodeHeight())
}

// rebalance - AVL rule: the left and right subtrees of every node differ in height by at most 1
func (n *TreeMapNode) rebalance() *TreeMapNode {

	n.updateHeight()
	balance := n.left.nodeHeight() - n.right.nodeHeight()

	// Left side too tall
	if balance > 1 {

		// Left-Right case: the extra height is under the left child's right side, so rotate the child first
		if n.left.right.nodeHeight() > n.left.left.nodeHeight() {
			n.left = n.left.rotateLeft()
		}

		return n.rotateRight()
	}

	// Right side too tall (mirror of the above)
	if balance < -1 {

		// Right-Left case
		if n.right.left.nodeHeight() > n.right.right.nodeHeight() {
			n.right = n.right.rotateRight()
		}

		return n.rotateLeft()
	}

	return n
}

// rotateRight - the left child becomes the root of this subtree, the sorted order (a < l < b < n < c) is kept
func (n *TreeMapNode) rotateRight() *TreeMapNode {

	//       n            l
	//      / \          / \
	//     l   c   ->   a   n
	//    / \              / \
	//   a   b            b   c
	l := n.left
	n.left = l.right
	l.right = n

	// n is now below l, so its height is updated first
	n.updateHeight()
	l.updateHeight()

	return l
}

// rotateLeft - mirror of rotateRight, the right child becomes the root of this subtree
func (n *TreeMapNode) rotateLeft() *TreeMapNode {

	r := n.right
	n.right = r.left
	r.left = n

	n.updateHeight()
	r.updateHeight()

	return r
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
	key          string
	value        int
	nextInBucket *LinkedMapEntry // For the HashMap (Collision chaining)
	prevInList   *LinkedMapEntry // For the Doubly-Linked List Insertion order
	nextInList   *LinkedMapEntry
}

type LinkedHashMap struct {
	buckets  []*LinkedMapEntry
	capacity int
	size     int
	head     *LinkedMapEntry
	tail     *LinkedMapEntry
}

func NewLinkedHashMap(capacity int) *LinkedHashMap {

	if capacity <= 0 {
		capacity = 16
	}

	return &LinkedHashMap{
		buckets:  make([]*LinkedMapEntry, capacity),
		capacity: capacity,
	}
}

func (m *LinkedHashMap) hash(key string) int {
	return hashKey(key, m.capacity)
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
		key:          key,
		value:        value,
		nextInBucket: m.buckets[index], // Insert at front of bucket list
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

	if float64(m.size)/float64(m.capacity) > maxLoadFactor {
		m.resize()
	}
}

// resize - doubles the buckets and rehashes, the insertion-order list is not touched
func (m *LinkedHashMap) resize() {

	newCapacity := m.capacity * 2
	newBuckets := make([]*LinkedMapEntry, newCapacity)

	// Walking the insertion-order list visits every entry exactly once
	current := m.head
	for current != nil {

		index := hashKey(current.key, newCapacity)

		// Insert at the front of the new bucket
		current.nextInBucket = newBuckets[index]
		newBuckets[index] = current

		current = current.nextInList
	}

	m.buckets = newBuckets
	m.capacity = newCapacity
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
// Growing is the exception: resize replaces the whole bucket array, so it locks the entire map (rarely, each time the size doubles)

type ConcurrentEntry struct {
	key   string
	value int
	next  *ConcurrentEntry
}

// Bucket - wrap a regular bucket in a RWMutex so each bucket can be locked independently
type Bucket struct {
	head *ConcurrentEntry
	lock sync.RWMutex
}

type ConcurrentHashMap struct {
	resizeLock sync.RWMutex // Put/Get hold it shared (many at once, each then locks only its bucket), resize holds it exclusively
	buckets    []*Bucket
	capacity   int
	size       atomic.Int64 // Goroutines holding different bucket locks update it at the same time, so it must be atomic
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
		buckets:  buckets,
		capacity: capacity,
	}
}

func (m *ConcurrentHashMap) hash(key string) int {
	return hashKey(key, m.capacity)
}

func (m *ConcurrentHashMap) Put(key string, value int) {

	// resize needs resizeLock exclusively, so it can only run once put has released its shared lock
	if m.put(key, value) {
		m.resize()
	}
}

// put - inserts or updates, returns true if the map got too full and should grow
func (m *ConcurrentHashMap) put(key string, value int) bool {

	// Shared lock: other Puts/Gets go ahead in parallel, only a resize has to wait
	m.resizeLock.RLock()
	defer m.resizeLock.RUnlock()

	index := m.hash(key)
	bucket := m.buckets[index]

	// Only lock this specific bucket. Other threads can write to other buckets simultaneously
	bucket.lock.Lock()
	defer bucket.lock.Unlock()

	current := bucket.head
	for current != nil {

		if current.key == key {
			current.value = value
			return false
		}

		current = current.next
	}

	newEntry := &ConcurrentEntry{
		key:   key,
		value: value,
		next:  bucket.head,
	}

	bucket.head = newEntry

	size := m.size.Add(1)
	return float64(size)/float64(m.capacity) > maxLoadFactor
}

func (m *ConcurrentHashMap) Get(key string) (int, bool) {

	m.resizeLock.RLock()
	defer m.resizeLock.RUnlock()

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

// resize - doubles the buckets and rehashes every entry
func (m *ConcurrentHashMap) resize() {

	// Exclusive: waits for every Put/Get in progress to finish, and keeps new ones out until the new buckets are in place
	m.resizeLock.Lock()
	defer m.resizeLock.Unlock()

	// Several Puts can cross the limit at the same time, an earlier resize may already have made enough room
	if float64(m.size.Load())/float64(m.capacity) <= maxLoadFactor {
		return
	}

	newCapacity := m.capacity * 2
	newBuckets := make([]*Bucket, newCapacity)
	for i := 0; i < newCapacity; i++ {
		newBuckets[i] = &Bucket{}
	}

	// No bucket locks needed: the exclusive resizeLock already keeps every other goroutine out
	for _, bucket := range m.buckets {

		current := bucket.head
		for current != nil {

			next := current.next
			index := hashKey(current.key, newCapacity)

			current.next = newBuckets[index].head
			newBuckets[index].head = current

			current = next
		}
	}

	m.buckets = newBuckets
	m.capacity = newCapacity
}
