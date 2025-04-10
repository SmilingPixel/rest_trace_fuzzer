// Package utils provides graph-related utilities using abstract interfaces.
// 
// To use the related graph algorithm, you must implement the following interfaces:
//
// - AbstractNode:
//     - EqualsTo(other AbstractNode) bool
//     - ID() string
//
// - AbstractGraph:
//     - HasNode(node AbstractNode) bool
//     - GetNeighborsOf(node AbstractNode) []AbstractNode
//
// The CanReach function allows you to determine whether one node can reach another
// within a given graph by using depth-first search.
// TODO: rewrite it using Golang generics @xunzhou24
package utils

// AbstractNode represents a node in a graph.
//
// Types that implement this interface must provide methods to compare equality,
// and return a unique identifier.
type AbstractNode interface {

	// ID returns a unique identifier for the node.
	ID() string
}

// AbstractGraph represents a graph structure composed of nodes.
//
// Types that implement this interface must provide methods to check for
// node existence and retrieve neighbor nodes.
type AbstractGraph interface {
	// HasNode returns true if the node exists in the graph.
	HasNode(node AbstractNode) bool

	// GetNeighborsOf returns a slice of neighboring nodes connected to the given node.
	GetNeighborsOf(node AbstractNode) []AbstractNode
}

// CanReach determines if the target node can be reached from the 'from' node
// in the given graph.
//
// The function performs a depth-first search and returns true if a path
// exists from 'from' to 'target'; otherwise, it returns false.
func CanReach(graph AbstractGraph, from, target AbstractNode) bool {
	visited := make(map[string]bool)
	return dfs(graph, from, target, visited)
}

// dfs is a recursive helper function that performs depth-first search
// to determine reachability between two nodes in a graph.
//
// It returns true if the target node can be reached from the current node.
func dfs(graph AbstractGraph, current, target AbstractNode, visited map[string]bool) bool {
	if current.ID() == target.ID() {
		return true
	}
	if visited[current.ID()] {
		return false
	}

	visited[current.ID()] = true
	for _, neighbor := range graph.GetNeighborsOf(current) {
		if dfs(graph, neighbor, target, visited) {
			return true
		}
	}
	return false
}
