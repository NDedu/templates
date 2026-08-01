package main

// BinaryTree - branching factor 2

type BinaryTreeNode struct {
	Value	int
	Left	*BinaryTreeNode
	Right	*BinaryTreeNode
	Parent	*BinaryTreeNode
}

// Insert at target node if it has space for children (from left to right)
func (node *BinaryTreeNode) Insert(child *BinaryTreeNode) {

	if child == nil {
		return
	}

	if node.Left == nil {

		node.Left = child
		child.Parent = node

	} else if node.Right == nil {

		node.Right = child
		child.Parent = node

	} else {

		panic("node has no space for children")
	}
}

// Traversal DFS

// Pre-Order Search: visit node -> recurse left -> recurse right. gets root node first
func binaryPreWalk(curr *BinaryTreeNode, path []int) []int {

	if curr == nil {
		return path
	}

	path = append(path, curr.Value)

	path = binaryPreWalk(curr.Left, path)
	path = binaryPreWalk(curr.Right, path)

	return path
}

func (node *BinaryTreeNode) PreOrderSearch() []int {
	return binaryPreWalk(node, []int{})
}

// In-Order Search: recurse left -> visit node -> recurse right. root node found in the middle (when all values printed)
func binaryInWalk(curr *BinaryTreeNode, path []int) []int {

	if curr == nil {
		return path
	}

	path = binaryInWalk(curr.Left, path)
	path = append(path, curr.Value)
	path = binaryInWalk(curr.Right, path)

	return path
}

func (node *BinaryTreeNode) InOrderSearch() []int {
	return binaryInWalk(node, []int{})
}

// Post-Order Search: recurse left -> recurse right -> visit node. root node found at the end
func binaryPostWalk(curr *BinaryTreeNode, path []int) []int {

	if curr == nil {
		return path
	}

	path = binaryPostWalk(curr.Left, path)
	path = binaryPostWalk(curr.Right, path)
	path = append(path, curr.Value)

	return path
}

func (node *BinaryTreeNode) PostOrderSearch() []int {
	return binaryPostWalk(node, []int{})
}

// Breadth First Search - gets values (from left to right) from each tree level from (nodes that are at the same "height")
// Visit root add it's children to the queue, print the root. Visit the first children (next value in queue after root), add it's childrento queue, then pop it; and so on.

func (node *BinaryTreeNode) BreadthFirstSearch(needle int) bool {

	if node == nil {
		return false
	}

	queue := []*BinaryTreeNode{node} // A Queue data structure would work better, but mine is on Node not BinaryTreeNode

	for len(queue) > 0 {

		curr := queue[0]
		queue = queue[1:] // Deque

		if curr.Value == needle {
			return true
		}

		// Enqueue
		if curr.Left != nil {
			queue = append(queue, curr.Left)
		}
		if curr.Right != nil {
			queue = append(queue, curr.Right)
		}
	}

	return false
}

func CompareBinaryTree(a, b *BinaryTreeNode) bool {

	// Can't recurse anymore
	// Structural check
	if a == nil && b == nil {
		return true
	}

	// Because of the above, one is not nil
	// Structural check
	if a == nil || b == nil {
		return false
	}

	// Value check
	if a.Value != b.Value {
		return false
	}

	return CompareBinaryTree(a.Left, b.Left) && CompareBinaryTree(a.Right, b.Right)
}
