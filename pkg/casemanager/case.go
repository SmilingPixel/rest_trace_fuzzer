package casemanager

import (
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"
	"resttracefuzzer/pkg/utils/http"

	"maps"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

const (
	// When increasing or decreasing the energy of a test scenario by a random value (normal distribution),
	// the mean and standard deviation of the normal distribution.
	EnergyIncrMean = 5
	EnergyIncrStdDev = 2
	EnergyDecrMean = 3
	EnergyDecrStdDev = 1

	// Maximal and minimal values for the energy of a test scenario.
	MaxEnergy = 20
	MinEnergy = 0
)

// An OperationCase includes 3 parts:
//  1. Definition of the API method, including the method type, path, schema, etc.
//  2. The request instance, including the request body, headers, etc. (This should be filled during the test case generation)
//  3. The expected response, including the response body, headers, etc. (This should be filled after the test case execution)
//
// Part 2 and 3 will be re-filled each time the test case is executed.
type OperationCase struct {
	// APIMethod is the definition of the API method being tested.
	APIMethod static.SimpleAPIMethod `json:"APIMethod"`

	// Operation is the OpenAPI operation definition.
	Operation *openapi3.Operation `json:"operation"`

	// RequestHeaders contains the headers to be sent with the request.
	RequestHeaders map[string]string `json:"requestHeaders"`

	// RequestPathParams contains the path parameters to be sent with the request.
	RequestPathParams map[string]string `json:"requestPathParams"`

	// RequestQueryParams contains the query parameters to be sent with the request.
	RequestQueryParams map[string]string `json:"requestQueryParams"`

	// RequestBody contains the body to be sent with the request.
	// It is a json object as a byte array.
	RequestBody []byte `json:"requestBody"`

	// ResponseHeaders contains the expected headers in the response.
	ResponseHeaders map[string]string `json:"responseHeaders"`

	// ResponseStatusCode is the expected status code of the response.
	ResponseStatusCode int `json:"responseStatusCode"`

	// ResponseBody contains the expected body of the response.
	// It is a json object as a byte array.
	ResponseBody []byte `json:"responseBody"`
}

// A TestScenario is a sequence of [resttracefuzzer/pkg/casemanager/OperationCase].
type TestScenario struct {
	// OperationCases is a sequence of [resttracefuzzer/pkg/casemanager/OperationCase].
	OperationCases []*OperationCase `json:"operationCases"`

	// ExecutedCount is the number of times the test scenario is executed.
	ExecutedCount int `json:"executedCount"`

	// Energy is the energy of the test scenario.
	// It is used to prioritize the test scenarios.
	// The higher the energy, the higher the priority.
	Energy int `json:"energy"`
}

// NewTestScenario creates a new TestScenario.
func NewTestScenario(operationCases []*OperationCase) *TestScenario {
	return &TestScenario{
		OperationCases: operationCases,
		ExecutedCount:  0,
		Energy:         0,
	}
}

// IsExecutedSuccessfully checks whether the test scenario is executed successfully.
// It only checks whether the last operation is successful for now.
func (ts *TestScenario) IsExecutedSuccessfully() bool {
	if len(ts.OperationCases) == 0 {
		log.Warn().Msg("[TestScenario.IsExecutedSuccessfully] Test scenario is empty")
		return false
	}
	lastOperationCase := ts.OperationCases[len(ts.OperationCases)-1]
	return lastOperationCase.ResponseStatusCode == 200
}

// IncreaseEnergyByRandom increases the energy of the test scenario by a random value (normal distribution).
func (ts *TestScenario) IncreaseEnergyByRandom() {
	added := max(0, int(utils.NormInt64(EnergyIncrMean, EnergyIncrStdDev)))
	ts.Energy = min(ts.Energy + added, MaxEnergy)
}

// DecreaseEnergyByRandom decreases the energy of the test scenario by a random value (normal distribution).
func (ts *TestScenario) DecreaseEnergyByRandom() {
	subtracted := max(0, int(utils.NormInt64(EnergyDecrMean, EnergyDecrStdDev)))
	ts.Energy = max(ts.Energy - subtracted, MinEnergy)
}

// Copy creates a deep copy of the test scenario.
func (ts *TestScenario) Copy() *TestScenario {
	operationCases := make([]*OperationCase, len(ts.OperationCases))
	for i, operationCase := range ts.OperationCases {
		operationCases[i] = operationCase.Copy()
	}
	return &TestScenario{
		OperationCases: operationCases,
		ExecutedCount:  ts.ExecutedCount,
		Energy:         ts.Energy,
	}
}

// Reset resets the test scenario.
// It resets the executed count and energy to 0.
func (ts *TestScenario) Reset() {
	ts.ExecutedCount = 0
	ts.Energy = 0
}

// IsExecutedSuccessfully checks whether the operation case is executed successfully.
// It only checks the response status code for now.
func (oc *OperationCase) IsExecutedSuccessfully() bool {
	return http.IsStatusCodeSuccess(oc.ResponseStatusCode)
}

// Copy creates a deep copy of the operation case.
// TODO: deep copy the request and response body. @xunzhou24
func (oc *OperationCase) Copy() *OperationCase {
	requestHeaders := make(map[string]string)
	maps.Copy(requestHeaders, oc.RequestHeaders)
	requestParams := make(map[string]string)
	maps.Copy(requestParams, oc.RequestQueryParams)
	requestBody := make([]byte, len(oc.RequestBody))
	copy(requestBody, oc.RequestBody)
	responseHeaders := make(map[string]string)
	maps.Copy(responseHeaders, oc.ResponseHeaders)
	responseBody := make([]byte, len(oc.ResponseBody))
	copy(responseBody, oc.ResponseBody)
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

// AppendOperationCase appends an operation case to the test scenario.
func (ts *TestScenario) AppendOperationCase(operationCase *OperationCase) {
	ts.OperationCases = append(ts.OperationCases, operationCase)
}
