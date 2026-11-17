package main

// Randomized tests: each data structure and algorithm runs thousands of random operations next to a simple
// reference (Go maps and slices, Floyd–Warshall, BFS) and must give the same answers, while its internal
// invariants (links, heap order, BST/AVL shape, trie pruning, load factor) stay intact.
// The seeds are fixed, so a failure always reproduces.

import (
	"fmt"
	"math"
	"math/rand"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestRandomSearch(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for iter := 0; iter < 3000; iter++ {
		haystack := make([]int, r.Intn(40))
		switch r.Intn(3) {
		case 0: // small values, many duplicates
			spread := 1 + r.Intn(50)
			for i := range haystack {
				haystack[i] = r.Intn(2*spread+1) - spread
			}
		case 1: // huge values with small gaps, beyond float64 precision
			for i := range haystack {
				haystack[i] = bigID + r.Intn(600)
			}
		case 2: // anywhere in the int range
			for i := range haystack {
				haystack[i] = int(r.Uint64())
			}
		}
		sort.Ints(haystack)

		// Every value, its neighbours (wrapping around at the int limits is fine) and a few others
		needles := []int{math.MinInt, math.MaxInt, 0, int(r.Uint64())}
		for _, v := range haystack {
			needles = append(needles, v, v-1, v+1)
		}

		for _, needle := range needles {
			want := slices.Contains(haystack, needle)
			got := map[string]bool{
				"LinearSearch":        LinearSearch(haystack, needle),
				"BinarySearch":        BinarySearch(haystack, needle),
				"BinarySearchRange":   BinarySearchRange(haystack, 0, len(haystack), needle),
				"InterpolationSearch": InterpolationSearch(haystack, needle),
				"ExponentialSearch":   ExponentialSearch(haystack, needle),
				"JumpSearch":          JumpSearch(haystack, needle),
			}
			for name, g := range got {
				if g != want {
					t.Fatalf("%s(%v, %d) = %v, want %v", name, haystack, needle, g, want)
				}
			}

			low := r.Intn(len(haystack) + 1)
			high := low + r.Intn(len(haystack)-low+1)
			if g, w := BinarySearchRange(haystack, low, high, needle), slices.Contains(haystack[low:high], needle); g != w {
				t.Fatalf("BinarySearchRange(%v, %d, %d, %d) = %v, want %v", haystack, low, high, needle, g, w)
			}
		}
	}
}

func TestTwoCrystalBallsExhaustive(t *testing.T) {
	for n := 0; n <= 400; n++ {
		for first := 0; first <= n; first++ {
			breaks := make([]bool, n)
			for i := first; i < n; i++ {
				breaks[i] = true
			}
			want := -1
			if first < n {
				want = first
			}
			if got := TwoCrystalBalls(breaks); got != want {
				t.Fatalf("TwoCrystalBalls(n=%d, first break at %d) = %d, want %d", n, first, got, want)
			}
		}
	}
}

func TestRandomSorts(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	sorts := []struct {
		name string
		sort func([]int) []int
	}{
		{"BubbleSort", func(a []int) []int { BubbleSort(a); return a }},
		{"InsertionSort", func(a []int) []int { InsertionSort(a); return a }},
		{"QuickSort", func(a []int) []int { QuickSort(a); return a }},
		{"MergeSort", MergeSort},
		{"HeapSort", func(a []int) []int { HeapSort(a); return a }},
	}

	for iter := 0; iter < 5000; iter++ {
		n := r.Intn(60)
		input := make([]int, n)
		mode := r.Intn(4)
		for i := range input {
			switch mode {
			case 0:
				input[i] = r.Intn(1000) - 500
			case 1:
				input[i] = r.Intn(3) // mostly duplicates
			case 2:
				input[i] = i // already sorted
			case 3:
				input[i] = n - i // reversed
			}
		}
		want := slices.Clone(input)
		sort.Ints(want)

		for _, s := range sorts {
			if got := s.sort(slices.Clone(input)); !slices.Equal(got, want) {
				t.Fatalf("%s(%v) = %v, want %v", s.name, input, got, want)
			}
		}
	}
}

func TestRandomArray(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for iter := 0; iter < 300; iter++ {
		n := 1 + r.Intn(20)
		array := NewArray(n)
		model := make([]int, n)
		for op := 0; op < 200; op++ {
			switch r.Intn(4) {
			case 0:
				i, v := r.Intn(n), r.Intn(10)
				array.Set(i, v)
				model[i] = v
			case 1:
				if r.Intn(10) == 0 {
					array.Clear()
					clear(model)
				}
			case 2:
				v := r.Intn(10)
				if got, want := array.IndexOf(v), slices.Index(model, v); got != want {
					t.Fatalf("IndexOf(%d) = %d, want %d", v, got, want)
				}
			case 3:
				clone := array.Clone()
				snapshot := slices.Clone(model)
				array.Set(0, -1) // must not change the clone
				model[0] = -1
				for i := range snapshot {
					if clone.Get(i) != snapshot[i] {
						t.Fatalf("Clone is wrong or shares memory at %d", i)
					}
				}
			}
			if array.Len() != n {
				t.Fatalf("Len = %d, want %d", array.Len(), n)
			}
			for i := range model {
				if array.Get(i) != model[i] {
					t.Fatalf("Get(%d) = %d, want %d", i, array.Get(i), model[i])
				}
			}
		}
	}
}

func TestRandomDynamicArray(t *testing.T) {
	r := rand.New(rand.NewSource(4))
	for iter := 0; iter < 300; iter++ {
		array := NewDynamicArray()
		var model []int
		for op := 0; op < 300; op++ {
			v := r.Intn(1000)
			switch r.Intn(6) {
			case 0:
				array.Push(v)
				model = append(model, v)
			case 1:
				if len(model) > 0 {
					if got, want := array.Pop(), model[len(model)-1]; got != want {
						t.Fatalf("Pop = %d, want %d", got, want)
					}
					model = model[:len(model)-1]
				}
			case 2:
				array.Prepend(v)
				model = slices.Insert(model, 0, v)
			case 3:
				if len(model) > 0 {
					if got, want := array.Shift(), model[0]; got != want {
						t.Fatalf("Shift = %d, want %d", got, want)
					}
					model = slices.Clone(model[1:])
				}
			case 4:
				i := r.Intn(len(model) + 1)
				array.InsertAt(i, v)
				model = slices.Insert(model, i, v)
			case 5:
				if len(model) > 0 {
					i := r.Intn(len(model))
					array.RemoveAt(i)
					model = slices.Delete(model, i, i+1)
				}
			}
			if array.Len() != len(model) {
				t.Fatalf("Len = %d, want %d", array.Len(), len(model))
			}
			for i := range model {
				if array.Get(i) != model[i] {
					t.Fatalf("Get(%d) = %d, want %d", i, array.Get(i), model[i])
				}
			}
		}
	}
}

func checkSinglyList(t *testing.T, list *SinglyList, model []int) {
	t.Helper()
	if list.Length != len(model) {
		t.Fatalf("Length = %d, want %d", list.Length, len(model))
	}
	curr := list.Head
	for _, v := range model {
		if curr == nil || curr.Value != v {
			t.Fatalf("list differs from %v", model)
		}
		curr = curr.Next
	}
	if curr != nil {
		t.Fatalf("extra nodes after %v", model)
	}
	for i, v := range model {
		if list.Get(i) != v {
			t.Fatalf("Get(%d) = %d, want %d", i, list.Get(i), v)
		}
	}
}

func TestRandomSinglyList(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	for iter := 0; iter < 300; iter++ {
		list := &SinglyList{}
		var model []int
		for op := 0; op < 200; op++ {
			v := r.Intn(1000)
			switch r.Intn(5) {
			case 0:
				i := r.Intn(len(model) + 1)
				list.InsertAt(i, v)
				model = slices.Insert(model, i, v)
			case 1:
				if len(model) > 0 {
					list.Remove()
					model = model[:len(model)-1]
				}
			case 2:
				if len(model) > 0 {
					i := r.Intn(len(model))
					list.RemoveAt(i)
					model = slices.Delete(model, i, i+1)
				}
			case 3:
				list.Append(v)
				model = append(model, v)
			case 4:
				list.Prepend(v)
				model = slices.Insert(model, 0, v)
			}
			checkSinglyList(t, list, model)
		}
	}
}

func checkDoublyList(t *testing.T, list *DoublyList, model []int) {
	t.Helper()
	if list.Length != len(model) {
		t.Fatalf("Length = %d, want %d", list.Length, len(model))
	}
	if len(model) == 0 {
		if list.Head != nil || list.Tail != nil {
			t.Fatalf("empty list with a Head or Tail")
		}
		return
	}
	if list.Head.Prev != nil || list.Tail.Next != nil {
		t.Fatalf("Head.Prev or Tail.Next is not nil")
	}
	curr := list.Head
	var prev *Node
	for _, v := range model {
		if curr == nil || curr.Value != v || curr.Prev != prev {
			t.Fatalf("Next/Prev links differ from %v", model)
		}
		prev, curr = curr, curr.Next
	}
	if curr != nil || prev != list.Tail {
		t.Fatalf("Tail is not the last node")
	}
	for i, v := range model {
		if list.Get(i) != v || list.GetNode(i).Value != v {
			t.Fatalf("Get/GetNode(%d) differs from %d", i, v)
		}
	}
}

func TestRandomDoublyList(t *testing.T) {
	r := rand.New(rand.NewSource(6))
	for iter := 0; iter < 300; iter++ {
		list := &DoublyList{}
		var model []int
		for op := 0; op < 200; op++ {
			v := r.Intn(1000)
			switch r.Intn(6) {
			case 0:
				i := r.Intn(len(model) + 1)
				list.InsertAt(i, v)
				model = slices.Insert(model, i, v)
			case 1:
				if len(model) > 0 {
					i := r.Intn(len(model))
					list.RemoveAt(i)
					model = slices.Delete(model, i, i+1)
				}
			case 2:
				list.InsertHead(v)
				model = slices.Insert(model, 0, v)
			case 3:
				list.InsertTail(v)
				model = append(model, v)
			case 4:
				if len(model) > 0 {
					list.RemoveHead()
					model = slices.Clone(model[1:])
				}
			case 5:
				if len(model) > 0 {
					list.RemoveTail()
					model = model[:len(model)-1]
				}
			}
			checkDoublyList(t, list, model)
		}
	}
}

func TestRandomQueueStackRingBuffer(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	queue, stack, ring := &Queue{}, &Stack{}, NewRingBuffer(5)
	var queueModel, stackModel, ringModel []int
	for op := 0; op < 200_000; op++ {
		v := r.Int()

		if r.Intn(2) == 0 {
			queue.Enqueue(v)
			queueModel = append(queueModel, v)
		} else if len(queueModel) > 0 {
			if queue.Peek() != queueModel[0] || queue.Deque() != queueModel[0] {
				t.Fatalf("Queue Peek/Deque differ from %d", queueModel[0])
			}
			queueModel = queueModel[1:]
		}
		if queue.Length != len(queueModel) ||
			(len(queueModel) > 0 && queue.Tail.Value != queueModel[len(queueModel)-1]) ||
			(len(queueModel) == 0 && (queue.Head != nil || queue.Tail != nil)) {
			t.Fatalf("Queue state differs from the model")
		}

		if r.Intn(2) == 0 {
			stack.Push(v)
			stackModel = append(stackModel, v)
		} else if len(stackModel) > 0 {
			top := stackModel[len(stackModel)-1]
			if stack.Peek() != top || stack.Pop() != top {
				t.Fatalf("Stack Peek/Pop differ from %d", top)
			}
			stackModel = stackModel[:len(stackModel)-1]
		}
		if stack.Length != len(stackModel) {
			t.Fatalf("Stack Length = %d, want %d", stack.Length, len(stackModel))
		}

		if r.Intn(2) == 0 {
			if len(ringModel) < 5 {
				ring.Enqueue(v)
				ringModel = append(ringModel, v)
			}
		} else if len(ringModel) > 0 {
			if ring.Peek() != ringModel[0] || ring.Dequeue() != ringModel[0] {
				t.Fatalf("RingBuffer Peek/Dequeue differ from %d", ringModel[0])
			}
			ringModel = ringModel[1:]
		}
		if ring.length != len(ringModel) {
			t.Fatalf("RingBuffer length = %d, want %d", ring.length, len(ringModel))
		}
	}
}

var testKeys = []string{"", "a", "b", "ab", "ba", "abc", "Aa", "BB", "AaAa", "BBBB", "AaBB", "BBAa"} // Aa and BB hash the same

func TestRandomHashMap(t *testing.T) {
	r := rand.New(rand.NewSource(8))
	for iter := 0; iter < 300; iter++ {
		hashMap := NewHashMap(r.Intn(9) - 1)
		values := map[string]int{}
		for op := 0; op < 300; op++ {
			key := testKeys[r.Intn(len(testKeys))] + strings.Repeat("x", r.Intn(3))
			switch r.Intn(5) {
			case 0, 1:
				v := r.Intn(100)
				hashMap.Put(key, v)
				values[key] = v
			case 2:
				hashMap.Delete(key)
				delete(values, key)
			case 3:
				want, wantOK := values[key]
				if got, ok := hashMap.Get(key); ok != wantOK || got != want || hashMap.Has(key) != wantOK {
					t.Fatalf("Get(%q) = %d, %v, want %d, %v", key, got, ok, want, wantOK)
				}
			case 4:
				if r.Intn(20) == 0 {
					hashMap.Clear()
					clear(values)
				}
			}
			if hashMap.Size() != len(values) {
				t.Fatalf("Size = %d, want %d", hashMap.Size(), len(values))
			}
			if load := float64(hashMap.size) / float64(hashMap.capacity); load > maxLoadFactor {
				t.Fatalf("load factor %.2f", load)
			}
		}
		for key, want := range values {
			if got, ok := hashMap.Get(key); !ok || got != want {
				t.Fatalf("final Get(%q) = %d, %v, want %d", key, got, ok, want)
			}
		}
	}
}

// checkAVL verifies the key order, the stored heights and the AVL balance, and returns the subtree height
func checkAVL(t *testing.T, node *TreeMapNode, low, high *string) int {
	t.Helper()
	if node == nil {
		return 0
	}
	if (low != nil && node.key <= *low) || (high != nil && node.key >= *high) {
		t.Fatalf("key order broken at %q", node.key)
	}
	left := checkAVL(t, node.left, low, &node.key)
	right := checkAVL(t, node.right, &node.key, high)
	if node.height != 1+max(left, right) {
		t.Fatalf("stale height at %q: %d, want %d", node.key, node.height, 1+max(left, right))
	}
	if left-right > 1 || right-left > 1 {
		t.Fatalf("unbalanced at %q: left height %d, right height %d", node.key, left, right)
	}
	return node.height
}

func TestRandomTreeMapLinkedHashMap(t *testing.T) {
	r := rand.New(rand.NewSource(9))
	for iter := 0; iter < 300; iter++ {
		treeMap := NewTreeMap()
		linkedMap := NewLinkedHashMap(r.Intn(6))
		values := map[string]int{}
		var order []string // insertion order
		for op := 0; op < 200; op++ {
			var key string
			switch r.Intn(3) {
			case 0:
				key = testKeys[r.Intn(len(testKeys))] + strings.Repeat("y", r.Intn(3))
			case 1:
				key = fmt.Sprintf("up%03d", op) // increasing: the worst case for a plain BST
			case 2:
				key = fmt.Sprintf("down%03d", 999-op) // decreasing
			}

			if r.Intn(3) != 0 {
				v := r.Intn(100)
				treeMap.Put(key, v)
				linkedMap.Put(key, v)
				if _, exists := values[key]; !exists {
					order = append(order, key)
				}
				values[key] = v
			} else {
				want, wantOK := values[key]
				if got, ok := treeMap.Get(key); ok != wantOK || got != want {
					t.Fatalf("TreeMap.Get(%q) = %d, %v, want %d, %v", key, got, ok, want, wantOK)
				}
				if got, ok := linkedMap.Get(key); ok != wantOK || got != want {
					t.Fatalf("LinkedHashMap.Get(%q) = %d, %v, want %d, %v", key, got, ok, want, wantOK)
				}
			}

			sorted := make([]string, 0, len(values))
			for k := range values {
				sorted = append(sorted, k)
			}
			sort.Strings(sorted)
			if got := treeMap.InOrderKeys(); !slices.Equal(got, sorted) || treeMap.size != len(values) {
				t.Fatalf("TreeMap.InOrderKeys = %q, want %q", got, sorted)
			}
			checkAVL(t, treeMap.root, nil, nil)

			if got := linkedMap.GetKeysInOrder(); !slices.Equal(got, order) || linkedMap.size != len(values) {
				t.Fatalf("LinkedHashMap.GetKeysInOrder = %q, want %q", got, order)
			}
			if load := float64(linkedMap.size) / float64(linkedMap.capacity); load > maxLoadFactor {
				t.Fatalf("LinkedHashMap load factor %.2f", load)
			}
		}
	}
}

func TestRandomHeaps(t *testing.T) {
	r := rand.New(rand.NewSource(10))
	heapOrdered := func(array []int, isMin bool) bool {
		for i := 1; i < len(array); i++ {
			parent := (i - 1) / 2
			if (isMin && array[i] < array[parent]) || (!isMin && array[i] > array[parent]) {
				return false
			}
		}
		return true
	}

	for iter := 0; iter < 300; iter++ {
		minHeap, maxHeap := &MinHeap{}, &MaxHeap{}
		var minValues, maxValues []int // references, kept sorted
		for op := 0; op < 300; op++ {
			if len(minValues) == 0 || r.Intn(3) != 0 {
				v := r.Intn(50)
				minHeap.Insert(v)
				minValues = slices.Insert(minValues, sort.SearchInts(minValues, v), v)
			} else {
				if got := minHeap.Delete(); got != minValues[0] {
					t.Fatalf("MinHeap.Delete = %d, want %d", got, minValues[0])
				}
				minValues = minValues[1:]
			}

			if len(maxValues) == 0 || r.Intn(3) != 0 {
				v := r.Intn(50)
				maxHeap.Insert(v)
				maxValues = slices.Insert(maxValues, sort.SearchInts(maxValues, v), v)
			} else {
				if got := maxHeap.Delete(); got != maxValues[len(maxValues)-1] {
					t.Fatalf("MaxHeap.Delete = %d, want %d", got, maxValues[len(maxValues)-1])
				}
				maxValues = maxValues[:len(maxValues)-1]
			}

			if !heapOrdered(minHeap.array, true) || len(minHeap.array) != len(minValues) {
				t.Fatalf("MinHeap order or size broken")
			}
			if !heapOrdered(maxHeap.array, false) || len(maxHeap.array) != len(maxValues) {
				t.Fatalf("MaxHeap order or size broken")
			}
		}
	}
}

// checkBST verifies the order and the Parent pointers, and returns the number of nodes
func checkBST(t *testing.T, node, parent *BSTNode, low, high int) int {
	t.Helper()
	if node == nil {
		return 0
	}
	if node.Parent != parent {
		t.Fatalf("wrong Parent pointer at %d", node.Value)
	}
	if node.Value <= low || node.Value >= high {
		t.Fatalf("BST order broken at %d (must be between %d and %d)", node.Value, low, high)
	}
	return 1 + checkBST(t, node.Left, node, low, node.Value) + checkBST(t, node.Right, node, node.Value, high)
}

func TestRandomBST(t *testing.T) {
	r := rand.New(rand.NewSource(11))
	for iter := 0; iter < 300; iter++ {
		tree := &BST{}
		values := map[int]bool{}
		for op := 0; op < 150; op++ {
			v := r.Intn(40)
			if r.Intn(2) == 0 {
				tree.Insert(v)
				values[v] = true
			} else {
				// Any value, including the root, and down to an empty tree
				if got := tree.Delete(v); got != values[v] {
					t.Fatalf("Delete(%d) = %v, want %v", v, got, values[v])
				}
				delete(values, v)
			}

			if count := checkBST(t, tree.Root, nil, math.MinInt, math.MaxInt); count != len(values) {
				t.Fatalf("tree has %d nodes, want %d", count, len(values))
			}
			for k := range 40 {
				if (tree.Find(k) != nil) != values[k] {
					t.Fatalf("Find(%d) is wrong", k)
				}
			}
			if len(values) == 0 {
				if tree.Root != nil || tree.FindMin() != nil {
					t.Fatalf("empty tree still has a root")
				}
			} else {
				smallest := math.MaxInt
				for k := range values {
					smallest = min(smallest, k)
				}
				if tree.FindMin().Value != smallest {
					t.Fatalf("FindMin = %d, want %d", tree.FindMin().Value, smallest)
				}
			}
		}
	}
}

func TestRandomTrie(t *testing.T) {
	r := rand.New(rand.NewSource(12))
	alphabet := []rune("aé日") // 1, 2 and 3 bytes in UTF-8

	randomWord := func(length int) string {
		word := make([]rune, length)
		for i := range word {
			word[i] = alphabet[r.Intn(len(alphabet))]
		}
		return string(word)
	}

	// Every non-empty prefix up to 3 runes: StartsWith must be false once no word uses it (Delete pruned it)
	var prefixes []string
	var addPrefixes func(prefix string)
	addPrefixes = func(prefix string) {
		if prefix != "" {
			prefixes = append(prefixes, prefix)
		}
		if len([]rune(prefix)) < 3 {
			for _, c := range alphabet {
				addPrefixes(prefix + string(c))
			}
		}
	}
	addPrefixes("")

	for iter := 0; iter < 200; iter++ {
		trie := NewTrieNode()
		words := map[string]bool{}
		for op := 0; op < 150; op++ {
			word := randomWord(r.Intn(5))
			switch r.Intn(3) {
			case 0:
				trie.Insert(word)
				words[word] = true
			case 1:
				if got := trie.Delete(word); got != words[word] {
					t.Fatalf("Delete(%q) = %v, want %v", word, got, words[word])
				}
				delete(words, word)
			case 2:
				if got := trie.Search(word); got != words[word] {
					t.Fatalf("Search(%q) = %v, want %v", word, got, words[word])
				}
			}

			for _, prefix := range prefixes {
				want := false
				for w := range words {
					if strings.HasPrefix(w, prefix) {
						want = true
						break
					}
				}
				if got := trie.StartsWith(prefix); got != want {
					t.Fatalf("StartsWith(%q) = %v, want %v (words %v)", prefix, got, want, words)
				}
			}
		}

		for w := range words {
			trie.Delete(w)
		}
		if len(trie.Children) != 0 || trie.IsWord {
			t.Fatalf("nodes left behind after deleting every word")
		}
	}
}

func TestRandomGraphs(t *testing.T) {
	r := rand.New(rand.NewSource(13))
	const inf = math.MaxInt / 4 // reference "infinity": above any real distance below, and safe to add twice

	for iter := 0; iter < 3000; iter++ {
		n := 1 + r.Intn(8)
		maxWeight := 10
		if r.Intn(4) == 0 {
			maxWeight = math.MaxInt / 64 // far above MaxInt32 on 64-bit, yet a 7-edge path stays below inf
		}

		matrix := make([][]int, n)     // BFS input: weight + 1, so every edge is > 0
		list := make([][]GraphEdge, n) // DFS and Dijkstra input: weight 0 is a valid edge here
		weight := make([][]int, n)     // reference, -1 means no edge
		for u := range n {
			matrix[u] = make([]int, n)
			weight[u] = make([]int, n)
			for v := range n {
				weight[u][v] = -1
				if r.Intn(3) == 0 {
					w := r.Intn(maxWeight)
					weight[u][v] = w
					list[u] = append(list[u], GraphEdge{To: v, Weight: w})
					matrix[u][v] = w + 1
				}
			}
		}

		// Floyd–Warshall: shortest distances and fewest edges between every pair
		dist := make([][]int, n)
		hops := make([][]int, n)
		for u := range n {
			dist[u] = make([]int, n)
			hops[u] = make([]int, n)
			for v := range n {
				dist[u][v], hops[u][v] = inf, inf
				if weight[u][v] >= 0 {
					dist[u][v], hops[u][v] = weight[u][v], 1
				}
			}
			dist[u][u], hops[u][u] = 0, 0
		}
		for k := range n {
			for i := range n {
				for j := range n {
					dist[i][j] = min(dist[i][j], dist[i][k]+dist[k][j])
					hops[i][j] = min(hops[i][j], hops[i][k]+hops[k][j])
				}
			}
		}

		// pathCost returns the total weight, and false if the path is broken or visits a node twice
		pathCost := func(path []int, source, needle int) (int, bool) {
			if len(path) == 0 || path[0] != source || path[len(path)-1] != needle {
				return 0, false
			}
			cost := 0
			visited := map[int]bool{}
			for i, node := range path {
				if visited[node] {
					return 0, false
				}
				visited[node] = true
				if i > 0 {
					if weight[path[i-1]][node] < 0 {
						return 0, false
					}
					cost += weight[path[i-1]][node]
				}
			}
			return cost, true
		}

		for source := range n {
			for needle := range n {
				bfs := (&WeightedAdjMatrix{Matrix: matrix}).BFS(source, needle)
				dfs := (&WeightedAdjList{List: list}).DFS(source, needle)
				dijkstra := (&WeightedAdjList{List: list}).Dijkstra(source, needle)

				if dist[source][needle] == inf {
					if bfs != nil || dfs != nil || dijkstra != nil {
						t.Fatalf("%d->%d is unreachable, got bfs=%v dfs=%v dijkstra=%v", source, needle, bfs, dfs, dijkstra)
					}
					continue
				}
				if _, ok := pathCost(bfs, source, needle); !ok || len(bfs)-1 != hops[source][needle] {
					t.Fatalf("BFS %d->%d = %v, want a path of %d edges", source, needle, bfs, hops[source][needle])
				}
				if _, ok := pathCost(dfs, source, needle); !ok {
					t.Fatalf("DFS %d->%d = %v is not a valid path", source, needle, dfs)
				}
				if cost, ok := pathCost(dijkstra, source, needle); !ok || cost != dist[source][needle] {
					t.Fatalf("Dijkstra %d->%d = %v (cost %d), want cost %d", source, needle, dijkstra, cost, dist[source][needle])
				}
			}
		}
	}
}

func TestRandomLRU(t *testing.T) {
	r := rand.New(rand.NewSource(14))
	for iter := 0; iter < 500; iter++ {
		capacity := 1 + r.Intn(4)
		cache := NewLRUCache(capacity)
		values := map[string]int{}
		var recency []string // most recently used first
		touch := func(key string) {
			if i := slices.Index(recency, key); i >= 0 {
				recency = slices.Delete(recency, i, i+1)
			}
			recency = slices.Insert(recency, 0, key)
		}

		for op := 0; op < 200; op++ {
			key := string(rune('a' + r.Intn(6)))
			if r.Intn(2) == 0 {
				v := r.Intn(100)
				cache.Put(key, v)
				values[key] = v
				touch(key)
				if len(recency) > capacity {
					delete(values, recency[len(recency)-1])
					recency = recency[:len(recency)-1]
				}
			} else {
				want, wantOK := values[key]
				if got, ok := cache.Get(key); ok != wantOK || (ok && got != want) {
					t.Fatalf("Get(%q) = %d, %v, want %d, %v", key, got, ok, want, wantOK)
				}
				if wantOK {
					touch(key)
				}
			}

			var got []string
			for node := cache.head.next; node != cache.tail; node = node.next {
				if node.next.prev != node {
					t.Fatalf("broken prev link after %q", node.key)
				}
				got = append(got, node.key)
			}
			if !slices.Equal(got, recency) || len(cache.cache) != len(recency) {
				t.Fatalf("recency order %v, want %v", got, recency)
			}
		}
	}
}

func TestRandomMaze(t *testing.T) {
	r := rand.New(rand.NewSource(15))
	directions := []Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	for iter := 0; iter < 5000; iter++ {
		// Rows of random lengths (even empty), and sometimes no rows at all
		maze := make([]string, r.Intn(7))
		for y := range maze {
			row := make([]byte, r.Intn(7))
			for x := range row {
				row[x] = ' '
				if r.Intn(3) == 0 {
					row[x] = '#'
				}
			}
			maze[y] = string(row)
		}
		// Start and end can be off the map too
		start := Point{r.Intn(7), r.Intn(len(maze) + 1)}
		end := Point{r.Intn(7), r.Intn(len(maze) + 1)}

		open := func(p Point) bool {
			return p.y >= 0 && p.y < len(maze) && p.x >= 0 && p.x < len(maze[p.y]) && maze[p.y][p.x] != '#'
		}

		// Reference: BFS reachability
		reachable := false
		if open(start) {
			seen := map[Point]bool{start: true}
			queue := []Point{start}
			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]
				if curr == end {
					reachable = true
					break
				}
				for _, d := range directions {
					next := Point{curr.x + d.x, curr.y + d.y}
					if open(next) && !seen[next] {
						seen[next] = true
						queue = append(queue, next)
					}
				}
			}
		}

		path := MazeSolver(maze, '#', start, end)
		if reachable != (path != nil) {
			t.Fatalf("maze %q %v->%v: reachable = %v, but got path %v", maze, start, end, reachable, path)
		}
		if path == nil {
			continue
		}
		if path[0] != start || path[len(path)-1] != end {
			t.Fatalf("path %v does not go from %v to %v", path, start, end)
		}
		visited := map[Point]bool{}
		for i, p := range path {
			if !open(p) || visited[p] {
				t.Fatalf("path %v steps on a wall, off the map, or twice on %v", path, p)
			}
			visited[p] = true
			if i > 0 {
				dx, dy := p.x-path[i-1].x, p.y-path[i-1].y
				if dx*dx+dy*dy != 1 {
					t.Fatalf("path %v jumps between %v and %v", path, path[i-1], p)
				}
			}
		}
	}
}

