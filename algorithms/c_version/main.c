#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>

#include "algorithms.h"

/* ------------------------------------------------------------------ */
/* Which sections to run                                               */
/* ------------------------------------------------------------------ */

#ifndef TEST_ALL
#define TEST_ALL 1
#endif

#ifndef TEST_ARRAY
#define TEST_ARRAY TEST_ALL /* 01_array + 07_ringbuffer */
#endif
#ifndef TEST_MAP
#define TEST_MAP TEST_ALL /* 02_map + 02_map_extra */
#endif
#ifndef TEST_SEARCH
#define TEST_SEARCH TEST_ALL /* 03_search */
#endif
#ifndef TEST_SORT
#define TEST_SORT TEST_ALL /* 04_sort */
#endif
#ifndef TEST_LINKED_LIST
#define TEST_LINKED_LIST TEST_ALL /* 05_linked_lists */
#endif
#ifndef TEST_QUEUE_STACK
#define TEST_QUEUE_STACK TEST_ALL /* 06_queue_stack */
#endif
#ifndef TEST_HEAP
#define TEST_HEAP TEST_ALL /* 08_heap */
#endif
#ifndef TEST_RECURSION
#define TEST_RECURSION TEST_ALL /* 09_recursion */
#endif
#ifndef TEST_TREE
#define TEST_TREE TEST_ALL /* 10_tree + 11_binary_tree */
#endif
#ifndef TEST_BST
#define TEST_BST TEST_ALL /* 12_binary_search_tree */
#endif
#ifndef TEST_TRIE
#define TEST_TRIE TEST_ALL /* 13_trie_tree */
#endif
#ifndef TEST_GRAPH
#define TEST_GRAPH TEST_ALL /* 14_graph */
#endif
#ifndef TEST_LRU
#define TEST_LRU TEST_ALL /* 15_lru_cache */
#endif

/* ------------------------------------------------------------------ */
/* Tiny assert harness                                                 */
/* ------------------------------------------------------------------ */

#if defined(__GNUC__) || defined(__clang__)
#define MAYBE_UNUSED __attribute__((unused))
#else
#define MAYBE_UNUSED
#endif

#define RED   "\033[31m"
#define GREEN "\033[32m"
#define DIM   "\033[2m"
#define RESET "\033[0m"

static int total_checks;
static int total_failures;
static int section_failures;

MAYBE_UNUSED static bool check_at(bool ok, const char *label, int line)
{
	total_checks++;

	if (!ok) {
		total_failures++;
		section_failures++;
		printf("  " RED "FAIL" RESET " %-32s " DIM "main.c:%d" RESET "\n", label, line);
	}

	return ok;
}

#define CHECK(cond, label) check_at((cond), (label), __LINE__)

#define REQUIRE(cond, label)                        \
	do {                                            \
		if (!check_at((cond), (label), __LINE__)) { \
			printf("  " DIM "(section aborted)" RESET "\n"); \
			return;                                 \
		}                                           \
	} while (0)

MAYBE_UNUSED static void run_section(const char *name, void (*fn)(void))
{
	printf(DIM "--" RESET " %s\n", name);

	section_failures = 0;
	fn();

	if (section_failures == 0) {
		printf("  " GREEN "ok" RESET "\n");
	}
}

MAYBE_UNUSED static bool int_arrays_equal(const int *a, const int *b, int n)
{
	for (int i = 0; i < n; i++) {
		if (a[i] != b[i]) {
			return false;
		}
	}

	return true;
}

MAYBE_UNUSED static bool points_equal(const Point *a, const Point *b, int n)
{
	for (int i = 0; i < n; i++) {
		if (a[i].x != b[i].x || a[i].y != b[i].y) {
			return false;
		}
	}

	return true;
}

MAYBE_UNUSED static bool strings_equal(char **got, const char *const *want, int n)
{
	for (int i = 0; i < n; i++) {
		if (got[i] == NULL || strcmp(got[i], want[i]) != 0) {
			return false;
		}
	}

	return true;
}

/* ------------------------------------------------------------------ */
/* Array                                                               */
/* ------------------------------------------------------------------ */

