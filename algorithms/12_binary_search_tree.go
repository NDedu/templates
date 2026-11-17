package main

// Binary Search Tree (BST) - ordered/sorted binary tree, and no duplicates
// Everything to the right (with respective to each node) is bigger than left
// Everything to the left of the root is smaller than the root
// Everything (meaning children of the left children too) to the left of a node is smaller than the node
// When searching something you can compare to the root and then to nodes so you don't traverse the entire tree (binary search)

type BSTNode struct {
	Value  int
	Left   *BSTNode
	Right  *BSTNode
	Parent *BSTNode
}

// BST - holds the root, so deleting the root node (or the last node, leaving an empty tree) can update it
type BST struct {
	Root *BSTNode
}

// Find - O(height)
func (tree *BST) Find(val int) *BSTNode {
	return tree.Root.find(val)
}

func (node *BSTNode) find(val int) *BSTNode {

	if node == nil {
		return nil
	}

	if val == node.Value {
		return node
	} else if val < node.Value {
		return node.Left.find(val)
	} else {
		return node.Right.find(val)
	}
}

// Insert - O(h)
func (tree *BST) Insert(val int) {

	if tree.Root == nil {
		tree.Root = &BSTNode{Value: val}
		return
	}

	tree.Root.insert(val)
}

func (node *BSTNode) insert(val int) {

	// Duplicate check
	if val == node.Value {
		return
	}

	if val < node.Value {

		if node.Left == nil {
			node.Left = &BSTNode{Value: val, Parent: node}
		} else {
			node.Left.insert(val)
		}

	} else {

		if node.Right == nil {
			node.Right = &BSTNode{Value: val, Parent: node}
		} else {
			node.Right.insert(val)
		}
	}
}

// FindMin - goes left, nil for an empty tree
func (tree *BST) FindMin() *BSTNode {

	if tree.Root == nil {
		return nil
	}

	return tree.Root.findMin()
}

func (node *BSTNode) findMin() *BSTNode {

	curr := node

	for curr.Left != nil {
		curr = curr.Left
	}

	return curr
}

// Delete - O(h), returns false if val is not in the tree
func (tree *BST) Delete(val int) bool {

	node := tree.Find(val)
	if node == nil {
		return false
	}

	tree.deleteNode(node)
	return true
}

func (tree *BST) deleteNode(node *BSTNode) {

	// Node has 2 children
	if node.Left != nil && node.Right != nil {

		successor := node.Right.findMin()

		// Swap values
		node.Value = successor.Value

		// Call delete to go to case where it has 0/1 children
		tree.deleteNode(successor)
		return
	}

	// Node has 0/1 children
	var child *BSTNode
	if node.Left != nil {
		child = node.Left
	} else if node.Right != nil {
		child = node.Right
	}

	// Update parent to point to child
	// The root has no parent: the tree's root pointer moves to the child instead (nil when it was the last node)
	if node.Parent == nil {
		tree.Root = child
	} else if node.Parent.Left == node {
		node.Parent.Left = child
	} else {
		node.Parent.Right = child
	}

	if child != nil {
		child.Parent = node.Parent
	}

	// Cleanup
	node.Parent = nil
	node.Left = nil
	node.Right = nil
}
