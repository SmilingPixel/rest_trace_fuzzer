package runtime

import (
	"resttracefuzzer/pkg/feedback/trace"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"
)

// CallInfoEdge represents an edge in the runtime graph of call info.
// It includes static info (source and target) and runtime call info (hit count).
type CallInfoEdge struct {
	Source   static.InternalServiceEndpoint `json:"source"`
	Target   static.InternalServiceEndpoint `json:"target"`
	HitCount int                             `json:"hitCount"`
}

func (c *CallInfoEdge) GetSource() static.InternalServiceEndpoint {
	return c.Source
}

func (c *CallInfoEdge) GetTarget() static.InternalServiceEndpoint {
	return c.Target
}

// CallInfoGraph represents the runtime graph of call info. It includes a list of edges.
// An issue found during development:
// The source service in callInfo is not the completely same as the source service in callInfoGraph, they may be in different cases.
// For example, callInfo.SourceService = "cartservice", but callInfoGraph.SourceService = "CartService".
// We handle it by converting both names into standard cases (when creating and updating).
type CallInfoGraph struct {
	*utils.Graph[static.InternalServiceEndpoint, *CallInfoEdge]
}

// NewCallInfoGraph creates a new CallInfoGraph.
// It initializes the edges from the static API dataflow graph.
func NewCallInfoGraph(APIDataflowGraph *static.APIDataflowGraph) *CallInfoGraph {
	graph := utils.NewGraph[static.InternalServiceEndpoint, *CallInfoEdge]()
	for _, edge := range APIDataflowGraph.Edges {
		// format service name
		source := edge.Source
		source.ServiceName = utils.FormatServiceName(edge.Source.ServiceName)
		target := edge.Target
		target.ServiceName = utils.FormatServiceName(edge.Target.ServiceName)
		callInfoEdge := &CallInfoEdge{
			Source:   source,
			Target:   target,
			HitCount: 0,
		}
		graph.AddEdge(callInfoEdge)
	}
	return &CallInfoGraph{
		Graph: graph,
	}
}

// UpdateFromCallInfos updates the runtime call info graph from the call information.
func (g *CallInfoGraph) UpdateFromCallInfos(callInfos []*trace.CallInfo) error {
	if len(callInfos) == 0 {
		return nil
	}

	// Group by source service
	sourceService2CallInfos := make(map[string][]*trace.CallInfo)
	for _, callInfo := range callInfos {
		sourceService2CallInfos[callInfo.SourceService] = append(sourceService2CallInfos[callInfo.SourceService], callInfo)
	}

	// Iterate over, and update the hit count of the edges.
	for _, edge := range g.Edges {
		for _, callInfo := range sourceService2CallInfos[edge.Source.ServiceName] {
			// TODO: A more graceful name matching strategy. @xunzhou24
			// TODO: handle: edge in callInfo is not included in parsed callInfoGraph. @xunzhou24
			// When conditions below are met, we consider the edge is hit:
			//  1. The source and target service names match (after being converted into standard case).
			//  2. The method in callInfo (i.e., the method called) must match the method in edge's source or target (i.e., target of data flow).
			if callInfo.TargetService == edge.Target.ServiceName &&
				(callInfo.Method == edge.Target.SimpleAPIMethod.Method || callInfo.Method == edge.Source.SimpleAPIMethod.Method) {
				edge.HitCount++
			}
		}
	}
	return nil
}

// GetEdgeCoverage returns the edge coverage of the runtime call info graph.
func (g *CallInfoGraph) GetEdgeCoverage() float64 {
	coveredEdges := 0
	for _, edge := range g.Edges {
		if edge.HitCount > 0 {
			coveredEdges++
		}
	}
	return float64(coveredEdges) / float64(len(g.Edges))
}
