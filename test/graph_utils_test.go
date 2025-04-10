package test

import (
	"resttracefuzzer/pkg/utils"
	"testing"
)

// TestNode is a concrete implementation of AbstractNode for testing.
type TestNode struct {
	id string
}

func (n *TestNode) ID() string {
	return n.id
}

// TestGraph is a concrete implementation of AbstractGraph for testing.
type TestGraph struct {
	nodes map[string]*TestNode
	edges map[string][]*TestNode
}

// NewTestGraph initializes a new test graph.
func NewTestGraph() *TestGraph {
	return &TestGraph{
		nodes: make(map[string]*TestNode),
		edges: make(map[string][]*TestNode),
	}
}

// AddNode adds a new node with the given ID to the graph.
func (g *TestGraph) AddNode(id string) *TestNode {
	node := &TestNode{id: id}
	g.nodes[id] = node
	return node
}

// AddEdge adds a directed edge from one node to another.
func (g *TestGraph) AddEdge(from, to *TestNode) {
	g.edges[from.id] = append(g.edges[from.id], to)
}

// HasNode checks if the node exists in the graph.
func (g *TestGraph) HasNode(node utils.AbstractNode) bool {
	_, exists := g.nodes[node.ID()]
	return exists
}

// GetNeighborsOf returns neighbors of the specified node.
func (g *TestGraph) GetNeighborsOf(node utils.AbstractNode) []utils.AbstractNode {
	rawNeighbors := g.edges[node.ID()]
	neighbors := make([]utils.AbstractNode, len(rawNeighbors))
	for i, n := range rawNeighbors {
		neighbors[i] = n
	}
	return neighbors
}

func TestCanReach(t *testing.T) {
	graph := NewTestGraph()

	// Create nodes
	nodeA := graph.AddNode("A")
	nodeB := graph.AddNode("B")
	nodeC := graph.AddNode("C")
	nodeD := graph.AddNode("D")

	// Create edges
	graph.AddEdge(nodeA, nodeB)
	graph.AddEdge(nodeB, nodeC)
	graph.AddEdge(nodeC, nodeD)

	// Define test cases
	tests := []struct {
		name     string
		from     *TestNode
		to       *TestNode
		expected bool
	}{
		{"A to D", nodeA, nodeD, true},
		{"A to C", nodeA, nodeC, true},
		{"B to A", nodeB, nodeA, false},
		{"D to A", nodeD, nodeA, false},
		{"Self reach", nodeB, nodeB, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.CanReach(graph, tc.from, tc.to)
			if result != tc.expected {
				t.Errorf("CanReach(%s -> %s) = %v; expected %v", tc.from.ID(), tc.to.ID(), result, tc.expected)
			}
		})
	}
}
