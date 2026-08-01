package main

// LRU Cache (Least Recently Used)
// Evicts the least recently used item when the cache reaches its capacity
// It uses a Hash Map (for O(1) lookups) combined with a Doubly-Linked List (for O(1) removals and moves)
// When an item is accessed (Get) or added (Put), it is moved to the front of the list (Most Recently Used)
// When the cache exceeds capacity, the item at the back of the list (Least Recently Used) is removed

type LRUNode struct {
	key		string
	value	int
	prev	*LRUNode
	next	*LRUNode
}

type LRUCache struct {
	capacity	int
	cache		map[string]*LRUNode
	head		*LRUNode // head and tail for easy access
	tail		*LRUNode
}

func NewLRUCache(capacity int) *LRUCache {

	lru := &LRUCache{
		capacity:	capacity,
		cache:		make(map[string]*LRUNode),
		head:		&LRUNode{},
		tail:		&LRUNode{},
	}
	
	// Connect dummy head and tail
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
	
	return lru
}

// removeNode from the doubly-linked list
func (cache *LRUCache) removeNode(node *LRUNode) {

	node.prev.next = node.next
	node.next.prev = node.prev
}

// insertHead inserts node after the head (Most Recently Used)
func (cache *LRUCache) insertHead(node *LRUNode) {

	node.next = cache.head.next
	node.prev = cache.head
	cache.head.next.prev = node
	cache.head.next = node
}

// Get returns the value for the key if it exists, otherwise returns -1 and false
// It also moves the accessed node to the front of the list (Most Recenlty Used)
func (cache *LRUCache) Get(key string) (int, bool) {

	if node, exists := cache.cache[key]; exists {

		// Move to Most Recently Used (front)
		cache.removeNode(node)
		cache.insertHead(node)
		return node.value, true
	}

	return -1, false
}

// Put inserts or updates a key-value pair
// If the cache reaches capacity, it evicts the least recently used item
func (cache *LRUCache) Put(key string, value int) {

	if node, exists := cache.cache[key]; exists {

		// Update value and move to front
		node.value = value
		cache.removeNode(node)
		cache.insertHead(node)
	
	} else {

		// Create new node
		newNode := &LRUNode{
			key:	key,
			value:	value,
		}

		cache.cache[key] = newNode
		cache.insertHead(newNode)
		
		// Evict if over capacity
		if len(cache.cache) > cache.capacity {

			// Least Recently Used is before tail
			lru := cache.tail.prev
			cache.removeNode(lru)
			delete(cache.cache, lru.key)
		}
	}
}
