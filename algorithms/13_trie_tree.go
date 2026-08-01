package main

// Trie - like an autocomplete with O(length of word), where each node is a character and it's children construct a word
// IsWord flag marks complete word
// Words share prefixes (cat and card share the ca path)
//
//          root
//        /      \
//       c        m
//      /          \
//     a            a
//    / \            \
//   t   r            r
//  / \   \            \
// s   t   d            c
//      \
//       t
//        \
//         l
//          \
//           e

type TrieNode struct {
	Children	map[rune]*TrieNode
	IsWord		bool
}

func NewTrieNode() *TrieNode {
	return &TrieNode{ Children: map[rune]*TrieNode{} }
}

// Insert - O(l)
func (node *TrieNode) Insert(word string) {

	curr := node

	for _, char := range word {

		next, ok := curr.Children[char]
		if !ok {
			next = NewTrieNode()
			curr.Children[char] = next
		}

		curr = next
	}

	curr.IsWord = true
}

// Search - O(l) exact word match
func (node *TrieNode) Search(word string) bool {

	curr := node

	for _, char := range word {

		next, ok := curr.Children[char]
		if !ok {
			return false
		}

		curr = next
	}

	return curr.IsWord
}

// StartsWith - O(l) any word with prefix
func (node *TrieNode) StartsWith(prefix string) bool {

	curr := node

	for _, char := range prefix {

		next, ok := curr.Children[char]
		if !ok {
			return false
		}

		curr = next
	}

	return true
}

// Delete - O(l) - deletes the word and any nodes that are no longer part of another word
func (node *TrieNode) Delete(word string) bool {

	if !node.Search(word) {
		return false
	}

	deleteWalk(node, word, 0)
	return true
}

// deleteWalk returns true if the parent should remove this child
func deleteWalk(curr *TrieNode, word string, depth int) bool {

	if curr == nil {
		return false
	}

	// End of word reached
	if depth == len(word) {

		// Word was not in the trie
		if !curr.IsWord {
			return false
		}

		curr.IsWord = false

		// Delete if no other word branches from here
		return len(curr.Children) == 0
	}

	char := rune(word[depth])
	child, ok := curr.Children[char]
	if !ok {
		return false
	}

	shouldDeleteNodes := deleteWalk(child, word, depth + 1)

	if shouldDeleteNodes {

		delete(curr.Children, char)
		return !curr.IsWord && len(curr.Children) == 0
	}

	return false
}
