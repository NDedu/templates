#ifndef ALGORITHMS_H
#define ALGORITHMS_H

#include <stdbool.h>
#include <stddef.h>

/* ------------------------------------------------------------------ */
/* 01_array                                                            */
/* ------------------------------------------------------------------ */

typedef struct Array Array;

Array *Array_New(int capacity);
void   Array_Free(Array *arr);
int    Array_Len(const Array *arr);
int    Array_Get(const Array *arr, int index);
void   Array_Set(Array *arr, int index, int val);
void   Array_Clear(Array *arr);
int    Array_IndexOf(const Array *arr, int targetValue);
Array *Array_Clone(const Array *arr);

typedef struct DynamicArray DynamicArray;

DynamicArray *DynamicArray_New(void);
void DynamicArray_Free(DynamicArray *arr);
int  DynamicArray_Len(const DynamicArray *arr);
int  DynamicArray_Get(const DynamicArray *arr, int index);
void DynamicArray_Push(DynamicArray *arr, int val);
int  DynamicArray_Pop(DynamicArray *arr);
void DynamicArray_Prepend(DynamicArray *arr, int val);
int  DynamicArray_Shift(DynamicArray *arr);
void DynamicArray_InsertAt(DynamicArray *arr, int index, int val);
void DynamicArray_RemoveAt(DynamicArray *arr, int index);

/* ------------------------------------------------------------------ */
/* 02_map                                                              */
/* ------------------------------------------------------------------ */

typedef struct HashMap HashMap;

HashMap *HashMap_New(int capacity);
void HashMap_Free(HashMap *m);
void HashMap_Put(HashMap *m, const char *key, int value);
bool HashMap_Get(const HashMap *m, const char *key, int *outValue);
void HashMap_Delete(HashMap *m, const char *key);
bool HashMap_Has(const HashMap *m, const char *key);
int  HashMap_Size(const HashMap *m);
void HashMap_Clear(HashMap *m);

/* ------------------------------------------------------------------ */
/* 02_map_extra                                                        */
/* ------------------------------------------------------------------ */

typedef struct TreeMap TreeMap;

TreeMap *TreeMap_New(void);
void TreeMap_Free(TreeMap *t);
void TreeMap_Put(TreeMap *t, const char *key, int value);
bool TreeMap_Get(const TreeMap *t, const char *key, int *outValue);
char **TreeMap_InOrderKeys(const TreeMap *t, int *outLen);

typedef struct LinkedHashMap LinkedHashMap;

LinkedHashMap *LinkedHashMap_New(int capacity);
void LinkedHashMap_Free(LinkedHashMap *m);
void LinkedHashMap_Put(LinkedHashMap *m, const char *key, int value);
bool LinkedHashMap_Get(const LinkedHashMap *m, const char *key, int *outValue);
char **LinkedHashMap_GetKeysInOrder(const LinkedHashMap *m, int *outLen);

typedef struct ConcurrentHashMap ConcurrentHashMap;

ConcurrentHashMap *ConcurrentHashMap_New(int capacity);
void ConcurrentHashMap_Free(ConcurrentHashMap *m);
void ConcurrentHashMap_Put(ConcurrentHashMap *m, const char *key, int value);
bool ConcurrentHashMap_Get(ConcurrentHashMap *m, const char *key, int *outValue);

/* ------------------------------------------------------------------ */
/* 03_search - all ranges are inclusive low, exclusive high [low, high) */
/* ------------------------------------------------------------------ */

bool LinearSearch(const int *haystack, int n, int needle);
bool BinarySearch(const int *haystack, int n, int needle);
bool BinarySearchRange(const int *haystack, int low, int high, int needle);
bool InterpolationSearch(const int *haystack, int n, int needle);
bool ExponentialSearch(const int *haystack, int n, int needle);
bool JumpSearch(const int *haystack, int n, int needle);
int  TwoCrystalBalls(const bool *breaks, int n);

/* ------------------------------------------------------------------ */
/* 04_sort                                                             */
/* ------------------------------------------------------------------ */

void BubbleSort(int *array, int n);
void InsertionSort(int *array, int n);
void QuickSort(int *array, int n);
void HeapSort(int *array, int n);
int *MergeSort(const int *array, int n);

/* ------------------------------------------------------------------ */
/* 05_linked_lists                                                     */
/* ------------------------------------------------------------------ */

