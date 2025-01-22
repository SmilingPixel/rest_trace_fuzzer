package feedback

import "resttracefuzzer/pkg/static"

// RuntimeEdge represents an edge in the runtime graph.
// It includes static info(source and target) and runtime info(hit count).
type RuntimeEdge struct {
	Source   *static.APIDataflowNode
	Target   *static.APIDataflowNode
	HitCount int
}

// RuntimeGraph represents the runtime graph. It includes a list of edges.
type RuntimeGraph struct {
	Edges []*RuntimeEdge
}

// NewRuntimeGraph creates a new RuntimeGraph.
// It initializes the edges from the static API dataflow graph.
func NewRuntimeGraph(APIDataflowGraph *static.APIDataflowGraph) *RuntimeGraph {
	edges := make([]*RuntimeEdge, 0)
	for _, edge := range APIDataflowGraph.Edges {
		edges = append(edges, &RuntimeEdge{
			Source:   edge.Source,
			Target:   edge.Target,
			HitCount: 0,
		})
	}
	return &RuntimeGraph{
		Edges: edges,
	}
}

// UpdateFromCallInfos updates the runtime graph from the call information.
func (g *RuntimeGraph) UpdateFromCallInfos(callInfos []*CallInfo) error {
	// TODO: Implement this method. @xunzhou24
	// Group by source service
	service2CallInfos := make(map[string][]*CallInfo)
	for _, callInfo := range callInfos {
		sourceService := callInfo.SourceService
		service2CallInfos[sourceService] = append(service2CallInfos[sourceService], callInfo)
	}

	// Iterate over, and update the hit count of the edges.
	for _, edge := range g.Edges {
		sourceService := edge.Source.ServiceName
		for _, callInfo := range service2CallInfos[sourceService] {
			// TODO: A more graceful name matching strategy. @xunzhou24
			// TODO: handle: edge in callInfo is not included in parsed runtimeGraph. @xunzhou24
			if callInfo.TargetService == edge.Target.ServiceName &&
				callInfo.TargetMethodTraceName == edge.Target.SimpleAPIMethod.Method &&
				callInfo.SourceMethodTraceName == edge.Source.SimpleAPIMethod.Method {
				edge.HitCount++
			}
		}
	}
	return nil
}
