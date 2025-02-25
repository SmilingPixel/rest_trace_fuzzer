package casemanager

import (
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

// An OperationCase includes 3 parts:
//  1. Definition of the API method, including the method type, path, schema, etc.
//  2. The request instance, including the request body, headers, etc. (This should be filled during the test case generation)
//  3. The expected response, including the response body, headers, etc. (This should be filled after the test case execution)
//
// Part 2 and 3 will be re-filled each time the test case is executed.
type OperationCase struct {
	// APIMethod is the definition of the API method being tested.
	APIMethod static.SimpleAPIMethod `json:"api_method"`

	// Operation is the OpenAPI operation definition.
	Operation *openapi3.Operation `json:"operation"`

	// RequestHeaders contains the headers to be sent with the request.
	RequestHeaders map[string]string `json:"request_headers"`

	// RequestPathParams contains the path parameters to be sent with the request.
	RequestPathParams map[string]string `json:"request_path_params"`

	// RequestQueryParams contains the query parameters to be sent with the request.
	RequestQueryParams map[string]string `json:"request_query_params"`

	// RequestBody contains the body to be sent with the request.
	// It is a json object.
	RequestBody interface{} `json:"request_body"`

	// ResponseHeaders contains the expected headers in the response.
	ResponseHeaders map[string]string `json:"response_headers"`

	// ResponseStatusCode is the expected status code of the response.
	ResponseStatusCode int `json:"response_status_code"`

	// ResponseBody contains the expected body of the response.
	// It is a json object.
	ResponseBody interface{} `json:"response_body"`
}

// A TestScenario is a sequence of [resttracefuzzer/pkg/casemanager/OperationCase].
type TestScenario struct {
	// OperationCases is a sequence of [resttracefuzzer/pkg/casemanager/OperationCase].
	OperationCases []*OperationCase `json:"operation_cases"`

	// ExecutedTimes is the number of times this test scenario has been executed
	ExecutedTimes int `json:"executed_times"`
}

// NewTestScenario creates a new TestScenario.
func NewTestScenario(operationCases []*OperationCase) *TestScenario {
	return &TestScenario{
		OperationCases: operationCases,
		ExecutedTimes:  0,
	}
}

// IsExecutedSuccessfully checks whether the test scenario is executed successfully.
// It only checks the last operation for now.
func (ts *TestScenario) IsExecutedSuccessfully() bool {
	if len(ts.OperationCases) == 0 {
		log.Warn().Msg("[TestScenario.IsExecutedSuccessfully] Test scenario is empty")
		return false
	}
	lastOperationCase := ts.OperationCases[len(ts.OperationCases)-1]
	return lastOperationCase.ResponseStatusCode == 200
}

// Copy creates a deep copy of the test scenario.
func (ts *TestScenario) Copy() *TestScenario {
	operationCases := make([]*OperationCase, len(ts.OperationCases))
	for i, operationCase := range ts.OperationCases {
		operationCases[i] = operationCase.Copy()
	}
	return &TestScenario{
		OperationCases: operationCases,
		ExecutedTimes:  ts.ExecutedTimes,
	}
}

// Reset resets the test scenario.
// It sets the executed times to 0.
func (ts *TestScenario) Reset() {
	ts.ExecutedTimes = 0
}

// IsExecutedSuccessfully checks whether the operation case is executed successfully.
// It only checks the response status code for now.
func (oc *OperationCase) IsExecutedSuccessfully() bool {
	return utils.IsStatusCodeSuccess(oc.ResponseStatusCode)
}

// Copy creates a deep copy of the operation case.
// TODO: deep copy the request and response body. @xunzhou24
func (oc *OperationCase) Copy() *OperationCase {
	requestHeaders := make(map[string]string)
	for k, v := range oc.RequestHeaders {
		requestHeaders[k] = v
	}
	requestParams := make(map[string]string)
	for k, v := range oc.RequestQueryParams {
		requestParams[k] = v
	}
	requestBody := oc.RequestBody
	responseHeaders := make(map[string]string)
	for k, v := range oc.ResponseHeaders {
		responseHeaders[k] = v
	}
	responseBody := oc.ResponseBody
	return &OperationCase{
		APIMethod:          oc.APIMethod,
		Operation:          oc.Operation,
		RequestHeaders:     requestHeaders,
		RequestQueryParams: requestParams,
		RequestBody:        requestBody,
		ResponseHeaders:    responseHeaders,
		ResponseStatusCode: oc.ResponseStatusCode,
		ResponseBody:       responseBody,
	}
}
