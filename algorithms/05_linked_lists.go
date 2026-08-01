package main

// Linked Lists - a Node type data structure: A <-> B <-> C

type Node struct {
	Value	int
	Next	*Node
	Prev	*Node // not part of Singly Linked List but used for Doubly
}

func InsertionLinkedList(node *Node, newNode *Node) {

	newNode.Next = node.Next
	node.Next = newNode
}

// DeletionLinkedList deletes the node after the parameter node
func DeletionLinkedList(node *Node) {

	node.Next = node.Next.Next
}

// Singly Linked List operations working directly with Head are O(1); traversal O(n) but we can traverse only forward (only with Next)

type SinglyList struct {
	Head	*Node
	Length	int
}

func (list *SinglyList) InsertAt(index int, val int) {

	if index < 0 || index > list.Length {
		panic("index out of bounds")
	}

	// Insertion at Head
	if index == 0 {
		newNode := &Node{ Value: val }
		newNode.Next = list.Head
		list.Head = newNode
		list.Length++
		return
	}

	// Traverse the list to reach index position
	current := list.Head
	for i := 0; i < index - 1; i++ {
		current = current.Next
	}

	newNode := &Node{ Value: val }
	newNode.Next = current.Next
	current.Next = newNode
	list.Length++
}

func (list *SinglyList) Remove() {

	if list.Length == 0 {
		panic("empty list")
	}

	if list.Length == 1 {
		list.Head = nil
		list.Length = 0
		return
	}

	current := list.Head
	for range (list.Length - 2) {
		current = current.Next
	}
	
	current.Next = nil
	list.Length--
}

func (list *SinglyList) RemoveAt(index int) {

	if index < 0 || index >= list.Length {
		panic("index out of bounds")
	}

	if index == 0 {
		list.Head = list.Head.Next
		list.Length--
		return
	}

	if index == list.Length - 1 {
		list.Remove()
		return
	}

	current := list.Head
	for i := 0; i < index - 1; i++ {
		current = current.Next
	}

	current.Next = current.Next.Next
	list.Length--
}

func (list *SinglyList) Append(val int) {

	newNode := &Node{ Value: val }

	if list.Head == nil {
		list.Head = newNode
		list.Length++
		return
	}

	current := list.Head
	for current.Next != nil {
		current = current.Next
	}

	current.Next = newNode
	list.Length++
}

func (list *SinglyList) Prepend(val int) {

	newNode := &Node{ Value: val }
	newNode.Next = list.Head
	list.Head = newNode
	list.Length++
}

func (list *SinglyList) Get(index int) int {

	if index < 0 || index >= list.Length {
		panic("index out of bounds")
	}

	current := list.Head
	for i := 0; i < index; i++ {
		current = current.Next
	}

	return current.Value
}

// Doubly Linked List - operations working directly with Head/Tail are O(1); traversal O(n) but compared to singly we can traverse bidirectional (with Next or Prev)

type DoublyList struct {
	Head	*Node
	Tail	*Node
	Length	int
}

func (list *DoublyList) GetNode(index int) *Node {

	// This works for edge cases where index is at tail or head with O(1)
	// Bidirectional traversal choosing the most eficient path
	if index <= list.Length / 2 {

		current := list.Head
		for i := 0; i < index; i++ {
			current = current.Next
		}

		return current
	
	} else {

		current := list.Tail
		for i := list.Length - 1; i > index; i-- {
			current = current.Prev
		}
		
		return current
	}
}

func (list *DoublyList) InsertAt(index int, val int) {

	if index < 0 || index > list.Length {
		panic("index out of bounds")
	}

	// Insertion at Head
	if index == 0 {
		list.InsertHead(val)
		return
	}

	// Insertion at Tail
	if index == list.Length {
		list.InsertTail(val)
		return
	}
	
	node := list.GetNode(index)
	newNode := &Node{ Value: val }

	newNode.Prev = node.Prev
	newNode.Next = node
	node.Prev.Next = newNode
	node.Prev = newNode
	list.Length++
}

func (list *DoublyList) RemoveAt(index int) {

	if index < 0 || index >= list.Length {
		panic("index out of bounds")
	}

	if index == 0 {
		list.RemoveHead()
		return
	}

	if index == list.Length - 1 {
		list.RemoveTail()
		return
	}

	node := list.GetNode(index)
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
	list.Length--
}

func (list *DoublyList) InsertHead(val int) {

	newNode := &Node{ Value: val }

	if list.Length == 0 {
		list.Head = newNode
		list.Tail = newNode
		list.Length++
		return
	}

	newNode.Next = list.Head
	list.Head.Prev = newNode
	list.Head = newNode
	list.Length++
}

func (list *DoublyList) InsertTail(val int) {

	newNode := &Node{ Value: val }

	if list.Head == nil || list.Tail == nil {
		list.Head = newNode
		list.Tail = newNode
		list.Length++
		return
	}

	newNode.Prev = list.Tail
	list.Tail.Next = newNode
	list.Tail = newNode
	list.Length++
}

func (list *DoublyList) RemoveHead() {

	if list.Length == 0 {
		panic("empty list")
	}

	if list.Length == 1 {
		list.Head = nil
		list.Tail = nil
		list.Length = 0
		return
	}
	
	list.Head = list.Head.Next
	list.Head.Prev = nil
	list.Length--
}

func (list *DoublyList) RemoveTail() {

	if list.Length == 0 {
		panic("empty list")
	}

	if list.Length == 1 {
		list.Head = nil
		list.Tail = nil
		list.Length = 0
		return
	}
	
	list.Tail = list.Tail.Prev
	list.Tail.Next = nil
	list.Length--
}

func (list *DoublyList) Get(index int) int {

	if index < 0 || index >= list.Length {
		panic("index out of bounds")
	}
	
	node := list.GetNode(index)
	return node.Value
}
