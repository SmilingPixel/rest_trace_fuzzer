package feedback

import "resttracefuzzer/pkg/static"


// RuntimeEdge represents an edge in the runtime graph.
// It includes static info(source and target) and runtime info(hit count).
type RuntimeEdge struct {
	Source *static.APIDataflowNode
	Target *static.APIDataflowNode
	HitCount int
}

// RuntimeGraph represents the runtime graph. It includes a list of edges.
type RuntimeGraph struct {
	Edges []*RuntimeEdge
}

// NewRuntimeGraph creates a new RuntimeGraph.
func NewRuntimeGraph() *RuntimeGraph {
	edges := make([]*RuntimeEdge, 0)
	return &RuntimeGraph{
		Edges: edges,
	}
}
