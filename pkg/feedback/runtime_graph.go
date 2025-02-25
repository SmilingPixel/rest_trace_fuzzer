package feedback

import (
	"resttracefuzzer/pkg/feedback/trace"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"
)

// RuntimeEdge represents an edge in the runtime graph.
// It includes static info(source and target) and runtime info(hit count).
type RuntimeEdge struct {
	Source   *static.APIDataflowNode `json:"source"`
	Target   *static.APIDataflowNode `json:"target"`
	HitCount int 				   `json:"hit_count"`
}

// RuntimeGraph represents the runtime graph. It includes a list of edges.
type RuntimeGraph struct {
	Edges []*RuntimeEdge `json:"edges"`
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
func (g *RuntimeGraph) UpdateFromCallInfos(callInfos []*trace.CallInfo) error {
	// Group by source service
	// An issue found during development:
	// The source service in callInfo is not the completely same as the source service in runtimeGraph, they may be in different cases.
	// For example, callInfo.SourceService = "cartservice", but runtimeGraph.SourceService = "CartService".
	// We handle it by converting both names into standard cases, and compare them.
	// TODO: why not save standard case in the first place? @xunzhou24
	service2CallInfos := make(map[string][]*trace.CallInfo)
	for _, callInfo := range callInfos {
		sourceService := utils.ConvertToStandardCase(callInfo.SourceService)
		service2CallInfos[sourceService] = append(service2CallInfos[sourceService], callInfo)
	}

	// Iterate over, and update the hit count of the edges.
	for _, edge := range g.Edges {
		sourceService := utils.ConvertToStandardCase(edge.Source.ServiceName)
		for _, callInfo := range service2CallInfos[sourceService] {
			// TODO: A more graceful name matching strategy. @xunzhou24
			// TODO: handle: edge in callInfo is not included in parsed runtimeGraph. @xunzhou24
			// Method or operation name in trace may contain other information, e.g., /oteldemo.CartService/GetCart
			// We extract the last segment of the method name, and compare it with the method name in runtimeGraph.
			if utils.ConvertToStandardCase(callInfo.TargetService) == utils.ConvertToStandardCase(edge.Target.ServiceName) &&
				utils.ExtractLastSegment(callInfo.TargetMethodTraceName, "./") == (edge.Target.SimpleAPIMethod.Method) &&
				utils.ExtractLastSegment(callInfo.SourceMethodTraceName, "./") == (edge.Source.SimpleAPIMethod.Method) {
				edge.HitCount++
			}
		}
	}
	return nil
}

// GetEdgeCoverage returns the edge coverage of the runtime graph.
func (g *RuntimeGraph) GetEdgeCoverage() float64 {
	coveredEdges := 0
	for _, edge := range g.Edges {
		if edge.HitCount > 0 {
			coveredEdges++
		}
	}
	return float64(coveredEdges) / float64(len(g.Edges))
}
