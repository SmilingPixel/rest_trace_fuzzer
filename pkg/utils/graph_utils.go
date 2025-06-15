// Package utils provides generic graph structures and algorithms using Go generics.
//
// To use this package, you must define your own node type that satisfies the `AbstractNode` interface (i.e., is `comparable`),
// and your own edge type that implements the `AbstractEdge[N]` interface.
//
// You can then instantiate a graph using `NewGraph[MyNode, MyEdge]()` and use methods like `AddEdge`, `HasNode`, and `CanReach`.
package utils

// AbstractNode defines a node type that can be used in a graph.
// It must be comparable so it can serve as a map key and support equality checks.
type AbstractNode interface {
	comparable
}

// AbstractEdge represents a directional connection between two nodes of type N.
//
// Types implementing this interface must define methods to retrieve the source and target nodes.
type AbstractEdge[N AbstractNode] interface {
	// GetSource returns the source node of the edge.
	GetSource() N

	// GetTarget returns the target node of the edge.
	GetTarget() N
}

// Graph represents a directed graph with nodes of type N and edges of type E.
//
// It maintains both an edge list and an adjacency list for efficient traversal and storage.
type Graph[N AbstractNode, E AbstractEdge[N]] struct {
	// Edges stores all edges in the graph.
	Edges []E `json:"edges"`

	// AdjacencyList maps each node to a list of outgoing edges.
	// This field is not serialized to JSON, as it duplicates the information in Edges.
	AdjacencyList map[N][]E `json:"-"`
}

// NewGraph creates and returns a new empty Graph.
func NewGraph[N AbstractNode, E AbstractEdge[N]]() *Graph[N, E] {
	g := &Graph[N, E]{
		Edges:         make([]E, 0),
		AdjacencyList: make(map[N][]E),
	}
	return g
}

// AddEdge adds an edge to the graph and updates the adjacency list.
//
// The edge is appended to the edge list and also inserted into the adjacency list
// under its source node.
func (g *Graph[N, E]) AddEdge(edge E) {
	g.Edges = append(g.Edges, edge)
	source := edge.GetSource()
	g.AdjacencyList[source] = append(g.AdjacencyList[source], edge)

	// If target does not have an entry in adjacency list, add a corresponding key,
	// for `HasNode` check the adjacency list to determine if a node exists.
	target := edge.GetTarget()
	if _, exists := g.AdjacencyList[target]; !exists {
		g.AdjacencyList[target] = []E{}
	}
}

// HasNode checks whether the graph contains the specified node.
//
// It returns true if the node exists in the graph’s adjacency list.
func (g *Graph[N, E]) HasNode(node N) bool {
	_, exists := g.AdjacencyList[node]
	return exists
}

// GetAllNodes returns a slice of all nodes in the graph.
// It iterates over the adjacency list and collects all unique nodes.
func (g *Graph[N, E]) GetAllNodes() []N {
	nodes := make(map[N]struct{})
	for node := range g.AdjacencyList {
		nodes[node] = struct{}{}
		for _, edge := range g.AdjacencyList[node] {
			nodes[edge.GetTarget()] = struct{}{}
		}
	}
	allNodes := make([]N, 0, len(nodes))
	for node := range nodes {
		allNodes = append(allNodes, node)
	}
	return allNodes
}

// CanReach determines whether there is a path from the source node to the target node.
//
// It performs a breadth-first search (BFS) starting from the source. Returns true if the
// target node is reachable; otherwise, returns false.
func (g *Graph[N, E]) CanReach(source, target N) bool {
	if !g.HasNode(source) || !g.HasNode(target) {
		return false
	}

	visited := make(map[N]bool)
	queue := []N{source}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node == target {
			return true
		}

		if visited[node] {
			continue
		}
		visited[node] = true

		for _, edge := range g.AdjacencyList[node] {
			queue = append(queue, edge.GetTarget())
		}
	}

	return false
}

// GetDistanceMapBySource returns the distance from the source node to all other nodes in the graph.
// It uses a breadth-first search (BFS) algorithm to calculate the shortest path lengths.
// It returns a map where the keys are nodes and the values are their respective distances from the source node.
// Unreachable nodes will not be included in the map.
func (g *Graph[N, E]) GetDistanceMapBySource(source N) map[N]int {
	if !g.HasNode(source) {
		return nil
	}

	distance := make(map[N]int)
	visited := make(map[N]bool)
	queue := []N{source}
	distance[source] = 0

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if visited[node] {
			continue
		}
		visited[node] = true

		for _, edge := range g.AdjacencyList[node] {
			target := edge.GetTarget()
			if _, exists := distance[target]; !exists {
				distance[target] = distance[node] + 1
				queue = append(queue, target)
			}
		}
	}

	return distance
}