#if TEST_ARRAY
static void test_array(void)
{
	Array *customArray = Array_New(3);
	REQUIRE(customArray != NULL, "array new");

	Array_Set(customArray, 0, 1);
	Array_Set(customArray, 1, 2);
	Array_Set(customArray, 2, 3);

	Array *clonedCustomArray = Array_Clone(customArray);
	REQUIRE(clonedCustomArray != NULL, "array clone");

	CHECK(Array_Get(customArray, 0) == 1 &&
	      Array_Get(customArray, 1) == 2 &&
	      Array_Get(customArray, 2) == 3, "array set/get");

	CHECK(Array_Get(customArray, 0) == Array_Get(clonedCustomArray, 0) &&
	      Array_Get(customArray, 1) == Array_Get(clonedCustomArray, 1) &&
	      Array_Get(customArray, 2) == Array_Get(clonedCustomArray, 2), "array clone");

	CHECK(Array_IndexOf(customArray, 2) == 1 &&
	      Array_IndexOf(customArray, 99) == -1, "array index of");

	Array_Clear(customArray);

	CHECK(Array_Len(customArray) == 3 &&
	      Array_Get(customArray, 0) == 0 &&
	      Array_Get(customArray, 1) == 0 &&
	      Array_Get(customArray, 2) == 0, "array clear");

	CHECK(Array_Get(clonedCustomArray, 0) == 1, "array clone is a deep copy");

	Array_Free(customArray);
	Array_Free(clonedCustomArray);

	DynamicArray *dynamicArray = DynamicArray_New();
	REQUIRE(dynamicArray != NULL, "dynamic array new");

	DynamicArray_Push(dynamicArray, 30);
	DynamicArray_Prepend(dynamicArray, 10);
	DynamicArray_InsertAt(dynamicArray, 1, 20);

	REQUIRE(DynamicArray_Len(dynamicArray) == 3, "dynamic array length");
	CHECK(DynamicArray_Get(dynamicArray, 0) == 10 &&
	      DynamicArray_Get(dynamicArray, 1) == 20 &&
	      DynamicArray_Get(dynamicArray, 2) == 30, "dynamic array");

	DynamicArray_RemoveAt(dynamicArray, 1);
	DynamicArray_Pop(dynamicArray);
	DynamicArray_Shift(dynamicArray);

	CHECK(DynamicArray_Len(dynamicArray) == 0, "dynamic array remove");

	for (int i = 0; i < 10; i++) {
		DynamicArray_Push(dynamicArray, i * 100);
	}

	REQUIRE(DynamicArray_Len(dynamicArray) == 10, "dynamic array resize length");
	CHECK(DynamicArray_Get(dynamicArray, 0) == 0 &&
	      DynamicArray_Get(dynamicArray, 9) == 900, "dynamic array resize");

	DynamicArray_Free(dynamicArray);

	RingBuffer *ringBuffer = RingBuffer_New(3);
	REQUIRE(ringBuffer != NULL, "ring buffer new");

	RingBuffer_Enqueue(ringBuffer, 10);
	RingBuffer_Enqueue(ringBuffer, 20);

	int ringBufferValue1 = RingBuffer_Peek(ringBuffer);
	RingBuffer_Dequeue(ringBuffer);
	int ringBufferValue2 = RingBuffer_Peek(ringBuffer);
	RingBuffer_Dequeue(ringBuffer);

	CHECK(ringBufferValue1 == 10 && ringBufferValue2 == 20, "ring buffer");

	RingBuffer_Enqueue(ringBuffer, 1);
	RingBuffer_Enqueue(ringBuffer, 2);
	RingBuffer_Enqueue(ringBuffer, 3);

	CHECK(RingBuffer_Dequeue(ringBuffer) == 1 &&
	      RingBuffer_Dequeue(ringBuffer) == 2 &&
	      RingBuffer_Dequeue(ringBuffer) == 3, "ring buffer wrap");

	RingBuffer_Free(ringBuffer);
}
#endif

/* ------------------------------------------------------------------ */
/* Map                                                                 */
/* ------------------------------------------------------------------ */