typedef struct Node Node;
struct Node {
	int   Value;
	Node *Next;
	Node *Prev; /* not part of a Singly Linked List but used by Doubly */
};

void InsertionLinkedList(Node *node, Node *newNode);
void DeletionLinkedList(Node *node); /* deletes the node after the given node */

typedef struct {
	Node *Head;
	int   Length;
} SinglyList;

void SinglyList_Free(SinglyList *list);
void SinglyList_InsertAt(SinglyList *list, int index, int val);
void SinglyList_Remove(SinglyList *list); /* removes the tail */
void SinglyList_RemoveAt(SinglyList *list, int index);
void SinglyList_Append(SinglyList *list, int val);
void SinglyList_Prepend(SinglyList *list, int val);
int  SinglyList_Get(const SinglyList *list, int index);

typedef struct {
	Node *Head;
	Node *Tail;
	int   Length;
} DoublyList;

void  DoublyList_Free(DoublyList *list);
Node *DoublyList_GetNode(const DoublyList *list, int index);
void  DoublyList_InsertAt(DoublyList *list, int index, int val);
void  DoublyList_RemoveAt(DoublyList *list, int index);
void  DoublyList_InsertHead(DoublyList *list, int val);
void  DoublyList_InsertTail(DoublyList *list, int val);
void  DoublyList_RemoveHead(DoublyList *list);
void  DoublyList_RemoveTail(DoublyList *list);
int   DoublyList_Get(const DoublyList *list, int index);

/* ------------------------------------------------------------------ */
/* 06_queue_stack                                                      */
/* ------------------------------------------------------------------ */

typedef struct {
	int   Length;
	Node *Head;
	Node *Tail;
} Queue;

void Queue_Free(Queue *queue);
void Queue_Enqueue(Queue *queue, int val);
int  Queue_Deque(Queue *queue);
int  Queue_Peek(const Queue *queue);

typedef struct {
	int   Length;
	Node *Head;
} Stack;

void Stack_Free(Stack *stack);
void Stack_Push(Stack *stack, int val);
int  Stack_Pop(Stack *stack);
int  Stack_Peek(const Stack *stack);

/* ------------------------------------------------------------------ */
/* 07_ringbuffer                                                       */
/* ------------------------------------------------------------------ */

typedef struct RingBuffer RingBuffer;

RingBuffer *RingBuffer_New(int capacity);
void RingBuffer_Free(RingBuffer *rb);
void RingBuffer_Enqueue(RingBuffer *rb, int val);
int  RingBuffer_Dequeue(RingBuffer *rb);
int  RingBuffer_Peek(const RingBuffer *rb);

/* ------------------------------------------------------------------ */
/* 08_heap - a Go slice becomes a buffer + length + capacity           */
/* ------------------------------------------------------------------ */

typedef struct {
	int *array;
	int  length;
	int  capacity;
} MinHeap; /* zero initialise it: MinHeap h = {0}; */

void MinHeap_Free(MinHeap *heap);
void MinHeap_Insert(MinHeap *heap, int val);
int  MinHeap_Delete(MinHeap *heap);

typedef struct {
	int *array;
	int  length;
	int  capacity;
} MaxHeap;

void MaxHeap_Free(MaxHeap *heap);
void MaxHeap_Insert(MaxHeap *heap, int val);
int  MaxHeap_Delete(MaxHeap *heap);

/* ------------------------------------------------------------------ */
/* 09_recursion                                                        */
/* ------------------------------------------------------------------ */

int RecursiveSum(int n);

typedef struct {
	int x;
	int y;
} Point;

Point *MazeSolver(const char **maze, int rows, char wall, Point start, Point end, int *outLen);

/* ------------------------------------------------------------------ */
/* 10_tree                                                             */
/* ------------------------------------------------------------------ */

typedef struct TreeNode TreeNode;
struct TreeNode {
	int        Value;
	TreeNode **Children; /* the []*TreeNode slice, grown as needed */
	int        ChildCount;
	int        ChildCapacity;
	TreeNode  *Parent;
};

