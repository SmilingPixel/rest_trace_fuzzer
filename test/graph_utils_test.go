package test

import (
	"resttracefuzzer/pkg/utils"
	"testing"
)

// TestNode is a simple string-based node
type TestNode string

// TestEdge is a basic directed edge implementation
type TestEdge struct {
	From TestNode
	To   TestNode
}

func (e TestEdge) GetSource() TestNode { return e.From }
func (e TestEdge) GetTarget() TestNode { return e.To }

func TestGraph_AddEdge_And_HasNode(t *testing.T) {
	g := utils.NewGraph[TestNode, TestEdge]()

	n1 := TestNode("A")
	n2 := TestNode("B")
	edge := TestEdge{From: n1, To: n2}

	g.AddEdge(edge)

	if !g.HasNode(n1) {
		t.Errorf("Expected graph to have node %v", n1)
	}

	if g.HasNode(n2) {
		t.Errorf("Expected graph NOT to have node %v (target-only node)", n2)
	}
}

func TestGraph_CanReach(t *testing.T) {
	g := utils.NewGraph[TestNode, TestEdge]()

	// Create nodes
	a := TestNode("A")
	b := TestNode("B")
	c := TestNode("C")
	d := TestNode("D")

	// Create edges
	g.AddEdge(TestEdge{From: a, To: b})
	g.AddEdge(TestEdge{From: b, To: c})
	g.AddEdge(TestEdge{From: c, To: d})

	tests := []struct {
		from     TestNode
		to       TestNode
		expected bool
	}{
		{a, d, true},  // A → B → C → D
		{a, c, true},  // A → B → C
		{b, a, false}, // no back edge
		{d, a, false}, // disconnected
		{a, a, true},  // trivial self reach
	}

	for _, tt := range tests {
		result := g.CanReach(tt.from, tt.to)
		if result != tt.expected {
			t.Errorf("CanReach(%v → %v) = %v; expected %v", tt.from, tt.to, result, tt.expected)
		}
	}
}
