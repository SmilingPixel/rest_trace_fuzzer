package static

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

// APIManager represents an API manager that manages the API definition,
// dependency graph, dataflow graph and other static information of the API.
type APIManager struct {
	// The OpenAPI definition of the API.
	APIDoc *openapi3.T

	// The map from the simple API method to the OpenAPI operation.
	APIMap map[SimpleAPIMethod]*openapi3.Operation

	// Internal APIs of the services in the system.
	InternalServiceAPIDoc *openapi3.T

	// The map from the service name to the map from the method to the OpenAPI operation.
	InternalServiceAPIMap map[string]map[SimpleAPIMethod]*openapi3.Operation

	// The dependency graph of the API.
	APIDependencyGraph *APIDependencyGraph

	// The Dataflow graph of the internal APIs.
	APIDataflowGraph *APIDataflowGraph
}

// NewAPIManager creates a new APIManager.
func NewAPIManager() *APIManager {
	return &APIManager{}
}

// InitFromDoc initializes the API manager from an OpenAPI document.
// The document is of interfaces of the whole system.
func (m *APIManager) InitFromSystemDoc(doc *openapi3.T) {
	m.APIDoc = doc
	m.APIMap = make(map[SimpleAPIMethod]*openapi3.Operation)
	for path, pathItem := range doc.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			// By default, the type of the API is HTTP.
			m.APIMap[SimpleAPIMethod{Method: method, Endpoint: path, Typ: SimpleAPIMethodTypeHTTP}] = operation
		}
	}
}

// InitFromServiceDocs initializes the API manager from the OpenAPI document of the services.
func (m *APIManager) InitFromServiceDoc(doc *openapi3.T) {
	m.InternalServiceAPIDoc = doc
	m.InternalServiceAPIMap = make(map[string]map[SimpleAPIMethod]*openapi3.Operation)
	// TODO: Implement dependency graph and dataflow graph initialization @xunzhou24
	for _, pathItem := range doc.Paths.Map() {
		for _, operation := range pathItem.Operations() {
			// In OpenAPI generated by protoc-gen-openapi, operationID is in the format of `{Service}_{Method}`.
			// We can use this format to extract the service name and method name.
			operationID := operation.OperationID
			// Split the operationID by `_`.
			operationIDParts := strings.Split(operationID, "_")
			if len(operationIDParts) != 2 {
				log.Warn().Msgf("[APIManager.InitFromServiceDoc] Invalid operationID: %s", operationID)
				continue
			}
			serviceName := operationIDParts[0]
			methodName := operationIDParts[1]
			// TODO: we treat all the methods as gRPC methods for now. @xunzhou24
			simpleMethod := SimpleAPIMethod{Method: methodName, Typ: SimpleAPIMethodTypeGRPC}
			if _, exists := m.InternalServiceAPIMap[serviceName]; !exists {
				m.InternalServiceAPIMap[serviceName] = make(map[SimpleAPIMethod]*openapi3.Operation)
			}
			m.InternalServiceAPIMap[serviceName][simpleMethod] = operation
		}
	}

	// Generate the dataflow graph of the internal APIs.
	m.APIDataflowGraph = NewAPIDataflowGraph()
	m.APIDataflowGraph.ParseFromServiceDocument(m.InternalServiceAPIMap)
}
