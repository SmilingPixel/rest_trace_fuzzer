package static

import (
	"resttracefuzzer/pkg/utils"

	"github.com/bytedance/sonic"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

// APIDataflowNode represents a node in the dataflow graph of the internal APIs.
type APIDataflowNode struct {
	ServiceName     string
	SimpleAPIMethod SimpleAPIMethod
}

// APIDataflowEdge represents an edge in the dataflow graph of the internal APIs.
//
// The edge represents the dataflow between two nodes.
// The data pass from SourceData to TargetData, both of which are parameters of the API.
// For example, `placeOrder` of CheckoutService passes `userInfo` to `emptyCart` of CartService.
type APIDataflowEdge struct {
	Source         *APIDataflowNode
	Target         *APIDataflowNode
	SourceProperty SimpleAPIProperty
	TargetProperty SimpleAPIProperty
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
func (g *APIDataflowGraph) ParseFromServiceDocument(serviceDocMap map[string]map[SimpleAPIMethod]*openapi3.Operation) {
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

	// log parsed dataflow graph, for debugging
	dfgJson, err := sonic.MarshalString(g)
	if err != nil {
		log.Error().Err(err).Msg("[APIDataflowGraph.ParseFromServiceDocument] Failed to marshal dataflow graph")
	} else {
		log.Debug().Msgf("[APIDataflowGraph.ParseFromServiceDocument] Dataflow graph: %s", dfgJson)
	}
}

// parseServiceOperationPair parses the dataflow between two operations.
func (g *APIDataflowGraph) parseServiceOperationPair(
	sourceService string,
	sourceMethod SimpleAPIMethod,
	sourceOperation *openapi3.Operation,
	targetService string,
	targetMethod SimpleAPIMethod,
	targetOperation *openapi3.Operation,
) {
	// Retrive all properties from parameters, request and response bodies
	sourceInProperties := make([]SimpleAPIProperty, 0)
	targetInProperties := make([]SimpleAPIProperty, 0)

	// Parameter
	sourceParameters := sourceOperation.Parameters
	for _, sourceParamRef := range sourceParameters {
		sourceParam := sourceParamRef.Value
		simpleAPIProperty := SimpleAPIProperty{
			Name: sourceParam.Name,
		}
		sourceInProperties = append(sourceInProperties, simpleAPIProperty)
	}

	targetParameters := targetOperation.Parameters
	for _, targetParamRef := range targetParameters {
		targetParam := targetParamRef.Value
		simpleAPIProperty := SimpleAPIProperty{
			Name: targetParam.Name,
		}
		targetInProperties = append(targetInProperties, simpleAPIProperty)
	}

	// Request body
	if sourceOperation.RequestBody != nil {
		flattenedSourceRequestBody, err := utils.FlattenSchema(sourceOperation.RequestBody.Value.Content.Get("application/json").Schema)
		if err != nil {
			log.Error().Err(err).Msg("[parseServiceOperationPair] Failed to flatten source request body")
		}
		for schemaName := range flattenedSourceRequestBody {
			simpleAPIProperty := SimpleAPIProperty{
				Name: schemaName,
			}
			sourceInProperties = append(sourceInProperties, simpleAPIProperty)
		}
	}

	if targetOperation.RequestBody != nil {
		flattenedTargetRequestBody, err := utils.FlattenSchema(targetOperation.RequestBody.Value.Content.Get("application/json").Schema)
		if err != nil {
			log.Error().Err(err).Msg("[parseServiceOperationPair] Failed to flatten target request body")
		}
		for schemaName := range flattenedTargetRequestBody {
			simpleAPIProperty := SimpleAPIProperty{
				Name: schemaName,
			}
			targetInProperties = append(targetInProperties, simpleAPIProperty)
		}
	}

	// Response body
	// TODO: implement it @xunzhou24

	for _, sourceProp := range sourceInProperties {
		for _, targetProp := range targetInProperties {
			// TODO: better algorithm for matching parameters @xunzhou24
			if utils.MatchVariableNames(sourceProp.Name, targetProp.Name) {
				sourceNode := &APIDataflowNode{
					ServiceName:     sourceService,
					SimpleAPIMethod: sourceMethod,
				}
				targetNode := &APIDataflowNode{
					ServiceName:     targetService,
					SimpleAPIMethod: targetMethod,
				}
				g.AddEdge(sourceNode, targetNode, sourceProp, targetProp)
			}
		}
	}
}

// AddEdge adds an edge to the dataflow graph.
func (g *APIDataflowGraph) AddEdge(source, target *APIDataflowNode, sourceProp, targetProp SimpleAPIProperty) {
	edge := &APIDataflowEdge{
		Source:         source,
		Target:         target,
		SourceProperty: sourceProp,
		TargetProperty: targetProp,
	}
	log.Info().Msgf("[AddEdge] Adding edge: %v -> %v", source, target)
	g.Edges = append(g.Edges, edge)
}
