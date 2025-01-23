package casemanager

import (
	"resttracefuzzer/pkg/static"

	"github.com/getkin/kin-openapi/openapi3"
)

// An OperationCase includes 3 parts:
//  1. Definition of the API method, including the method type, path, schema, etc.
//  2. The request instance, including the request body, headers, etc. (This should be filled during the test case generation)
//  3. The expected response, including the response body, headers, etc. (This should be filled after the test case execution)
//
// OperationCase represents a test case for an API operation.
type OperationCase struct {
	// APIMethod is the definition of the API method being tested.
	APIMethod static.SimpleAPIMethod

	// Operation is the OpenAPI operation definition.
	Operation *openapi3.Operation

	// RequestHeaders contains the headers to be sent with the request.
	RequestHeaders map[string]string

	// RequestParams contains the parameters to be sent with the request.
	RequestParams map[string]string

	// RequestBody contains the body to be sent with the request.
	// It is a json object.
	RequestBody map[string]interface{}

	// ResponseHeaders contains the expected headers in the response.
	ResponseHeaders map[string]string

	// ResponseStatusCode is the expected status code of the response.
	ResponseStatusCode int

	// ResponseBody contains the expected body of the response.
	// It is a json object.
	ResponseBody map[string]interface{}
}

// A TestScenario is a sequence of [resttracefuzzer/pkg/casemanager/OperationCase].
type TestScenario struct {
	// OperationCases is a sequence of [resttracefuzzer/pkg/casemanager/OperationCase].
	OperationCases []*OperationCase
}
