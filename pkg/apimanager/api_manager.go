package apimanager

import "github.com/getkin/kin-openapi/openapi3"

// APIManager represents an API manager that manages the API definition and its dependencies.
type APIManager struct {
	APIDefinition *openapi3.T
	APIDependencyGraph *APIDependencyGraph
}

// NewAPIManager creates a new APIManager.
func NewAPIManager() *APIManager {
	return &APIManager{}
}