#if TEST_MAP
static void test_map(void)
{
	int val;

	HashMap *hashMap = HashMap_New(4);
	REQUIRE(hashMap != NULL, "map new");

	HashMap_Put(hashMap, "apple", 1);
	HashMap_Put(hashMap, "banana", 2);
	HashMap_Put(hashMap, "cherry", 3);

	val = 0;
	CHECK(HashMap_Get(hashMap, "banana", &val) && val == 2, "map get");

	HashMap_Put(hashMap, "apple", 10); /* Update */
	val = 0;
	CHECK(HashMap_Get(hashMap, "apple", &val) && val == 10, "map update");

	CHECK(HashMap_Has(hashMap, "cherry") && HashMap_Size(hashMap) == 3, "map has/size");

	CHECK(!HashMap_Get(hashMap, "durian", &val), "map get missing");

	HashMap_Delete(hashMap, "banana");
	CHECK(!HashMap_Has(hashMap, "banana") && HashMap_Size(hashMap) == 2, "map delete");

	HashMap_Clear(hashMap);
	CHECK(HashMap_Size(hashMap) == 0 && !HashMap_Has(hashMap, "apple"), "map clear");

	HashMap_Free(hashMap);

	/* TreeMap (Ordered) */

	TreeMap *treeMap = TreeMap_New();
	REQUIRE(treeMap != NULL, "tree map new");

	TreeMap_Put(treeMap, "c", 3);
	TreeMap_Put(treeMap, "a", 1);
	TreeMap_Put(treeMap, "b", 2);

	int orderedLen = 0;
	char **orderedKeys = TreeMap_InOrderKeys(treeMap, &orderedLen);
	const char *const wantOrdered[] = { "a", "b", "c" };

	CHECK(orderedKeys != NULL && orderedLen == 3 &&
	      strings_equal(orderedKeys, wantOrdered, 3), "tree map order");

	free(orderedKeys);

	val = 0;
	CHECK(TreeMap_Get(treeMap, "b", &val) && val == 2, "tree map get");
	CHECK(!TreeMap_Get(treeMap, "z", &val), "tree map get missing");

	TreeMap_Free(treeMap);

	/* LinkedHashMap (Insertion-Ordered) */

	LinkedHashMap *linkedMap = LinkedHashMap_New(4);
	REQUIRE(linkedMap != NULL, "linked map new");

	LinkedHashMap_Put(linkedMap, "c", 3);
	LinkedHashMap_Put(linkedMap, "a", 1);
	LinkedHashMap_Put(linkedMap, "b", 2);

	int insertionLen = 0;
	char **insertionKeys = LinkedHashMap_GetKeysInOrder(linkedMap, &insertionLen);
	const char *const wantInsertion[] = { "c", "a", "b" };

	CHECK(insertionKeys != NULL && insertionLen == 3 &&
	      strings_equal(insertionKeys, wantInsertion, 3), "linked map order");

	free(insertionKeys);

	val = 0;
	CHECK(LinkedHashMap_Get(linkedMap, "a", &val) && val == 1, "linked map get");

	LinkedHashMap_Free(linkedMap);

	/* ConcurrentHashMap (Thread-Safe) */

	ConcurrentHashMap *concurrentMap = ConcurrentHashMap_New(16);
	REQUIRE(concurrentMap != NULL, "concurrent map new");

	ConcurrentHashMap_Put(concurrentMap, "safe", 99);

	val = 0;
	CHECK(ConcurrentHashMap_Get(concurrentMap, "safe", &val) && val == 99, "concurrent map get");

	ConcurrentHashMap_Free(concurrentMap);
}
#endif

/* ------------------------------------------------------------------ */
/* Search                                                              */
/* ------------------------------------------------------------------ */

#if TEST_SEARCH
static void test_search(void)
{
	const int array[] = { 21, 49, 42, 12, 37, 8, 5, 2, 19, 33 };
	const int primes[] = { 2, 3, 5, 7, 11, 13, 17, 19, 23, 29 };
	const int n = 10;

	CHECK(LinearSearch(array, n, 5), "linear search");
	CHECK(!LinearSearch(array, n, 6), "linear search missing");

	CHECK(BinarySearch(primes, n, 5) && BinarySearch(primes, n, 29), "binary search");
	CHECK(!BinarySearch(primes, n, 4), "binary search missing");

	CHECK(BinarySearchRange(primes, 0, n, 5) &&
	      BinarySearchRange(primes, 0, n, 29), "binary search range");

	CHECK(InterpolationSearch(primes, n, 5) &&
	      InterpolationSearch(primes, n, 29), "interpolation search");
	CHECK(!InterpolationSearch(primes, n, 4), "interpolation search missing");

	CHECK(ExponentialSearch(primes, n, 5) &&
	      ExponentialSearch(primes, n, 29), "exponential search");
	CHECK(ExponentialSearch(primes, n, 2), "exponential search first");
	CHECK(!ExponentialSearch(primes, n, 30), "exponential search missing");

	CHECK(JumpSearch(primes, n, 5) && JumpSearch(primes, n, 29), "jump search");
	CHECK(!JumpSearch(primes, n, 4), "jump search missing");

	const bool twoCrystalBalls[] = { false, false, false, true, true };
	CHECK(TwoCrystalBalls(twoCrystalBalls, 5) == 3, "two crystal balls");

	const bool neverBreaks[] = { false, false, false, false, false };
	CHECK(TwoCrystalBalls(neverBreaks, 5) == -1, "two crystal balls never breaks");
}
#endif

/* ------------------------------------------------------------------ */
/* Sort                                                                */
/* ------------------------------------------------------------------ */

