package main

// Regression tests for the bugs found in review: each of these failed (or crashed) before its fix

import (
	"fmt"
	"math"
	"slices"
	"sync"
	"testing"
	"time"
	"unsafe"
)

// bigID is 2^60 on 64-bit platforms (the size of 64-bit IDs), far beyond 2^53 where float64 stops holding every integer
const bigID = math.MaxInt/8 + 1

func expectPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: expected a panic", name)
		}
	}()
	f()
}

// 01_array.go
func TestNewArrayCapacityOverflow(t *testing.T) {
	// The smallest capacity whose size in bytes overflows uintptr (2^61 + 1 on 64-bit): capacity * 8 wrapped
	// around to 8, so the old []byte allocation was 8 bytes long and Set(1, x) passed the bounds check and
	// wrote past it. make([]int, n) detects the overflow and panics
	capacity := int(^uintptr(0)/unsafe.Sizeof(int(0)) + 2)
	expectPanic(t, fmt.Sprintf("NewArray(%d)", capacity), func() { NewArray(capacity) })
}

// 02_map.go, 02_map_extra.go
func TestHashMinIntKey(t *testing.T) {
	// With int arithmetic this key hashes to exactly MinInt64, and -MinInt64 is still MinInt64, so the old
	// hash returned a negative bucket index for every capacity that is not a power of 2
	const key = " 1#('.(<-1 6(="

	var signed int64
	for i := 0; i < len(key); i++ {
		signed = signed*31 + int64(key[i])
	}
	if signed != math.MinInt64 {
		t.Fatalf("test key no longer hashes to MinInt64")
	}

	for _, capacity := range []int{3, 10, 16} {
		hashMap := NewHashMap(capacity)
		linkedMap := NewLinkedHashMap(capacity)
		concurrentMap := NewConcurrentHashMap(capacity)
		hashMap.Put(key, 1)
		linkedMap.Put(key, 2)
		concurrentMap.Put(key, 3)

		if v, ok := hashMap.Get(key); !ok || v != 1 {
			t.Errorf("HashMap(capacity %d): Get = %d, %v", capacity, v, ok)
		}
		if v, ok := linkedMap.Get(key); !ok || v != 2 {
			t.Errorf("LinkedHashMap(capacity %d): Get = %d, %v", capacity, v, ok)
		}
		if v, ok := concurrentMap.Get(key); !ok || v != 3 {
			t.Errorf("ConcurrentHashMap(capacity %d): Get = %d, %v", capacity, v, ok)
		}
	}
}

func TestHashMapsResize(t *testing.T) {
	const n = 10_000
	hashMap := NewHashMap(4)
	linkedMap := NewLinkedHashMap(4)
	for i := range n {
		key := fmt.Sprintf("key-%d", i)
		hashMap.Put(key, i)
		linkedMap.Put(key, i)
	}

	// They used to keep their starting capacity forever: 4 buckets means chains of ~2500 entries
	if load := float64(hashMap.size) / float64(hashMap.capacity); load > maxLoadFactor {
		t.Errorf("HashMap load factor %.2f after %d puts", load, n)
	}
	if load := float64(linkedMap.size) / float64(linkedMap.capacity); load > maxLoadFactor {
		t.Errorf("LinkedHashMap load factor %.2f after %d puts", load, n)
	}

	longest := 0
	for _, entry := range hashMap.buckets {
		chain := 0
		for ; entry != nil; entry = entry.next {
			chain++
		}
		longest = max(longest, chain)
	}
	if longest > 8 {
		t.Errorf("longest HashMap chain has %d entries", longest)
	}

	for i := range n {
		key := fmt.Sprintf("key-%d", i)
		if v, ok := hashMap.Get(key); !ok || v != i {
			t.Fatalf("HashMap lost %s while resizing", key)
		}
		if v, ok := linkedMap.Get(key); !ok || v != i {
			t.Fatalf("LinkedHashMap lost %s while resizing", key)
		}
	}

	// Resizing rebuilds the buckets only, the insertion order survives
	for i, key := range linkedMap.GetKeysInOrder() {
		if key != fmt.Sprintf("key-%d", i) {
			t.Fatalf("insertion order broken at %d: %s", i, key)
		}
	}
}

