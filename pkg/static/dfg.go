package static

import (
	"resttracefuzzer/pkg/utils"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

// APIDataflowEdge represents an edge in the dataflow graph of the internal APIs.
//
// The edge represents the dataflow between two nodes.
// The data pass from SourceData to TargetData, both of which are parameters of the API.
// For example, `placeOrder` of CheckoutService passes `userInfo` to `emptyCart` of CartService.
type APIDataflowEdge struct {
	Source         InternalServiceEndpoint `json:"source"`
	Target         InternalServiceEndpoint `json:"target"`
	SourceProperty SimpleAPIProperty        `json:"sourceProperty"`
	TargetProperty SimpleAPIProperty        `json:"targetProperty"`
}

func (e *APIDataflowEdge) GetSource() InternalServiceEndpoint {
	return e.Source
}

func (e *APIDataflowEdge) GetTarget() InternalServiceEndpoint {
	return e.Target
}

// APIDataflowGraph represents the dataflow graph of the internal APIs.
// It implements [resttracefuzzer/pkg/utils/AbstractGraph] interface, to support graph related algorithms.
type APIDataflowGraph struct {
	*utils.Graph[InternalServiceEndpoint, *APIDataflowEdge]
}

// NewAPIDataflowGraph creates a new APIDataflowGraph.
func NewAPIDataflowGraph() *APIDataflowGraph {
	graph := utils.NewGraph[InternalServiceEndpoint, *APIDataflowEdge]()
	return &APIDataflowGraph{
		Graph: graph,
	}
}

// ParseFromServiceDocument parses the dataflow graph from the service document.
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
	dfgJson, err := sonic.MarshalString(g.Edges[0].Source.SimpleAPIMethod.Typ)
	if err != nil {
		log.Err(err).Msg("[APIDataflowGraph.ParseFromServiceDocument] Failed to marshal dataflow graph")
	} else {
		log.Debug().Msgf("[APIDataflowGraph.ParseFromServiceDocument] Dataflow graph: %s", dfgJson)
	}
}

// parseServiceOperationPair parses the dataflow between two operations, and update the dataflow graph.
func (g *APIDataflowGraph) parseServiceOperationPair(
	sourceService string,
	sourceMethod SimpleAPIMethod,
	sourceOperation *openapi3.Operation,
	targetService string,
	targetMethod SimpleAPIMethod,
	targetOperation *openapi3.Operation,
) {
	// Retrieve all properties from parameters, request and response bodies
	// sourceRequestProperties and targetRequestProperties are the properties that are passed from source request, respectively, including parameters and request body.
	// sourceResponseProperties and targetResponseProperties are the properties that are passed from source response, respectively.
	sourceRequestProperties := make([]SimpleAPIProperty, 0)
	targetRequestProperties := make([]SimpleAPIProperty, 0)
	sourceResponseProperties := make([]SimpleAPIProperty, 0)
	targetResponseProperties := make([]SimpleAPIProperty, 0)

	// Parameter
	sourceParameters := sourceOperation.Parameters
	for _, sourceParamRef := range sourceParameters {
		sourceParam := sourceParamRef.Value
		simpleAPIProperty := SimpleAPIProperty{
			Name: sourceParam.Name,
			Typ:  OpenAPITypes2SimpleAPIPropertyType(sourceParam.Schema.Value.Type),
		}
		sourceRequestProperties = append(sourceRequestProperties, simpleAPIProperty)
	}

	targetParameters := targetOperation.Parameters
	for _, targetParamRef := range targetParameters {
		targetParam := targetParamRef.Value
		simpleAPIProperty := SimpleAPIProperty{
			Name: targetParam.Name,
			Typ:  OpenAPITypes2SimpleAPIPropertyType(targetParam.Schema.Value.Type),
		}
		targetRequestProperties = append(targetRequestProperties, simpleAPIProperty)
	}

	// Request body
	if sourceOperation.RequestBody != nil {
		sourceRequestProperties = append(
			sourceRequestProperties,
			extractPropertiesFromSchema(sourceOperation.RequestBody.Value.Content.Get("application/json").Schema)...,
		)
	}

	if targetOperation.RequestBody != nil {
		targetRequestProperties = append(
			targetRequestProperties,
			extractPropertiesFromSchema(targetOperation.RequestBody.Value.Content.Get("application/json").Schema)...,
		)
	}

	// Response body
	// We only handle response with status code 200 (OK), 201 (Created), 202 (Accepted)
	successStatusCode := []int{consts.StatusOK, consts.StatusCreated, consts.StatusAccepted}
	if sourceOperation.Responses != nil {
		var sourceResponse *openapi3.ResponseRef
		var exist bool
		for _, statusCode := range successStatusCode {
			sourceResponse, exist = sourceOperation.Responses.Map()[strconv.FormatInt(int64(statusCode), 10)]
			if exist {
				break
			}
		}
		if !exist {
			log.Warn().Msgf("[APIDataflowGraph.parseServiceOperationPair] No response with status codes 200, 201, or 202 found, operation ID: %s", sourceOperation.OperationID)
		} else {
			contentMap := sourceResponse.Value.Content
			if len(contentMap) == 0 {
				log.Debug().Msgf("[APIDataflowGraph.parseServiceOperationPair] No response content found, operation ID: %s", sourceOperation.OperationID)
			} else {
				sourceResponseProperties = extractPropertiesFromSchema(sourceResponse.Value.Content.Get("application/json").Schema)
			}
		}
	}

	if targetOperation.Responses != nil {
		var targetResponse *openapi3.ResponseRef
		var exist bool
		for _, statusCode := range successStatusCode {
			targetResponse, exist = targetOperation.Responses.Map()[strconv.FormatInt(int64(statusCode), 10)]
			if exist {
				break
			}
		}
		if !exist {
			log.Warn().Msgf("[APIDataflowGraph.parseServiceOperationPair] No response with success status codes (200, 201, 202) found, operation ID: %s", targetOperation.OperationID)
		} else {
			contentMap := targetResponse.Value.Content
			if len(contentMap) == 0 {
				log.Warn().Msgf("[APIDataflowGraph.parseServiceOperationPair] No response content found, operation ID: %s", targetOperation.OperationID)
			} else {
				targetResponseProperties = extractPropertiesFromSchema(targetResponse.Value.Content.Get("application/json").Schema)
			}
		}
	}

	// Match the properties and update the dataflow graph
	g.tryMatchPropertiesAndUpdateGraph(
		sourceService, sourceMethod, sourceRequestProperties,
		targetService, targetMethod, targetRequestProperties,
	)
	g.tryMatchPropertiesAndUpdateGraph(
		sourceService, sourceMethod, sourceResponseProperties,
		targetService, targetMethod, targetResponseProperties,
	)
}