func TestRandomTrees(t *testing.T) {
	r := rand.New(rand.NewSource(16))
	for iter := 0; iter < 500; iter++ {
		size := 1 + r.Intn(20)

		// General tree: each node goes under a random earlier node
		nodes := []*TreeNode{{Value: 0}}
		for i := 1; i < size; i++ {
			node := &TreeNode{Value: i}
			nodes[r.Intn(len(nodes))].Insert(node)
			nodes = append(nodes, node)
		}
		var pre, post []int
		var walkTree func(node *TreeNode)
		walkTree = func(node *TreeNode) {
			pre = append(pre, node.Value)
			for _, child := range node.Children {
				walkTree(child)
			}
			post = append(post, node.Value)
		}
		walkTree(nodes[0])
		if !slices.Equal(nodes[0].PreOrderSearch(), pre) || !slices.Equal(nodes[0].PostOrderSearch(), post) {
			t.Fatalf("general tree traversals are wrong")
		}
		if size > 1 {
			victim := nodes[1+r.Intn(size-1)]
			parent := victim.Parent
			victim.Delete()
			if victim.Parent != nil || slices.Contains(parent.Children, victim) {
				t.Fatalf("TreeNode.Delete did not detach the node")
			}
		}

		// Binary tree: each node goes into a free slot of a random earlier node
		binaryNodes := []*BinaryTreeNode{{Value: 0}}
		for i := 1; i < size; i++ {
			var free []*BinaryTreeNode
			for _, node := range binaryNodes {
				if node.Left == nil || node.Right == nil {
					free = append(free, node)
				}
			}
			node := &BinaryTreeNode{Value: i * 3}
			free[r.Intn(len(free))].Insert(node)
			binaryNodes = append(binaryNodes, node)
		}
		root := binaryNodes[0]
		var binaryPre, binaryIn, binaryPost []int
		var walkBinary func(node *BinaryTreeNode)
		walkBinary = func(node *BinaryTreeNode) {
			if node == nil {
				return
			}
			binaryPre = append(binaryPre, node.Value)
			walkBinary(node.Left)
			binaryIn = append(binaryIn, node.Value)
			walkBinary(node.Right)
			binaryPost = append(binaryPost, node.Value)
		}
		walkBinary(root)
		if !slices.Equal(root.PreOrderSearch(), binaryPre) || !slices.Equal(root.InOrderSearch(), binaryIn) ||
			!slices.Equal(root.PostOrderSearch(), binaryPost) {
			t.Fatalf("binary tree traversals are wrong")
		}
		for needle := -1; needle <= 3*size; needle++ {
			if root.BreadthFirstSearch(needle) != slices.Contains(binaryPre, needle) {
				t.Fatalf("BreadthFirstSearch(%d) is wrong", needle)
			}
		}

		var clone func(node *BinaryTreeNode) *BinaryTreeNode
		clone = func(node *BinaryTreeNode) *BinaryTreeNode {
			if node == nil {
				return nil
			}
			return &BinaryTreeNode{Value: node.Value, Left: clone(node.Left), Right: clone(node.Right)}
		}
		copied := clone(root)
		if !CompareBinaryTree(root, copied) {
			t.Fatalf("CompareBinaryTree(tree, copy) = false")
		}
		copied.Value++
		if CompareBinaryTree(root, copied) {
			t.Fatalf("CompareBinaryTree missed a different value")
		}
		copied.Value--
		if copied.Left != nil {
			copied.Left = nil
			if CompareBinaryTree(root, copied) {
				t.Fatalf("CompareBinaryTree missed a different shape")
			}
		}
	}
}