#if TEST_SORT
static void test_sort(void)
{
	const int sortedArray[] = { 1, 2, 3, 4, 5 };

	int sortArray1[] = { 5, 3, 2, 4, 1 };
	int sortArray2[5], sortArray3[5], sortArray4[5], sortArray5[5];

	memcpy(sortArray2, sortArray1, sizeof sortArray1);
	memcpy(sortArray3, sortArray1, sizeof sortArray1);
	memcpy(sortArray4, sortArray1, sizeof sortArray1);
	memcpy(sortArray5, sortArray1, sizeof sortArray1);

	BubbleSort(sortArray1, 5);
	InsertionSort(sortArray2, 5);
	QuickSort(sortArray3, 5);
	int *sortedArrayMerge = MergeSort(sortArray4, 5);
	HeapSort(sortArray5, 5);

	CHECK(int_arrays_equal(sortArray1, sortedArray, 5), "bubble sort");
	CHECK(int_arrays_equal(sortArray2, sortedArray, 5), "insertion sort");
	CHECK(int_arrays_equal(sortArray3, sortedArray, 5), "quick sort");
	CHECK(sortedArrayMerge != NULL &&
	      int_arrays_equal(sortedArrayMerge, sortedArray, 5), "merge sort");
	CHECK(int_arrays_equal(sortArray5, sortedArray, 5), "heap sort");

	free(sortedArrayMerge);

	const int wantDuplicates[] = { 1, 2, 2, 3, 3, 3, 9 };
	int duplicates[] = { 3, 1, 3, 2, 9, 2, 3 };
	QuickSort(duplicates, 7);
	CHECK(int_arrays_equal(duplicates, wantDuplicates, 7), "quick sort duplicates");

	int already[] = { 1, 2, 3, 4, 5 };
	QuickSort(already, 5);
	CHECK(int_arrays_equal(already, sortedArray, 5), "quick sort sorted input");

	int single[] = { 7 };
	BubbleSort(single, 1);
	InsertionSort(single, 1);
	QuickSort(single, 1);
	HeapSort(single, 1);
	CHECK(single[0] == 7, "sort single element");
}
#endif

/* ------------------------------------------------------------------ */
/* Linked Lists                                                        */
/* ------------------------------------------------------------------ */

#if TEST_LINKED_LIST
static void test_linked_list(void)
{
	Node node1 = { 0, NULL, NULL };
	node1.Value = 10;

	Node node3 = { 10, NULL, NULL };
	node1.Next = &node3;

	Node *node2 = calloc(1, sizeof *node2); /* Go's new(Node) */
	REQUIRE(node2 != NULL, "node alloc");
	node2->Value = 10;

	InsertionLinkedList(&node1, node2);
	CHECK(node1.Next == node2 && node2->Next == &node3, "linked list insertion");

	DeletionLinkedList(&node1);
	CHECK(node1.Next == &node3, "linked list deletion");

	free(node2);

	/* Singly Linked List */

	SinglyList singlyList = { NULL, 0 };
	SinglyList_Append(&singlyList, 10);
	SinglyList_InsertAt(&singlyList, 0, 20);
	SinglyList_Prepend(&singlyList, 30);
	SinglyList_InsertAt(&singlyList, 1, 40);
	SinglyList_Remove(&singlyList);
	SinglyList_RemoveAt(&singlyList, 0);

	REQUIRE(singlyList.Length == 2 && singlyList.Head != NULL, "singly linked list length");
	CHECK(SinglyList_Get(&singlyList, 0) == 40 &&
	      SinglyList_Get(&singlyList, 1) == 20 &&
	      singlyList.Head->Value == 40, "singly linked list");

	SinglyList_Free(&singlyList);

	/* Doubly Linked List */

	DoublyList doublyList = { NULL, NULL, 0 };
	DoublyList_InsertTail(&doublyList, 10);
	DoublyList_InsertAt(&doublyList, 0, 20);
	DoublyList_InsertHead(&doublyList, 30);
	DoublyList_InsertAt(&doublyList, 1, 40);
	DoublyList_RemoveTail(&doublyList);
	DoublyList_RemoveAt(&doublyList, 0);

	REQUIRE(doublyList.Length == 2 &&
	        doublyList.Head != NULL &&
	        doublyList.Tail != NULL, "doubly linked list length");

	CHECK(DoublyList_Get(&doublyList, 0) == 40 &&
	      DoublyList_Get(&doublyList, 1) == 20 &&
	      doublyList.Head->Value == 40, "doubly linked list");

	CHECK(doublyList.Head->Next == doublyList.Tail &&
	      doublyList.Tail->Prev == doublyList.Head &&
	      doublyList.Head->Prev == NULL &&
	      doublyList.Tail->Next == NULL, "doubly linked list links");

	CHECK(DoublyList_GetNode(&doublyList, 0) == doublyList.Head &&
	      DoublyList_GetNode(&doublyList, 1) == doublyList.Tail, "doubly linked list get node");

	DoublyList_RemoveHead(&doublyList);

	REQUIRE(doublyList.Length == 1 && doublyList.Head != NULL, "doubly linked list remove head");
	CHECK(DoublyList_Get(&doublyList, 0) == 20 &&
	      doublyList.Head == doublyList.Tail, "doubly linked list single node");

	DoublyList_Free(&doublyList);
}
#endif

/* ------------------------------------------------------------------ */
/* Queue and Stack                                                     */
/* ------------------------------------------------------------------ */

