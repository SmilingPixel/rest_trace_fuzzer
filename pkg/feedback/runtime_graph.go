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

// UpdateFromCallInfos updates the runtime graph from the call information.
func (g *RuntimeGraph) UpdateFromCallInfos(callInfos []*CallInfo) error {
	// TODO: Implement this method. @xunzhou24
	return nil
}