// extractPropertiesFromSchema extracts the properties from the schema.
// It returns all properties in the schema in a flattened way.
func extractPropertiesFromSchema(schema *openapi3.SchemaRef) []SimpleAPIProperty {
	// Flatten the schema, mapping from the schema name to the schema
	// For example:
	//  {
	//    "name": {
	//      "type": "string"
	//      ...
	//    },
	//    "age": {
	//      "type": "integer"
	//      ...
	//    }
	//  }
	flattenedSchemaMap, err := utils.FlattenSchema(schema)
	if err != nil {
		log.Err(err).Msg("[extractPropertiesFromSchema] Failed to flatten schema")
		return nil
	}
	var properties []SimpleAPIProperty
	for schemaName, schema := range flattenedSchemaMap {
		simpleAPIProperty := SimpleAPIProperty{
			Name: schemaName,
			Typ:  OpenAPITypes2SimpleAPIPropertyType(schema.Value.Type),
		}
		properties = append(properties, simpleAPIProperty)
	}
	return properties
}

// tryMatchPropertiesAndUpdateGraph tries to match the properties and update the dataflow graph.
// If a parameter in source request matches a parameter in target request, we can assume there exists a dataflow between the two operations.
// Multiple edges are not allowed between the same source and target nodes.
// Similarly, if a property in source response matches a property in target response, we can assume there exists a dataflow between the two operations.
// We use LevenshteinSimilarityCalculator to calculate the similarity between two strings, and the threshold is 0.75.
// TODO: make SimilarityCalculator and threshold configurable @xunzhou24
func (g *APIDataflowGraph) tryMatchPropertiesAndUpdateGraph(
	sourceService string,
	sourceMethod SimpleAPIMethod,
	sourceProperties []SimpleAPIProperty,
	targetService string,
	targetMethod SimpleAPIMethod,
	targetProperties []SimpleAPIProperty,
) {
	similarityCalculator := utils.NewLevenshteinSimilarityCalculator()
	threshold := 0.75
	for _, sourceProp := range sourceProperties {
		for _, targetProp := range targetProperties {
			// TODO: better algorithm for matching parameters @xunzhou24
			if utils.MatchVariableNames(sourceProp.Name, targetProp.Name, similarityCalculator, threshold) {
				sourceNode := InternalServiceEndpoint{
					ServiceName:     sourceService,
					SimpleAPIMethod: sourceMethod,
				}
				targetNode := InternalServiceEndpoint{
					ServiceName:     targetService,
					SimpleAPIMethod: targetMethod,
				}
				edge := &APIDataflowEdge{
					Source:         sourceNode,
					Target:         targetNode,
					SourceProperty: sourceProp,
					TargetProperty: targetProp,
				}
				g.AddEdge(edge)
				log.Trace().Msgf("[APIDataflowGraph.tryMatchPropertiesAndUpdateGraph] Adding edge: %v -> %v, source property:, %v, target property: %v", sourceNode, targetNode, sourceProp, targetProp)
				// Only one edge is allowed between the same source and target nodes
				return
			}
		}
	}
}