#if TEST_QUEUE_STACK
static void test_queue_stack(void)
{
	Queue queue = { 0, NULL, NULL };
	Queue_Enqueue(&queue, 10);
	Queue_Enqueue(&queue, 20);

	int dequeValue = Queue_Deque(&queue);
	int peekValueQueue = Queue_Peek(&queue);

	CHECK(dequeValue == 10 && peekValueQueue == 20 && queue.Length == 1, "queue");

	Queue_Deque(&queue);
	CHECK(queue.Length == 0 && queue.Head == NULL && queue.Tail == NULL, "queue empty");

	Queue_Enqueue(&queue, 30);
	CHECK(queue.Length == 1 && Queue_Peek(&queue) == 30, "queue reuse");

	Queue_Free(&queue);

	Stack stack = { 0, NULL };
	Stack_Push(&stack, 10);
	Stack_Push(&stack, 20);

	int popValue = Stack_Pop(&stack);
	int peekValueStack = Stack_Peek(&stack);

	CHECK(popValue == 20 && peekValueStack == 10 && stack.Length == 1, "stack");

	Stack_Pop(&stack);
	CHECK(stack.Length == 0 && stack.Head == NULL, "stack empty");

	Stack_Free(&stack);
}
#endif

/* ------------------------------------------------------------------ */
/* Heap                                                                */
/* ------------------------------------------------------------------ */

#if TEST_HEAP
static void test_heap(void)
{
	MinHeap minHeap = { NULL, 0, 0 };
	MinHeap_Insert(&minHeap, 50);
	MinHeap_Insert(&minHeap, 10);
	MinHeap_Insert(&minHeap, 40);
	MinHeap_Insert(&minHeap, 20);
	MinHeap_Insert(&minHeap, 30);
	MinHeap_Insert(&minHeap, 5);

	REQUIRE(minHeap.array != NULL && minHeap.length == 6, "min heap insert");
	CHECK(minHeap.array[0] == 5, "min heap root");

	const int minHeapExpected[] = { 5, 10, 20, 30, 40, 50 };
	int minHeapGot[6];

	for (int i = 0; i < 6; i++) {
		minHeapGot[i] = MinHeap_Delete(&minHeap);
	}

	CHECK(int_arrays_equal(minHeapGot, minHeapExpected, 6), "min heap delete");
	CHECK(minHeap.length == 0, "min heap empty");

	MinHeap_Free(&minHeap);

	MaxHeap maxHeap = { NULL, 0, 0 };
	MaxHeap_Insert(&maxHeap, 50);
	MaxHeap_Insert(&maxHeap, 10);
	MaxHeap_Insert(&maxHeap, 40);
	MaxHeap_Insert(&maxHeap, 20);
	MaxHeap_Insert(&maxHeap, 30);
	MaxHeap_Insert(&maxHeap, 5);

	REQUIRE(maxHeap.array != NULL && maxHeap.length == 6, "max heap insert");
	CHECK(maxHeap.array[0] == 50, "max heap root");

	const int maxHeapExpected[] = { 50, 40, 30, 20, 10, 5 };
	int maxHeapGot[6];

	for (int i = 0; i < 6; i++) {
		maxHeapGot[i] = MaxHeap_Delete(&maxHeap);
	}

	CHECK(int_arrays_equal(maxHeapGot, maxHeapExpected, 6), "max heap delete");
	CHECK(maxHeap.length == 0, "max heap empty");

	MaxHeap_Free(&maxHeap);
}
#endif

/* ------------------------------------------------------------------ */
/* Recursion                                                           */
/* ------------------------------------------------------------------ */

#if TEST_RECURSION
static void test_recursion(void)
{
	CHECK(RecursiveSum(5) == 15, "recursion");
	CHECK(RecursiveSum(1) == 1, "recursion base case");

	const char *maze[] = {
		"###S#",
		"#   #",
		"#   #",
		"#E###",
	};

	Point start = { 3, 0 };
	Point end = { 1, 3 };

	int pathLen = 0;
	Point *path = MazeSolver(maze, 4, '#', start, end, &pathLen);

	const Point expectedPath[] = { { 3, 0 }, { 3, 1 }, { 2, 1 }, { 1, 1 }, { 1, 2 }, { 1, 3 } };

	REQUIRE(path != NULL, "maze solver");
	CHECK(pathLen == 6 && points_equal(path, expectedPath, 6), "maze solver path");

	free(path);

	const char *closedMaze[] = {
		"###S#",
		"#####",
		"#   #",
		"#E###",
	};

	int closedLen = 0;
	Point *closedPath = MazeSolver(closedMaze, 4, '#', start, end, &closedLen);

	CHECK(closedPath == NULL, "maze solver no path");
	free(closedPath);
}
#endif

/* ------------------------------------------------------------------ */
/* Tree                                                                */
/* ------------------------------------------------------------------ */

