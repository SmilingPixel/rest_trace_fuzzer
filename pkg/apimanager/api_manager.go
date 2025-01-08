package apimanager

import "github.com/getkin/kin-openapi/openapi3"

// APIManager represents an API manager that manages the API definition and its dependencies.
type APIManager struct {
	// The OpenAPI definition of the API.
	APIDefinition *openapi3.T

	// The dependency graph of the API.
	APIDependencyGraph *APIDependencyGraph

	// The Dataflow graph of the internal APIs.
	APIDataflowGraph *APIDataflowGraph
}

// NewAPIManager creates a new APIManager.
func NewAPIManager() *APIManager {
	return &APIManager{}
}