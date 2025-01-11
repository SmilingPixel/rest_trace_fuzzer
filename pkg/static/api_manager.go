package static

import (

	"github.com/getkin/kin-openapi/openapi3"
)

// APIManager represents an API manager that manages the API definition,
// dependency graph, dataflow graph and other static information of the API.
type APIManager struct {
	// The OpenAPI definition of the API.
	APIDefinition *openapi3.T

	APIMap map[SimpleAPIMethod]*openapi3.Operation

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
	m.APIDefinition = doc
	m.APIMap = make(map[SimpleAPIMethod]*openapi3.Operation)
	for path, pathItem := range doc.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			m.APIMap[SimpleAPIMethod{Method: method, Endpoint: path}] = operation
		}
	}
}