#if TEST_TREE
static void test_tree(void)
{
	/*        7
	 *     /    \
	 *   23      3
	 *  /  \    /  \
	 * 5    4  18  21
	 */

	TreeNode *treeRoot = TreeNode_New(7);
	REQUIRE(treeRoot != NULL, "tree new");

	TreeNode_Insert(treeRoot, TreeNode_New(23));
	TreeNode_Insert(treeRoot, TreeNode_New(3));

	REQUIRE(treeRoot->ChildCount == 2 &&
	        treeRoot->Children != NULL &&
	        treeRoot->Children[0] != NULL &&
	        treeRoot->Children[1] != NULL, "tree insert");

	TreeNode_Insert(treeRoot->Children[0], TreeNode_New(5));
	TreeNode_Insert(treeRoot->Children[0], TreeNode_New(4));
	TreeNode_Insert(treeRoot->Children[1], TreeNode_New(18));
	TreeNode_Insert(treeRoot->Children[1], TreeNode_New(21));

	CHECK(treeRoot->Children[0]->Parent == treeRoot, "tree parent link");

	BinaryTreeNode *binaryTreeRoot = BinaryTreeNode_New(7);
	REQUIRE(binaryTreeRoot != NULL, "binary tree new");

	BinaryTreeNode_Insert(binaryTreeRoot, BinaryTreeNode_New(23));
	BinaryTreeNode_Insert(binaryTreeRoot, BinaryTreeNode_New(3));

	REQUIRE(binaryTreeRoot->Left != NULL && binaryTreeRoot->Right != NULL, "binary tree insert");

	BinaryTreeNode_Insert(binaryTreeRoot->Left, BinaryTreeNode_New(5));
	BinaryTreeNode_Insert(binaryTreeRoot->Left, BinaryTreeNode_New(4));
	BinaryTreeNode_Insert(binaryTreeRoot->Right, BinaryTreeNode_New(18));
	BinaryTreeNode_Insert(binaryTreeRoot->Right, BinaryTreeNode_New(21));

	const int preOrderSearch[] = { 7, 23, 5, 4, 3, 18, 21 };
	const int inOrderSearch[] = { 5, 23, 4, 7, 18, 3, 21 };
	const int postOrderSearch[] = { 5, 4, 23, 18, 21, 3, 7 };

	int len;
	int *walk;

	len = 0;
	walk = TreeNode_PreOrderSearch(treeRoot, &len);
	CHECK(walk != NULL && len == 7 && int_arrays_equal(walk, preOrderSearch, 7), "tree pre order");
	free(walk);

	len = 0;
	walk = TreeNode_PostOrderSearch(treeRoot, &len);
	CHECK(walk != NULL && len == 7 && int_arrays_equal(walk, postOrderSearch, 7), "tree post order");
	free(walk);

	len = 0;
	walk = BinaryTreeNode_PreOrderSearch(binaryTreeRoot, &len);
	CHECK(walk != NULL && len == 7 && int_arrays_equal(walk, preOrderSearch, 7), "binary tree pre order");
	free(walk);

	len = 0;
	walk = BinaryTreeNode_InOrderSearch(binaryTreeRoot, &len);
	CHECK(walk != NULL && len == 7 && int_arrays_equal(walk, inOrderSearch, 7), "binary tree in order");
	free(walk);

	len = 0;
	walk = BinaryTreeNode_PostOrderSearch(binaryTreeRoot, &len);
	CHECK(walk != NULL && len == 7 && int_arrays_equal(walk, postOrderSearch, 7), "binary tree post order");
	free(walk);

	CHECK(BinaryTreeNode_BreadthFirstSearch(binaryTreeRoot, 5), "binary tree bfs");
	CHECK(!BinaryTreeNode_BreadthFirstSearch(binaryTreeRoot, 99), "binary tree bfs missing");

	CHECK(CompareBinaryTree(binaryTreeRoot, binaryTreeRoot), "compare binary tree same");
	CHECK(!CompareBinaryTree(binaryTreeRoot, binaryTreeRoot->Left), "compare binary tree different");

	TreeNode *detached = treeRoot->Children[0];
	TreeNode_Delete(detached);

	CHECK(treeRoot->ChildCount == 1 &&
	      treeRoot->Children[0]->Value == 3 &&
	      detached->Parent == NULL, "tree delete");

	TreeNode_Delete(treeRoot); /* the root has no parent, this must be a no-op */
	CHECK(treeRoot->ChildCount == 1, "tree delete root is a no-op");

	TreeNode_Free(detached);
	TreeNode_Free(treeRoot);
	BinaryTreeNode_Free(binaryTreeRoot);
}
#endif

/* ------------------------------------------------------------------ */
/* Binary Search Tree                                                  */
/* ------------------------------------------------------------------ */

