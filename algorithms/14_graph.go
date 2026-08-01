package main

import "math"

// A series of nodes, with some amount of connections (there are no rules). Can be represented as Adj List (memory efficient, speed ineff)/Adj Matrix (Adjacency)
// Terminology: cycle (visit some nodes and come back to starting node), connected (every node can reach every other node), directed (direction of connections), weighted (connections have a values)
// dag: directed acyclic graph; node - vertex/point; edge - connection between 2  nodes

type CompleteGraphEdge struct {
	From	int
	To		int
	Weight	int
}

type GraphEdge struct {
	To		int
	Weight	int
}

type WeightedAdjList struct {
	List	[][]GraphEdge
}

type WeightedAdjMatrix struct {
	Matrix	[][]int // A number means weight
}

type AdjList struct {
	List	[][]int
}

type AdjMatrix struct {
	Matrix	[][]int // A 1 means connected
}


// BFS - O(vertex^2) now but O(v + e) if adj list
func (graph *WeightedAdjMatrix) BFS(source, needle int) []int {

	if len(graph.Matrix) == 0 {
		return nil
	}

	seen := make([]bool, len(graph.Matrix)) // Filled with false, have not seen anything yet
	prev := make([]int, len(graph.Matrix))

	for i := range prev {
		prev[i] = -1 // No parents yet
	}

	seen[source] = true // Seen source because queue contains source
	queue := []int{source} // Can use a proper Queue struct here

	for len(queue) > 0 {

		// Dequeu
		curr := queue[0]
		if curr == needle {
			break
		}
		queue = queue[1:]

		for neighbor, weight := range graph.Matrix[curr] {
			
			// If there is an edge and we haven't visited it yet
			if weight > 0 && !seen[neighbor] {
				seen[neighbor] = true
				prev[neighbor] = curr
				queue = append(queue, neighbor)
			}
		}
	}

	// Needle not found
	if prev[needle] == -1 && source != needle {
		return nil
	}

	// Build/walk backward from needle to source
	var path []int
	curr := needle
	for curr != -1 {
		path = append(path, curr)
		curr = prev[curr]
	}

	// The path is currently [needle, ..., source], needs to be reversed
	for i := 0; i < len(path)/2; i++ {
		j := len(path) - 1 - i
		temp := path[i]
		path[i] = path[j]
		path[j] = temp
	}

	return path
}

// Walk function for recursion
func walkDFS(graph *WeightedAdjList, curr, needle int, seen []bool, path []int) []int {

	if seen[curr] {
		return nil
	}

	seen[curr] = true
	path = append(path, curr)

	// Found it, return the path
	if curr == needle {
		return path
	}

	for _, edge := range graph.List[curr] {

		// Walk the bath and if needle found, return the path
		if result := walkDFS(graph, edge.To, needle, seen, path); result != nil {
			return result
		}
	}

	return nil
}

// DFS - O(vertex + edges that exist)
func (graph *WeightedAdjList) DFS(source, needle int) []int {

	if len(graph.List) == 0 {
		return nil
	}
	
	seen := make([]bool, len(graph.List))
	var path []int

	return walkDFS(graph, source, needle, seen, path)
}

func hasUnvisited(seen []bool, dists []int) bool {

	for i, s := range seen {

		// Visit all seen and check if false and distance < infinity
		if !s && dists[i] < math.MaxInt32 {
			return true
		}
	}
	return false
}

// Return lowest unvisited index
func getLowestUnvisited(seen []bool, dists []int) int {

	idx := -1
	lowestDistance := math.MaxInt32

	for i := range len(seen) {

		if !seen[i] && dists[i] < lowestDistance {
			lowestDistance = dists[i]
			idx = i
		}
	}
	return idx
}

// Dijkstra - Shortest path, O(v^2) - but can be O((v + e) log v) if getLowestUnvisited changed to Priority Queue (MinHeap)
func (graph *WeightedAdjList) Dijkstra(source, needle int) []int {

	seen := make([]bool, len(graph.List))
	prev := make([]int, len(graph.List)) // To know way back
	dists := make([]int, len(graph.List)) // Array with distances

	// Fill the previous with no parents, and distances with "infinity" (nodes get max distance until visited)
	for i := 0; i < len(graph.List); i++ {
		prev[i] = -1
		dists[i] = math.MaxInt32
	}

	// Distance of the starting point
	dists[source] = 0

	// Always checks for shortest distance
	for hasUnvisited(seen, dists) {

		curr := getLowestUnvisited(seen, dists)
		if curr == -1 {
			break
		}
		
		seen[curr] = true

		// Check all edges
		for _, edge := range graph.List[curr] {

			if seen[edge.To] {
				continue
			}
			
			// Calculate distance and if it's better we push it
			dist := dists[curr] + edge.Weight
			if dist < dists[edge.To] {
				dists[edge.To] = dist
				prev[edge.To] = curr
			}
		}
	}

	// Walk distance backwards

	if prev[needle] == -1 && source != needle {
		return nil
	}

	var path []int
	curr := needle
	for curr != -1 {

		path = append(path, curr)
		curr = prev[curr]
	}

	for i := 0; i < len(path)/2; i++ {

		j := len(path) - 1 - i
		temp := path[i]
		path[i] = path[j]
		path[j] = temp
	}

	return path
}
