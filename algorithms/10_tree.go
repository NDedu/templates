package main

// A Node data structure with childrens and parent
// Terminology: root/head, leaves, height, branching factor, balanced tree

type TreeNode struct {
	Value		int
	Children	[]*TreeNode
	Parent		*TreeNode
}

func (node *TreeNode) Insert(child *TreeNode) {

	if child == nil {
		return
	}

	child.Parent = node
	node.Children = append(node.Children, child)
}

func (node *TreeNode) Delete() {

	// Don't delete root
	if node == nil || node.Parent == nil {
		return
	}

	parent := node.Parent

	// Find this exact node in the parent's children slice
	for i, child := range parent.Children {

		if child == node {

			node.Parent = nil

			// Shift elements
			copy(parent.Children[i:], parent.Children[i+1:])

			// Memory cleanup
			parent.Children[len(parent.Children) - 1] = nil

			// Decrease the children
			parent.Children = parent.Children[:len(parent.Children)-1]

			return
		}
	}
}

// Traversal (Depth First Search/Traversal - DFS) - DFS preserve shape, while BFS does not

// Pre-Order Search: visit node -> recurse left -> recurse right. gets root node first
func preWalk(curr *TreeNode, path []int) []int {

	if curr == nil {
		return path
	}

	path = append(path, curr.Value)

	for _, child := range curr.Children {
		path = preWalk(child, path)
	}

	return path
}

func (node *TreeNode) PreOrderSearch() []int {
	return preWalk(node, []int{})
}

// In-Order Search: recurse left -> visit node -> recurse right. root node found in the middle (when all values printed)
// for general tree there is no in-order search

// Post-Order Search: recurse left -> recurse right -> visit node. root node found at the end
func postWalk(curr *TreeNode, path []int) []int {

	if curr == nil {
		return path
	}

	for _, child := range curr.Children {
		path = postWalk(child, path)
	}

	path = append(path, curr.Value)

	return path
}

func (node *TreeNode) PostOrderSearch() []int {
	return postWalk(node, []int{})
}