func TestConcurrentHashMapResizesUnderLoad(t *testing.T) {
	const goroutines, perGoroutine = 8, 2000
	concurrentMap := NewConcurrentHashMap(2) // grows many times while the goroutines write

	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range perGoroutine {
				key := fmt.Sprintf("g%d-%d", g, i)
				concurrentMap.Put(key, i)
				if v, ok := concurrentMap.Get(key); !ok || v != i {
					t.Errorf("read after write %s = %d, %v", key, v, ok)
					return
				}
			}
		}()
	}
	wg.Wait()

	if size := concurrentMap.size.Load(); size != goroutines*perGoroutine {
		t.Fatalf("size = %d, want %d", size, goroutines*perGoroutine)
	}
	if load := float64(concurrentMap.size.Load()) / float64(concurrentMap.capacity); load > maxLoadFactor {
		t.Errorf("load factor %.2f after all puts", load)
	}
	for g := range goroutines {
		for i := range perGoroutine {
			if v, ok := concurrentMap.Get(fmt.Sprintf("g%d-%d", g, i)); !ok || v != i {
				t.Fatalf("lost g%d-%d", g, i)
			}
		}
	}
}

func TestTreeMapStaysBalanced(t *testing.T) {
	const n = 10_000
	treeMap := NewTreeMap()
	for i := range n {
		treeMap.Put(fmt.Sprintf("%06d", i), i) // sorted order turned the plain BST into a linked list
	}

	// An AVL tree is at most ~1.44 * log2(n) tall, the old tree was n tall
	if height, limit := treeMap.root.nodeHeight(), int(1.45*math.Log2(n+2)); height > limit {
		t.Errorf("height %d after %d sorted inserts, AVL limit is %d", height, n, limit)
	}
	if treeMap.size != n {
		t.Errorf("size = %d, want %d", treeMap.size, n)
	}
	if v, ok := treeMap.Get(fmt.Sprintf("%06d", n/2)); !ok || v != n/2 {
		t.Errorf("Get after rotations = %d, %v", v, ok)
	}
}

// 03_search.go
func TestInterpolationSearchLargeValues(t *testing.T) {
	ids := make([]int, 10)
	for i := range ids {
		ids[i] = bigID + i
	}

	cases := []struct {
		name     string
		haystack []int
		absent   []int
	}{
		// Neighbours round to the same float64: the old formula divided by zero
		{"consecutive 64-bit IDs", ids, []int{bigID - 1, bigID + 10}},
		// MaxInt - MinInt overflows int: the old formula computed a negative position
		{"full int range", []int{math.MinInt, -4, -3, -2, -1, 0, 1, 2, 3, 4, math.MaxInt}, []int{-5, 5}},
		// The old formula mixed an exact int difference with rounded float64 ones and computed a position past the end
		{"uneven float64 rounding", []int{bigID + 129, bigID + 150, bigID + 200, bigID + 250, bigID + 300, bigID + 350,
			bigID + 400, bigID + 450, bigID + 500, bigID + 550, bigID + 639}, []int{bigID + 151, bigID + 638}},
	}

	for _, c := range cases {
		for _, needle := range c.haystack {
			if !InterpolationSearch(c.haystack, needle) {
				t.Errorf("%s: %d not found", c.name, needle)
			}
		}
		for _, needle := range c.absent {
			if InterpolationSearch(c.haystack, needle) {
				t.Errorf("%s: absent %d found", c.name, needle)
			}
		}
	}
}

// 04_sort.go
func TestQuickSortDuplicates(t *testing.T) {
	array := make([]int, 1_000_000) // all zeros

	// The old partition peeled off one equal value per level: n^2/2 steps, about 2 minutes here; now well under a second
	done := make(chan struct{})
	go func() {
		QuickSort(array)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("QuickSort of 1M equal values took over 10s, duplicates hit the O(n^2) case")
	}

	if !slices.IsSorted(array) {
		t.Error("not sorted")
	}
}

// 05_linked_lists.go
func TestDeletionLinkedListLastNode(t *testing.T) {
	last := &Node{Value: 1}
	DeletionLinkedList(last) // there is no node after it, this used to dereference nil
	if last.Next != nil {
		t.Errorf("Next = %v, want nil", last.Next)
	}
}

func TestDoublyGetNodeBounds(t *testing.T) {
	list := &DoublyList{}
	expectPanic(t, "GetNode(0) on an empty list", func() { list.GetNode(0) })

	list.InsertTail(1)
	list.InsertTail(2)
	list.InsertTail(3)

	// These used to silently return the head (-1) or the tail (3, 100)
	for _, index := range []int{-1, 3, 100} {
		expectPanic(t, fmt.Sprintf("GetNode(%d)", index), func() { list.GetNode(index) })
	}

	if list.GetNode(0) != list.Head || list.GetNode(1).Value != 2 || list.GetNode(2) != list.Tail {
		t.Errorf("GetNode within bounds returned the wrong nodes")
	}
}

