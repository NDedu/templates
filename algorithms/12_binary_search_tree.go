package main

// Binary Search Tree (BST) - ordered/sorted binary tree, and no duplicates
// Everything to the right (with respective to each node) is bigger than left
// Everything to the left of the root is smaller than the root
// Everything (meaning children of the left children too) to the left of a node is smaller than the node
// When searching something you can compare to the root and then to nodes so you don't traverse the entire tree (binary search)

type BSTNode struct {
	Value	int
	Left	*BSTNode
	Right	*BSTNode
	Parent	*BSTNode
}

// Find - O(height)
func (node *BSTNode) Find(val int) *BSTNode {

	if node == nil {
		return nil
	}

	if val == node.Value {
		return node
	} else if val < node.Value {
		return node.Left.Find(val)
	} else {
		return node.Right.Find(val)
	}
}

// Insert - O(h)
func (node *BSTNode) Insert(val int) {

	// Duplicate check
	if val == node.Value {
		return
	}

	if val < node.Value {

		if node.Left == nil {
			node.Left = &BSTNode{ Value: val, Parent: node }
		} else {
			node.Left.Insert(val)
		}

	} else {

		if node.Right == nil {
			node.Right = &BSTNode{ Value: val, Parent: node }
		} else {
			node.Right.Insert(val)
		}
	}
}

// FindMin - goes left
func (node *BSTNode) FindMin() *BSTNode {

	curr := node

	for curr.Left != nil {
		curr = curr.Left
	}

	return curr
}

// Delete - O(h)
func (node *BSTNode) Delete() {

	if node == nil {
		return
	}

	// Node has 2 children
	if node.Left != nil && node.Right != nil {

		successor := node.Right.FindMin()
		
		// Swap values
		node.Value = successor.Value
		
		// Call delete to go to case where it has 0/1 children
		successor.Delete()
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
	if node.Parent != nil {

		if node.Parent.Left == node {
			node.Parent.Left = child
		} else {
			node.Parent.Right = child
		}
	}

	if child != nil {
		child.Parent = node.Parent
	}

	// Cleanup
	node.Parent = nil
	node.Left = nil
	node.Right = nil
}
