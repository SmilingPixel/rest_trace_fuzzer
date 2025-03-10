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
	HitCount int 				   `json:"hitCount"`
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
	sourceService2CallInfos := make(map[string][]*trace.CallInfo)
	for _, callInfo := range callInfos {
		sourceService := formatServiceName(callInfo.SourceService)
		sourceService2CallInfos[sourceService] = append(sourceService2CallInfos[sourceService], callInfo)
	}

	// Iterate over, and update the hit count of the edges.
	for _, edge := range g.Edges {
		sourceService := formatServiceName(edge.Source.ServiceName)
		for _, callInfo := range sourceService2CallInfos[sourceService] {
			// TODO: A more graceful name matching strategy. @xunzhou24
			// TODO: handle: edge in callInfo is not included in parsed runtimeGraph. @xunzhou24
			// When conditions below are met, we consider the edge is hit:
			//  1. The source and target service names match (after being converted into standard case).
			//  2. The method in callInfo (i.e., the method called) must match the method in edge's target (i.e., target of data flow).
			if formatServiceName(callInfo.TargetService) == formatServiceName(edge.Target.ServiceName) &&
				callInfo.Method == edge.Target.SimpleAPIMethod.Method {
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

// formatServiceName formats the service name.
// It does the following:
//  1. Convert the name to "standard case".(See [resttracefuzzer/pkg/utils.ConvertToStandardCase])
//  2. remove the suffix "service" if exists.
func formatServiceName(name string) string {
	name = utils.ConvertToStandardCase(name)
	if len(name) > 7 && name[len(name)-7:] == "service" {
		name = name[:len(name)-7]
	}
	return name
}
