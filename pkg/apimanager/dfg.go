package apimanager

import "github.com/getkin/kin-openapi/openapi3"

// APIDataflowNode represents a node in the dataflow graph of the internal APIs.
type APIDataflowNode struct {
	ServiceName string
	SimpleAPIMethod *SimpleAPIMethod
	Operation 	 *openapi3.Operation
}

// APIDataflowEdge represents an edge in the dataflow graph of the internal APIs.
// 
// The edge represents the dataflow between two nodes.
// The data pass from SourceData to TargetData, both of which are parameters of the API.
// For example, `placeOrder` of CheckoutService passes `userInfo` to `emptyCart` of CartService.
type APIDataflowEdge struct {
	Source *APIDataflowNode
	Target *APIDataflowNode
	SourceData *openapi3.Parameter
	TargetData *openapi3.Parameter
}

// APIDataflowGraph represents the dataflow graph of the internal APIs.
type APIDataflowGraph struct {
	Edges []*APIDataflowEdge
}

// NewAPIDataflowGraph creates a new APIDataflowGraph.
func NewAPIDataflowGraph() *APIDataflowGraph {
	edges := make([]*APIDataflowEdge, 0)
	return &APIDataflowGraph{
		Edges: edges,
	}
}


// AddEdge adds an edge to the dataflow graph.
//
// `serviceDocs` is a map from service name to the OpenAPI3 document of the service.
func (g *APIDataflowGraph) ParseFromServiceDocument(serviceDocs map[string]*openapi3.T) {
	// TODO: Implement this function @xunzhou24
}
	