// 08_heap.go
func TestHeapDeleteEmpty(t *testing.T) {
	// Delete used to return 0 on an empty heap, indistinguishable from a stored 0
	expectPanic(t, "MinHeap.Delete on an empty heap", func() { (&MinHeap{}).Delete() })
	expectPanic(t, "MaxHeap.Delete on an empty heap", func() { (&MaxHeap{}).Delete() })

	minHeap := &MinHeap{}
	minHeap.Insert(0)
	if v := minHeap.Delete(); v != 0 {
		t.Errorf("Delete = %d, want the stored 0", v)
	}
	expectPanic(t, "MinHeap.Delete after removing the last value", func() { minHeap.Delete() })
}

// 09_recursion.go
func TestRecursiveSumBaseCase(t *testing.T) {
	// 0 and negative n used to skip the base case and recurse until the stack overflowed
	for n, want := range map[int]int{-3: 0, 0: 0, 1: 1, 5: 15, 100: 5050} {
		if got := RecursiveSum(n); got != want {
			t.Errorf("RecursiveSum(%d) = %d, want %d", n, got, want)
		}
	}
}

func TestMazeSolverRaggedAndEmpty(t *testing.T) {
	if path := MazeSolver([]string{}, '#', Point{0, 0}, Point{0, 0}); path != nil {
		t.Errorf("empty maze: path = %v, want nil", path)
	}

	// Rows of different lengths: the old bounds check used len(maze[0]) for every row and indexed past the short one
	maze := []string{
		"    #",
		"  ",
		"#   ",
	}
	path := MazeSolver(maze, '#', Point{0, 0}, Point{3, 2})
	expected := []Point{{0, 0}, {1, 0}, {1, 1}, {1, 2}, {2, 2}, {3, 2}}
	if !slices.Equal(path, expected) {
		t.Errorf("path = %v, want %v", path, expected)
	}
}

// 12_binary_search_tree.go
func TestBSTDeleteRoot(t *testing.T) {
	tree := &BST{}
	tree.Insert(12)
	tree.Insert(20)

	// Root with one child: the old Delete kept 12 and dropped 20
	if !tree.Delete(12) || tree.Find(12) != nil || tree.Root == nil || tree.Root.Value != 20 || tree.Root.Parent != nil {
		t.Fatalf("after deleting the root 12, root = %+v", tree.Root)
	}

	// Last node: the tree becomes empty (the old Delete could not do that)
	if !tree.Delete(20) || tree.Root != nil || tree.FindMin() != nil {
		t.Fatalf("after deleting the last node, root = %+v", tree.Root)
	}

	if tree.Delete(20) {
		t.Errorf("Delete of a missing value returned true")
	}

	tree.Insert(5)
	if tree.Find(5) == nil || tree.Root.Value != 5 {
		t.Errorf("Insert into an emptied tree failed")
	}
}

// 13_trie_tree.go
func TestTrieDeleteNonASCII(t *testing.T) {
	trie := NewTrieNode()
	trie.Insert("café")
	trie.Insert("日本")

	// Delete walked bytes while Insert walks runes: it returned true and deleted nothing
	if !trie.Delete("café") || trie.Search("café") || trie.StartsWith("c") {
		t.Errorf(`Delete("café") left the word or its nodes behind`)
	}
	if !trie.Delete("日本") || trie.Search("日本") || len(trie.Children) != 0 {
		t.Errorf(`Delete("日本") left the word or its nodes behind`)
	}

	// é is the bytes C3 A9, and as runes those are Ã and ©: deleting "fé" byte by byte removed "fÃ©"
	trie.Insert("fé")
	trie.Insert("fÃ©")
	trie.Delete("fé")
	if trie.Search("fé") || !trie.Search("fÃ©") {
		t.Errorf(`Delete("fé") must remove "fé" and keep "fÃ©"`)
	}
}

// 14_graph.go
func TestDijkstraLargeWeights(t *testing.T) {
	if math.MaxInt == math.MaxInt32 {
		t.Skip("32-bit int: distances can't go past MaxInt32 anyway")
	}

	// Both paths are longer than MaxInt32 (~2.1 billion), the old "infinity", so node 2 looked unreachable
	graph := &WeightedAdjList{List: [][]GraphEdge{
		{{To: 1, Weight: 2_000_000_000}, {To: 3, Weight: 2_000_000_000}},
		{{To: 2, Weight: 2_000_000_000}},
		{},
		{{To: 2, Weight: 1_999_999_999}},
	}}

	if path := graph.Dijkstra(0, 2); !slices.Equal(path, []int{0, 3, 2}) {
		t.Errorf("path = %v, want [0 3 2]", path)
	}
}
