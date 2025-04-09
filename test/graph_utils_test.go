package test

import (
	"resttracefuzzer/pkg/utils"
	"testing"
)

// TestNode is a concrete implementation of AbstractNode for testing.
type TestNode struct {
	id string
}

func (n *TestNode) String() string {
	return n.id
}

func (n *TestNode) EqualsTo(other utils.AbstractNode) bool {
	otherNode, ok := other.(*TestNode)
	return ok && n.id == otherNode.id
}

func (n *TestNode) HashCode() uint64 {
	var hash uint64
	for _, char := range n.id {
		hash = hash*31 + uint64(char)
	}
	return hash
}

// TestGraph is a concrete implementation of AbstractGraph for testing.
type TestGraph struct {
	nodes map[string]*TestNode
	edges map[string][]*TestNode
}

func NewTestGraph() *TestGraph {
	return &TestGraph{
		nodes: make(map[string]*TestNode),
		edges: make(map[string][]*TestNode),
	}
}

func (g *TestGraph) AddNode(id string) *TestNode {
	node := &TestNode{id: id}
	g.nodes[id] = node
	return node
}

func (g *TestGraph) AddEdge(from, to *TestNode) {
	g.edges[from.id] = append(g.edges[from.id], to)
}

func (g *TestGraph) HasNode(node utils.AbstractNode) bool {
	_, exists := g.nodes[node.String()]
	return exists
}

func (g *TestGraph) GetNeighborsOf(node utils.AbstractNode) []utils.AbstractNode {
	neighbors := g.edges[node.String()]
	abstractNeighbors := make([]utils.AbstractNode, len(neighbors))
	for i, neighbor := range neighbors {
		abstractNeighbors[i] = neighbor
	}
	return abstractNeighbors
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

	// Test cases
	tests := []struct {
		from     *TestNode
		to       *TestNode
		expected bool
	}{
		{nodeA, nodeD, true},  // Path exists: A -> B -> C -> D
		{nodeA, nodeC, true},  // Path exists: A -> B -> C
		{nodeB, nodeA, false}, // No path exists
		{nodeD, nodeA, false}, // No path exists
	}

	for _, test := range tests {
		result := utils.CanReach(graph, test.from, test.to)
		if result != test.expected {
			t.Errorf("CanReach(%s, %s) = %v; want %v", test.from.String(), test.to.String(), result, test.expected)
		}
	}
}