#if TEST_BST
static void test_bst(void)
{
	/*       10
	 *     /    \
	 *    5      15
	 *  /  \    /  \
	 * 3    7  12   20
	 */

	BSTNode *bst = BSTNode_New(10);
	REQUIRE(bst != NULL, "bst new");

	BSTNode_Insert(bst, 5);
	BSTNode_Insert(bst, 15);
	BSTNode_Insert(bst, 3);
	BSTNode_Insert(bst, 7);
	BSTNode_Insert(bst, 12);
	BSTNode_Insert(bst, 20);

	BSTNode_Insert(bst, 10);
	BSTNode_Insert(bst, 7);

	CHECK(BSTNode_Find(bst, 7) != NULL &&
	      BSTNode_Find(bst, 3) != NULL &&
	      BSTNode_Find(bst, 20) != NULL &&
	      BSTNode_Find(bst, 10) != NULL &&
	      BSTNode_Find(bst, 99) == NULL, "bst find");

	REQUIRE(BSTNode_FindMin(bst) != NULL, "bst find min");
	CHECK(BSTNode_FindMin(bst)->Value == 3, "bst find min value");

	BSTNode *leaf = BSTNode_Find(bst, 3);
	REQUIRE(leaf != NULL, "bst find leaf");
	BSTNode_Delete(leaf);

	CHECK(BSTNode_Find(bst, 3) == NULL &&
	      BSTNode_FindMin(bst)->Value == 5, "bst delete leaf");

	BSTNode *oneChild = BSTNode_Find(bst, 5);
	REQUIRE(oneChild != NULL, "bst find one child node");
	BSTNode_Delete(oneChild);

	CHECK(BSTNode_Find(bst, 5) == NULL &&
	      bst->Left != NULL &&
	      bst->Left->Value == 7, "bst delete one children");

	BSTNode_Delete(bst); /* root, two children */

	CHECK(bst->Value == 12 &&
	      BSTNode_Find(bst, 12) != NULL &&
	      BSTNode_Find(bst, 7) != NULL &&
	      BSTNode_Find(bst, 15) != NULL &&
	      BSTNode_Find(bst, 20) != NULL, "bst delete two children");

	BSTNode_Free(bst);
}
#endif

/* ------------------------------------------------------------------ */
/* Trie                                                                */
/* ------------------------------------------------------------------ */

#if TEST_TRIE
static void test_trie(void)
{
	TrieNode *trie = TrieNode_New();
	REQUIRE(trie != NULL, "trie new");

	TrieNode_Insert(trie, "cat");
	TrieNode_Insert(trie, "card");
	TrieNode_Insert(trie, "cattle");
	TrieNode_Insert(trie, "marc");

	CHECK(TrieNode_Search(trie, "cat") &&
	      TrieNode_Search(trie, "card") &&
	      TrieNode_Search(trie, "cattle") &&
	      TrieNode_Search(trie, "marc"), "trie search");

	CHECK(!TrieNode_Search(trie, "ca") &&
	      !TrieNode_Search(trie, "cats"), "trie search prefix is not a word");

	CHECK(TrieNode_StartsWith(trie, "ca") &&
	      TrieNode_StartsWith(trie, "cat") &&
	      TrieNode_StartsWith(trie, "m") &&
	      TrieNode_StartsWith(trie, "marc"), "trie starts with");

	CHECK(TrieNode_Delete(trie, "cat"), "trie delete");
	CHECK(!TrieNode_Search(trie, "cat") &&
	      TrieNode_Search(trie, "card") &&
	      TrieNode_Search(trie, "cattle") &&
	      TrieNode_StartsWith(trie, "cat"), "trie delete prefix word");

	CHECK(TrieNode_Delete(trie, "card"), "trie delete");
	CHECK(!TrieNode_Search(trie, "card") &&
	      TrieNode_Search(trie, "cattle") &&
	      TrieNode_StartsWith(trie, "cat"), "trie delete leaf");

	CHECK(TrieNode_Delete(trie, "marc"), "trie delete");
	CHECK(!TrieNode_Search(trie, "marc") &&
	      !TrieNode_StartsWith(trie, "d") &&
	      !TrieNode_StartsWith(trie, "m"), "trie delete branch");

	CHECK(!TrieNode_Delete(trie, "nope"), "trie delete missing word");

	TrieNode_Free(trie);
}
#endif

/* ------------------------------------------------------------------ */
/* Graph                                                               */
/* ------------------------------------------------------------------ */

