package main

// Here are just checks that the algorithms are working

func main() {

	// Array
	
	customArray := NewArray(3)
	customArray.Set(0, 1)
	customArray.Set(1, 2)
	customArray.Set(2, 3)
	clonedCustomArray := customArray.Clone()

	if customArray.Get(0) != clonedCustomArray.Get(0) || customArray.Get(1) != clonedCustomArray.Get(1) || customArray.Get(2) != clonedCustomArray.Get(2) ||
		customArray.Get(0) != 1 || customArray.Get(1) != 2 || customArray.Get(2) != 3 {
		panic("array")
	}

	customArray.Clear()

	if customArray.Len() != 3 || customArray.Get(0) != 0 || customArray.Get(1) != 0 || customArray.Get(2) != 0 {
		panic("array")
	}

	dynamicArray := NewDynamicArray()
	dynamicArray.Push(30)
	dynamicArray.Prepend(10)
	dynamicArray.InsertAt(1, 20)

	if dynamicArray.Len() != 3 || dynamicArray.Get(0) != 10 || dynamicArray.Get(1) != 20 || dynamicArray.Get(2) != 30 {
		panic("dynamic array")
	}

	dynamicArray.RemoveAt(1)
	dynamicArray.Pop()
	dynamicArray.Shift()

	if dynamicArray.Len() != 0 {
		panic("dynamic array")
	}

	ringBuffer := NewRingBuffer(3)
	ringBuffer.Enqueue(10)
	ringBuffer.Enqueue(20)
	ringBufferValue1 := ringBuffer.Peek()
	ringBuffer.Dequeue()
	ringBufferValue2 := ringBuffer.Peek()
	ringBuffer.Dequeue()

	if ringBufferValue1 != 10 || ringBufferValue2 != 20 {
		panic("ring buffer")
	}

	// Map

	hashMap := NewHashMap(4)
	hashMap.Put("apple", 1)
	hashMap.Put("banana", 2)
	hashMap.Put("cherry", 3)
	
	if val, ok := hashMap.Get("banana"); !ok || val != 2 {
		panic("map get")
	}

	hashMap.Put("apple", 10) // Update
	if val, ok := hashMap.Get("apple"); !ok || val != 10 {
		panic("map update")
	}

	if !hashMap.Has("cherry") || hashMap.Size() != 3 {
		panic("map has/size")
	}

	hashMap.Delete("banana")
	if hashMap.Has("banana") || hashMap.Size() != 2 {
		panic("map delete")
	}

	hashMap.Clear()
	if hashMap.Size() != 0 || hashMap.Has("apple") {
		panic("map clear")
	}

	// TreeMap (Ordered)

	treeMap := NewTreeMap()
	treeMap.Put("c", 3)
	treeMap.Put("a", 1)
	treeMap.Put("b", 2)
	
	orderedKeys := treeMap.InOrderKeys()
	if len(orderedKeys) != 3 || orderedKeys[0] != "a" || orderedKeys[1] != "b" || orderedKeys[2] != "c" {
		panic("tree map order")
	}

	if val, ok := treeMap.Get("b"); !ok || val != 2 {
		panic("tree map get")
	}

	// LinkedHashMap (Insertion-Ordered)

	linkedMap := NewLinkedHashMap(4)
	linkedMap.Put("c", 3)
	linkedMap.Put("a", 1)
	linkedMap.Put("b", 2)
	
	insertionKeys := linkedMap.GetKeysInOrder()
	if len(insertionKeys) != 3 || insertionKeys[0] != "c" || insertionKeys[1] != "a" || insertionKeys[2] != "b" {
		panic("linked map order")
	}

	if val, ok := linkedMap.Get("a"); !ok || val != 1 {
		panic("linked map get")
	}

	// ConcurrentHashMap (Thread-Safe)

	concurrentMap := NewConcurrentHashMap(16)
	concurrentMap.Put("safe", 99)
	if val, ok := concurrentMap.Get("safe"); !ok || val != 99 {
		panic("concurrent map get")
	}

	// Search

	array := []int{21, 49, 42, 12, 37, 8, 5, 2, 19, 33}
	primes := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}

	if !LinearSearch(array, 5) {
		panic("linear search")
	}
	
	if !BinarySearch(primes, 5) || !BinarySearch(primes, 29) {
		panic("binary search")
	}

	if !BinarySearchRange(primes, 0, len(primes), 5) || !BinarySearchRange(primes, 0, len(primes), 29) {
		panic("binary search range")
	}

	if !InterpolationSearch(primes, 5) || !InterpolationSearch(primes, 29) {
		panic("interpolation search")
	}

	if !ExponentialSearch(primes, 5) || !ExponentialSearch(primes, 29) {
		panic("exponential search")
	}

	if !JumpSearch(primes, 5) || !JumpSearch(primes, 29) {
		panic("jump search")
	}
	
	twoCrystalBalls := []bool{false, false, false, true, true}
	if TwoCrystalBalls(twoCrystalBalls) != 3 {
		panic("two crystal balls")
	}

	// Sort

	sortedArray := []int{1, 2, 3, 4, 5}

	sortArray1 := []int{5, 3, 2, 4, 1}
	sortArray2 := append([]int(nil), sortArray1...)
	sortArray3 := append([]int(nil), sortArray1...)
	sortArray4 := append([]int(nil), sortArray1...)
	sortArray5 := append([]int(nil), sortArray1...)
	

	BubbleSort(sortArray1)
	InsertionSort(sortArray2)
	QuickSort(sortArray3)
	sortedArrayMerge := MergeSort(sortArray4)
	HeapSort(sortArray5)

	for k := range 5 {

		if sortArray1[k] != sortedArray[k] {
			panic("bubble sort")
		}

		if sortArray2[k] != sortedArray[k] {
			panic("insertion sort")
		}

		if sortArray3[k] != sortedArray[k] {
			panic("quick sort")
		}

		if sortedArrayMerge[k] != sortedArray[k] {
			panic("merge sort")
		}

		if sortArray5[k] != sortedArray[k] {
			panic("heap sort")
		}
	}

	// Linked Lists
	
	var node1 Node
	node1.Value = 10

	node3 := Node {
		Value: 10,
		Next: nil,
	}
	node1.Next = &node3

	node2 := new(Node) // Returns a pointer to the struct with 'new'
	node2.Value = 10

	InsertionLinkedList(&node1, node2)

	if node1.Next != node2 || node2.Next != &node3 {
		panic("linked list insertion")
	}

	DeletionLinkedList(&node1)

	if node1.Next != &node3 {
		panic("linked list deletion")
	}
	
	// Singly Linked List

	singlyList := &SinglyList{}
	singlyList.Append(10)
	singlyList.InsertAt(0, 20)
	singlyList.Prepend(30)
	singlyList.InsertAt(1, 40)
	singlyList.Remove()
	singlyList.RemoveAt(0)

	if singlyList.Length != 2 || singlyList.Get(0) != 40 || singlyList.Get(1) != 20 || singlyList.Head.Value != 40 {
		panic("singly linked list")
	}

	// Doubly Lined List
	
	doublyList := &DoublyList{}
	doublyList.InsertTail(10)
	doublyList.InsertAt(0, 20)
	doublyList.InsertHead(30)
	doublyList.InsertAt(1, 40)
	doublyList.RemoveTail()
	doublyList.RemoveAt(0)

	if doublyList.Length != 2 || doublyList.Get(0) != 40 || doublyList.Get(1) != 20 || doublyList.Head.Value != 40 {
		panic("doubly linked list")
	}

	if doublyList.Head.Next != doublyList.Tail || doublyList.Tail.Prev != doublyList.Head ||
		doublyList.GetNode(0) != doublyList.Head || doublyList.GetNode(1) != doublyList.Tail {
		panic("doubly linked list")
	}

	doublyList.RemoveHead()

	if doublyList.Length != 1 || doublyList.Get(0) != 20 || doublyList.Head != doublyList.Tail {
		panic("doubly linked list")
	}

	// Queue
	
	queue := &Queue{}
	queue.Enqueue(10)
	queue.Enqueue(20)
	dequeValue := queue.Deque()
	peekValueQueue := queue.Peek()

	if dequeValue != 10 || peekValueQueue != 20 || queue.Length != 1 {
		panic("queue")
	}

	// Stack
	
	stack := &Stack{}
	stack.Push(10)
	stack.Push(20)
	popValue := stack.Pop()
	peekValueStack := stack.Peek()

	if popValue != 20 || peekValueStack != 10 || stack.Length != 1 {
		panic("stack")
	}

	// Recursion

	if RecursiveSum(5) != 15 {
		panic("recursion")
	}

	maze := []string{
		"###S#",
		"#   #",
		"#   #",
		"#E###",
	}
	path := MazeSolver(maze, "#", Point{3, 0}, Point{1, 3})
	expectedPath := []Point{{3, 0}, {3, 1}, {2, 1}, {1, 1}, {1, 2}, {1, 3}}

	if len(path) != len(expectedPath) {
		panic("maze solver")
	}

	for i := range expectedPath {
		if path[i] != expectedPath[i] {
			panic("maze solver")
		}
	}

	// Tree
	//        7
	//     /    \
	//   23      3
	//  /  \    /  \
	// 5    4  18  21

	treeRoot := &TreeNode{ Value: 7 }
	treeRoot.Insert(&TreeNode{ Value: 23 })
	treeRoot.Insert(&TreeNode{ Value: 3 })
	treeRoot.Children[0].Insert(&TreeNode{ Value: 5 })
	treeRoot.Children[0].Insert(&TreeNode{ Value: 4 })
	treeRoot.Children[1].Insert(&TreeNode{ Value: 18 })
	treeRoot.Children[1].Insert(&TreeNode{ Value: 21 })

	binaryTreeRoot := &BinaryTreeNode{ Value: 7 }
	binaryTreeRoot.Insert(&BinaryTreeNode{ Value: 23 })
	binaryTreeRoot.Insert(&BinaryTreeNode{ Value: 3 })
	binaryTreeRoot.Left.Insert(&BinaryTreeNode{ Value: 5 })
	binaryTreeRoot.Left.Insert(&BinaryTreeNode{ Value: 4 })
	binaryTreeRoot.Right.Insert(&BinaryTreeNode{ Value: 18 })
	binaryTreeRoot.Right.Insert(&BinaryTreeNode{ Value: 21 })

	preOrderSearch := []int{7, 23, 5, 4, 3, 18, 21}
	inOrderSearch := []int{5, 23, 4, 7, 18, 3, 21}
	postOrderSearch := []int{5, 4, 23, 18, 21, 3, 7}

	for i := range len(preOrderSearch) {

		if treeRoot.PreOrderSearch()[i] != preOrderSearch[i] || treeRoot.PostOrderSearch()[i] != postOrderSearch[i] ||
			binaryTreeRoot.PreOrderSearch()[i] != preOrderSearch[i] || binaryTreeRoot.InOrderSearch()[i] != inOrderSearch[i] || binaryTreeRoot.PostOrderSearch()[i] != postOrderSearch[i] {

			panic("trees")
		}
	}

	if !binaryTreeRoot.BreadthFirstSearch(5) {
		panic("binary tree")
	}

	// Binary Search Tree
	//       10
	//     /    \
	//    5      15
	//  /  \    /  \
	// 3    7  12   20

	bst := &BSTNode{ Value: 10 }
	bst.Insert(5)
	bst.Insert(15)
	bst.Insert(3)
	bst.Insert(7)
	bst.Insert(12)
	bst.Insert(20)

	// Duplicates
	bst.Insert(10)
	bst.Insert(7)

	if bst.Find(7) == nil || bst.Find(3) == nil || bst.Find(20) == nil || bst.Find(10) == nil ||
		bst.Find(99) != nil || bst.FindMin().Value != 3 {
		panic("bst find")
	}

	bst.Find(3).Delete()
	if bst.Find(3) != nil || bst.FindMin().Value != 5 {
		panic("bst delete leaf")
	}

	bst.Find(5).Delete()
	if bst.Find(5) != nil || bst.Left == nil || bst.Left.Value != 7 {
		panic("bst delete one children")
	}

	bst.Delete()
	if bst.Value != 12 || bst.Find(12) == nil || bst.Find(7) == nil || bst.Find(15) == nil || bst.Find(20) == nil {
		panic("bst delete two children")
	}

	// Heap

	minHeap := &MinHeap{}
	minHeap.Insert(50)
	minHeap.Insert(10)
	minHeap.Insert(40)
	minHeap.Insert(20)
	minHeap.Insert(30)
	minHeap.Insert(5)

	if minHeap.array[0] != 5 || len(minHeap.array) != 6 {
		panic("min heap insert")
	}

	minHeapExpected := []int{5, 10, 20, 30, 40, 50} // the heap data structure is not sorted like this

	for i := range len(minHeapExpected) {

		if minHeap.Delete() != minHeapExpected[i] {
			panic("min heap delete")
		}
	}

	maxHeap := &MaxHeap{}
	maxHeap.Insert(50)
	maxHeap.Insert(10)
	maxHeap.Insert(40)
	maxHeap.Insert(20)
	maxHeap.Insert(30)
	maxHeap.Insert(5)

	if maxHeap.array[0] != 50 || len(maxHeap.array) != 6 {
		panic("max heap insert")
	}

	maxHeapExpected := []int{50, 40, 30, 20, 10, 5}

	for i := range len(maxHeapExpected) {

		if maxHeap.Delete() != maxHeapExpected[i] {
			panic("max heap delete")
		}
	}

	// Trie

	trie := NewTrieNode()
	trie.Insert("cat")
	trie.Insert("card")
	trie.Insert("cattle")
	trie.Insert("marc")

	if !trie.Search("cat") || !trie.Search("card") || !trie.Search("cattle") || !trie.Search("marc") {
		panic("trie search")
	}

	if !trie.StartsWith("ca") || !trie.StartsWith("cat") || !trie.StartsWith("m") || !trie.StartsWith("marc") {
		panic("trie starts with")
	}

	if !trie.Delete("cat") {
		panic("trie delete")
	}

	if trie.Search("cat") || !trie.Search("card") || !trie.Search("cattle") || !trie.StartsWith("cat") {
		panic("trie delete prefix word")
	}

	if !trie.Delete("card") {
		panic("trie delete")
	}

	if trie.Search("card") || !trie.Search("cattle") || !trie.StartsWith("cat") {
		panic("trie delete leaf")
	}

	if !trie.Delete("marc") {
		panic("trie delete")
	}

	if trie.Search("marc") || trie.StartsWith("d") {
		panic("trie delete branch")
	}

	// Graph

	matrix := [][]int{
		{0, 1, 5, 0},
		{0, 0, 2, 0},
		{0, 0, 0, 1},
		{0, 0, 0, 0},
	}
	graphMatrix := &WeightedAdjMatrix{Matrix: matrix}
	
	bfsPath := graphMatrix.BFS(0, 3)
	expectedBfsPath := []int{0, 2, 3}

	for i := range expectedBfsPath {

		if bfsPath[i] != expectedBfsPath[i] || len(bfsPath) != len(expectedBfsPath) {
			panic("bfs path")
		}
	}

	list := [][]GraphEdge{
		{{To: 1, Weight: 1}, {To: 2, Weight: 5}},
		{{To: 2, Weight: 2}},
		{{To: 3, Weight: 1}},
		{},
	}
	graphList := &WeightedAdjList{List: list}

	dfsPath := graphList.DFS(0, 3)
	expectedDfsPath := []int{0, 1, 2, 3}

	for i := range expectedDfsPath {

		if dfsPath[i] != expectedDfsPath[i] || len(dfsPath) != len(expectedDfsPath) {
			panic("dfs path")
		}
	}

	dijkstraList := [][]GraphEdge{
		{{To: 1, Weight: 3}, {To: 2, Weight: 1}},
		{{To: 3, Weight: 1}},
		{{To: 3, Weight: 7}, {To: 1, Weight: 1}},
		{},
	}
	dijkstraGraph := &WeightedAdjList{List: dijkstraList}

	dijkstraPath := dijkstraGraph.Dijkstra(0, 3)
	expectedDijkstraPath := []int{0, 2, 1, 3}

	for i := range expectedDijkstraPath {

		if dijkstraPath[i] != expectedDijkstraPath[i] || len(dijkstraPath) != len(expectedDijkstraPath) {
			panic("dijkstra path")
		}
	}

	// LRU Cache

	lru := NewLRUCache(2)
	lru.Put("1", 1)
	lru.Put("2", 2) // cache is {1=1, 2=2}

	if val, ok := lru.Get("1"); !ok || val != 1 { // returns 1
		panic("lru get")
	}

	// cache is {2=2, 1=1} (1 is most recently used)
	lru.Put("3", 3) // evicts key 2, cache is {1=1, 3=3}
	if _, ok := lru.Get("2"); ok { // returns not found
		panic("lru evict")
	}

	lru.Put("4", 4) // evicts key 1, cache is {3=3, 4=4}
	if _, ok := lru.Get("1"); ok { // returns not found
		panic("lru evict 2")
	}

	if val, ok := lru.Get("3"); !ok || val != 3 { // returns 3
		panic("lru get 3")
	}

	if val, ok := lru.Get("4"); !ok || val != 4 { // returns 4
		panic("lru get 4")
	}
}