TreeNode *TreeNode_New(int value);
void TreeNode_Free(TreeNode *node); /* frees the whole subtree */
void TreeNode_Insert(TreeNode *node, TreeNode *child);
void TreeNode_Delete(TreeNode *node); /* detaches from the parent (does not free), never the root */
int *TreeNode_PreOrderSearch(const TreeNode *node, int *outLen);
int *TreeNode_PostOrderSearch(const TreeNode *node, int *outLen);

/* ------------------------------------------------------------------ */
/* 11_binary_tree                                                      */
/* ------------------------------------------------------------------ */

typedef struct BinaryTreeNode BinaryTreeNode;
struct BinaryTreeNode {
	int             Value;
	BinaryTreeNode *Left;
	BinaryTreeNode *Right;
	BinaryTreeNode *Parent;
};

BinaryTreeNode *BinaryTreeNode_New(int value);
void BinaryTreeNode_Free(BinaryTreeNode *node); /* frees the whole subtree */
void BinaryTreeNode_Insert(BinaryTreeNode *node, BinaryTreeNode *child);
int *BinaryTreeNode_PreOrderSearch(const BinaryTreeNode *node, int *outLen);
int *BinaryTreeNode_InOrderSearch(const BinaryTreeNode *node, int *outLen);
int *BinaryTreeNode_PostOrderSearch(const BinaryTreeNode *node, int *outLen);
bool BinaryTreeNode_BreadthFirstSearch(const BinaryTreeNode *node, int needle);
bool CompareBinaryTree(const BinaryTreeNode *a, const BinaryTreeNode *b);

/* ------------------------------------------------------------------ */
/* 12_binary_search_tree                                               */
/* ------------------------------------------------------------------ */

typedef struct BSTNode BSTNode;
struct BSTNode {
	int      Value;
	BSTNode *Left;
	BSTNode *Right;
	BSTNode *Parent;
};

BSTNode *BSTNode_New(int value);
void BSTNode_Free(BSTNode *node); /* frees the whole subtree */
BSTNode *BSTNode_Find(BSTNode *node, int val); /* NULL when not found */
void BSTNode_Insert(BSTNode *node, int val);
BSTNode *BSTNode_FindMin(BSTNode *node);
void BSTNode_Delete(BSTNode *node);

/* ------------------------------------------------------------------ */
/* 13_trie_tree                                                        */
/* ------------------------------------------------------------------ */

typedef struct TrieNode TrieNode;

TrieNode *TrieNode_New(void);
void TrieNode_Free(TrieNode *node); /* frees the whole trie */
void TrieNode_Insert(TrieNode *node, const char *word);
bool TrieNode_Search(const TrieNode *node, const char *word);
bool TrieNode_StartsWith(const TrieNode *node, const char *prefix);
bool TrieNode_Delete(TrieNode *node, const char *word);

/* ------------------------------------------------------------------ */
/* 14_graph - a [][]T slice becomes rows + per row counts + size       */
/* ------------------------------------------------------------------ */

typedef struct {
	int From;
	int To;
	int Weight;
} CompleteGraphEdge;

typedef struct {
	int To;
	int Weight;
} GraphEdge;

typedef struct {
	int **Matrix; /* Matrix[from][to], a number means weight */
	int   Size;   /* vertices, so every row is Size long */
} WeightedAdjMatrix;

typedef struct {
	GraphEdge **List;   /* List[from] is an array of Counts[from] edges */
	int        *Counts;
	int         Size;
} WeightedAdjList;

typedef struct {
	int **List;
	int  *Counts;
	int   Size;
} AdjList;

typedef struct {
	int **Matrix; /* a 1 means connected */
	int   Size;
} AdjMatrix;

int *WeightedAdjMatrix_BFS(const WeightedAdjMatrix *graph, int source, int needle, int *outLen);
int *WeightedAdjList_DFS(const WeightedAdjList *graph, int source, int needle, int *outLen);
int *WeightedAdjList_Dijkstra(const WeightedAdjList *graph, int source, int needle, int *outLen);

/* ------------------------------------------------------------------ */
/* 15_lru_cache                                                        */
/* ------------------------------------------------------------------ */

typedef struct LRUCache LRUCache;

LRUCache *LRUCache_New(int capacity);
void LRUCache_Free(LRUCache *cache);
bool LRUCache_Get(LRUCache *cache, const char *key, int *outValue);
void LRUCache_Put(LRUCache *cache, const char *key, int value);

#endif /* ALGORITHMS_H */
