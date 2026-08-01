package main

// A function that calls itself until a condition
// Used when not able to use for loops, like when is a branching factor (for ex. the maze solver has a 4 branching factor for direction traversal)

// RecursiveSum - kills the stack for big numbers (stack overflow) as it has O(n) space complexity (each function called, even recursive, lives in the stack) (for loop sum has O(1) space)
func RecursiveSum(n int) int {

	// Base case
	if n == 1 {
		return 1
	}

	return n + RecursiveSum(n - 1) // returns n + (n - 1) + ((n - 1) - 1) + (((n - 1) - 1) - 1) + ... + 1
}

// Maze Solver

type Point struct {
	x	int
	y	int
}

func MazeSolver(maze []string, wall string, start, end Point) []Point {

	seen := make([][]bool, len(maze))
	for i := range seen {
		seen[i] = make([]bool, len(maze[0]))
	}

	path := []Point{}

	if walk(maze, wall, start, end, &path, seen) {
		return path
	}

	return nil
}

var dir = [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

func walk(maze []string, wall string, curr, end Point, path *[]Point, seen [][]bool) bool {

	// Base case 1: off the map (outside of maze)
	if curr.x < 0 || curr.x >= len(maze[0]) || curr.y < 0 || curr.y >= len(maze) {
		return false
	}

	// Base case 2: hit the wall
	if string(maze[curr.y][curr.x]) == wall {
		return false
	}

	// Base case 3: seen the position
	if seen[curr.y][curr.x] {
		return false
	}

	// Base case 4: reached the end
	if curr.x == end.x && curr.y == end.y {
		*path = append(*path, curr)
		return true
	}

	// recursion steps: pre condition -> recurse -> post condition
	// pre
	*path = append(*path, curr)
	seen[curr.y][curr.x] = true

	// recurse
	for i := range len(dir) {

		x := dir[i][0]
		y := dir[i][1]

		newCurr := Point{curr.x + x, curr.y + y}

		if walk(maze, wall, newCurr, end, path, seen) {
			return true
		}
	}

	// post
	*path = (*path)[:len(*path) - 1]

	return false
}
