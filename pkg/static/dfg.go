package static

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

// APIDataflowNode represents a node in the dataflow graph of the internal APIs.
type APIDataflowNode struct {
	ServiceName     string
	SimpleAPIMethod *SimpleAPIMethod
	Operation       *openapi3.Operation
}

// APIDataflowEdge represents an edge in the dataflow graph of the internal APIs.
//
// The edge represents the dataflow between two nodes.
// The data pass from SourceData to TargetData, both of which are parameters of the API.
// For example, `placeOrder` of CheckoutService passes `userInfo` to `emptyCart` of CartService.
type APIDataflowEdge struct {
	Source     *APIDataflowNode
	Target     *APIDataflowNode
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
// `serviceDocMap` is a map from the service name to the map from the method name to the OpenAPI operation.
func (g *APIDataflowGraph) ParseFromServiceDocument(serviceDocMap map[string]map[string]*openapi3.Operation) {
	for sourceService, sourceMethodMap := range serviceDocMap {
		for targetService, targetMethodMap := range serviceDocMap {
			if sourceService == targetService {
				continue
			}
			for sourceMethod, sourceOperation := range sourceMethodMap {
				for targetMethod, targetOperation := range targetMethodMap {
					g.parseServiceOperationPair(
						sourceService, sourceMethod, sourceOperation,
						targetService, targetMethod, targetOperation,
					)
				}
			}
		}
	}
}


// parseServiceOperationPair parses the dataflow between two operations.
func (g *APIDataflowGraph) parseServiceOperationPair(
	sourceService string,
	sourceMethod string,
	sourceOperation *openapi3.Operation,
	targetService string,
	targetMethod string,
	targetOperation *openapi3.Operation,
) {
	sourceInParameters := make([]*openapi3.Parameter, 0)
	// sourceOutParameters := make([]*openapi3.Parameter, 0)
	targetInParameters := make([]*openapi3.Parameter, 0)
	// targetOutParameters := make([]*openapi3.Parameter, 0)
	for _, sourceParamRef := range sourceOperation.Parameters {
		if sourceParam := sourceParamRef.Value; sourceParam != nil {
			sourceInParameters = append(sourceInParameters, sourceParam)
		}
	}
	for _, targetParamRef := range targetOperation.Parameters {
		if targetParam := targetParamRef.Value; targetParam != nil {
			targetInParameters = append(targetInParameters, targetParam)
		}
	}

	for _, sourceInParam := range sourceInParameters {
		for _, targetInParam := range targetInParameters {
			// TODO: better algorithm for matching parameters
			if sourceInParam.Name == targetInParam.Name {
				sourceNode := &APIDataflowNode{
					ServiceName: sourceService,
					SimpleAPIMethod: &SimpleAPIMethod{
						Method:   sourceMethod,
						Type:    SimpleAPIMethodTypeGRPC,
					},
					Operation: sourceOperation,
				}
				targetNode := &APIDataflowNode{
					ServiceName: targetService,
					SimpleAPIMethod: &SimpleAPIMethod{
						Method:   targetMethod,
						Type:    SimpleAPIMethodTypeGRPC,
					},
					Operation: targetOperation,
				}
				g.AddEdge(sourceNode, targetNode, sourceInParam, targetInParam)
			}
		}
	}
}

// AddEdge adds an edge to the dataflow graph.
func (g *APIDataflowGraph) AddEdge(source, target *APIDataflowNode, sourceData, targetData *openapi3.Parameter) {
	edge := &APIDataflowEdge{
		Source:     source,
		Target:     target,
		SourceData: sourceData,
		TargetData: targetData,
	}
	log.Info().Msgf("[AddEdge] Adding edge: %v -> %v", source, target)
	g.Edges = append(g.Edges, edge)
}
