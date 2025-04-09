package utils

// AbstractNode represents an abstract node in a graph.
type AbstractNode interface {
	String() string
	// EqualsTo checks if the current node is equal to another node.
	EqualsTo(other AbstractNode) bool
	// HashCode returns a hash code for the node.
	HashCode() uint64
}

// AbstractGraph represents an abstract graph structure.
type AbstractGraph interface {
	// HasNode checks if a node exists in the graph.
	HasNode(node AbstractNode) bool
	// GetNeighborsOf retrieves the neighbors of a given node.
	GetNeighborsOf(node AbstractNode) []AbstractNode
}

// CanReach determines if the target node can be reached from the 'from' node in the given graph.
func CanReach(graph AbstractGraph, from, target AbstractNode) bool {
	visited := make(map[uint64]bool)
	return dfs(graph, from, target, visited)
}

// dfs is a helper function for depth-first search.
func dfs(graph AbstractGraph, current, target AbstractNode, visited map[uint64]bool) bool {
	if current.EqualsTo(target) {
		return true
	}
	if visited[current.HashCode()] {
		return false
	}

	visited[current.HashCode()] = true
	for _, neighbor := range graph.GetNeighborsOf(current) {
		if dfs(graph, neighbor, target, visited) {
			return true
		}
	}
	return false
}