#if TEST_GRAPH
static void test_graph(void)
{
	int matrixRow0[] = { 0, 1, 5, 0 };
	int matrixRow1[] = { 0, 0, 2, 0 };
	int matrixRow2[] = { 0, 0, 0, 1 };
	int matrixRow3[] = { 0, 0, 0, 0 };
	int *matrix[] = { matrixRow0, matrixRow1, matrixRow2, matrixRow3 };

	WeightedAdjMatrix graphMatrix = { .Matrix = matrix, .Size = 4 };

	int bfsLen = 0;
	int *bfsPath = WeightedAdjMatrix_BFS(&graphMatrix, 0, 3, &bfsLen);
	const int expectedBfsPath[] = { 0, 2, 3 };

	CHECK(bfsPath != NULL && bfsLen == 3 &&
	      int_arrays_equal(bfsPath, expectedBfsPath, 3), "bfs path");
	free(bfsPath);

	GraphEdge listNode0[] = { { 1, 1 }, { 2, 5 } };
	GraphEdge listNode1[] = { { 2, 2 } };
	GraphEdge listNode2[] = { { 3, 1 } };
	GraphEdge *list[] = { listNode0, listNode1, listNode2, NULL };
	int listCounts[] = { 2, 1, 1, 0 };

	WeightedAdjList graphList = { .List = list, .Counts = listCounts, .Size = 4 };

	int dfsLen = 0;
	int *dfsPath = WeightedAdjList_DFS(&graphList, 0, 3, &dfsLen);
	const int expectedDfsPath[] = { 0, 1, 2, 3 };

	CHECK(dfsPath != NULL && dfsLen == 4 &&
	      int_arrays_equal(dfsPath, expectedDfsPath, 4), "dfs path");
	free(dfsPath);

	int dfsNoneLen = 0;
	int *dfsNone = WeightedAdjList_DFS(&graphList, 3, 0, &dfsNoneLen);
	CHECK(dfsNone == NULL, "dfs unreachable");
	free(dfsNone);

	GraphEdge dijkstraNode0[] = { { 1, 3 }, { 2, 1 } };
	GraphEdge dijkstraNode1[] = { { 3, 1 } };
	GraphEdge dijkstraNode2[] = { { 3, 7 }, { 1, 1 } };
	GraphEdge *dijkstraList[] = { dijkstraNode0, dijkstraNode1, dijkstraNode2, NULL };
	int dijkstraCounts[] = { 2, 1, 2, 0 };

	WeightedAdjList dijkstraGraph = {
		.List = dijkstraList, .Counts = dijkstraCounts, .Size = 4
	};

	int dijkstraLen = 0;
	int *dijkstraPath = WeightedAdjList_Dijkstra(&dijkstraGraph, 0, 3, &dijkstraLen);
	const int expectedDijkstraPath[] = { 0, 2, 1, 3 };

	CHECK(dijkstraPath != NULL && dijkstraLen == 4 &&
	      int_arrays_equal(dijkstraPath, expectedDijkstraPath, 4), "dijkstra path");
	free(dijkstraPath);
}
#endif

/* ------------------------------------------------------------------ */
/* LRU Cache                                                           */
/* ------------------------------------------------------------------ */

#if TEST_LRU
static void test_lru(void)
{
	int val;

	LRUCache *lru = LRUCache_New(2);
	REQUIRE(lru != NULL, "lru new");

	LRUCache_Put(lru, "1", 1);
	LRUCache_Put(lru, "2", 2); /* cache is {1=1, 2=2} */

	val = 0;
	CHECK(LRUCache_Get(lru, "1", &val) && val == 1, "lru get"); /* returns 1 */

	/* cache is {2=2, 1=1} (1 is most recently used) */
	LRUCache_Put(lru, "3", 3); /* evicts key 2, cache is {1=1, 3=3} */
	CHECK(!LRUCache_Get(lru, "2", &val), "lru evict"); /* returns not found */

	LRUCache_Put(lru, "4", 4); /* evicts key 1, cache is {3=3, 4=4} */
	CHECK(!LRUCache_Get(lru, "1", &val), "lru evict 2"); /* returns not found */

	val = 0;
	CHECK(LRUCache_Get(lru, "3", &val) && val == 3, "lru get 3"); /* returns 3 */

	val = 0;
	CHECK(LRUCache_Get(lru, "4", &val) && val == 4, "lru get 4"); /* returns 4 */

	LRUCache_Put(lru, "3", 33);
	val = 0;
	CHECK(LRUCache_Get(lru, "3", &val) && val == 33, "lru update");
	CHECK(LRUCache_Get(lru, "4", &val) && val == 4, "lru update keeps other key");

	LRUCache_Free(lru);
}
#endif

/* ------------------------------------------------------------------ */

int main(void)
{
	printf("\n");

#if TEST_ARRAY
	run_section("Array", test_array);
#endif
#if TEST_MAP
	run_section("Map", test_map);
#endif
#if TEST_SEARCH
	run_section("Search", test_search);
#endif
#if TEST_SORT
	run_section("Sort", test_sort);
#endif
#if TEST_LINKED_LIST
	run_section("Linked Lists", test_linked_list);
#endif
#if TEST_QUEUE_STACK
	run_section("Queue and Stack", test_queue_stack);
#endif
#if TEST_HEAP
	run_section("Heap", test_heap);
#endif
#if TEST_RECURSION
	run_section("Recursion", test_recursion);
#endif
#if TEST_TREE
	run_section("Tree", test_tree);
#endif
#if TEST_BST
	run_section("Binary Search Tree", test_bst);
#endif
#if TEST_TRIE
	run_section("Trie", test_trie);
#endif
#if TEST_GRAPH
	run_section("Graph", test_graph);
#endif
#if TEST_LRU
	run_section("LRU Cache", test_lru);
#endif

	if (total_checks == 0) {
		printf("\n" DIM "no sections enabled" RESET "\n\n");
		return 1;
	}

	if (total_failures == 0) {
		printf("\n" GREEN "%d checks passed" RESET "\n\n", total_checks);
		return 0;
	}

	printf("\n" RED "%d of %d checks failed" RESET "\n\n", total_failures, total_checks);
	return 1;
}
