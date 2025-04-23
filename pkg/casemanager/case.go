package casemanager

import (
	"resttracefuzzer/pkg/resource"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"
	"resttracefuzzer/pkg/utils/http"
	"strings"

	"maps"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	// When increasing or decreasing the energy of a test scenario or operation case by a random value (normal distribution),
	// the mean and standard deviation of the normal distribution.
	ScenarioEnergyIncrMean   = 5
	ScenarioEnergyIncrStdDev = 2
	ScenarioEnergyDecrMean   = 3
	ScenarioEnergyDecrStdDev = 1
	OperationCaseEnergyIncrMean   = 5
	OperationCaseEnergyIncrStdDev = 2
	OperationCaseEnergyDecrMean   = 3
	OperationCaseEnergyDecrStdDev = 1

	// Maximal and minimal values for the energy of a test scenario or operation case.
	MaxScenarioEnergy = 20
	MinScenarioEnergy = 0
	MaxOperationCaseEnergy = 20
	MinOperationCaseEnergy = 0
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

	// RequestPathParamResources is the resource representation of the path parameters.
	// It is used to generate or mutate the request path parameters.
	// The field would not be json encoded.
	RequestPathParamResources map[string]resource.Resource `json:"-"`

	// RequestQueryParamResources is the resource representation of the query parameters.
	// It is used to generate or mutate the request query parameters.
	// The field would not be json encoded.
	RequestQueryParamResources map[string]resource.Resource `json:"-"`

	// RequestBodyResource is the resource representation of the request body.
	// It is used to generate or mutate the request body.
	// The field would not be json encoded.
	RequestBodyResource resource.Resource `json:"-"`

	// Energy is the energy of the operation case.
	// It is used to prioritize the operation cases.
	// The higher the energy, the higher the priority.
	Energy int `json:"energy"`

	// ExecutedCount is the number of times the test operation case is executed.
	// TODO: How to handle when copy the operation case? @xunzhou24
	ExecutedCount int `json:"executedCount"`

	// UUID is the unique identifier of the test operation case.
	UUID uuid.UUID `json:"uuid"`
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

	// UUID is the unique identifier of the test scenario.
	UUID uuid.UUID `json:"uuid"`
}

// NewTestScenario creates a new TestScenario.
// Before executing the test scenario, the test scenario should have empty request and response fields.
func NewTestScenario(operationCases []*OperationCase) *TestScenario {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		log.Err(err).Msg("[NewTestScenario] Failed to generate UUID")
	}
	return &TestScenario{
		OperationCases: operationCases,
		ExecutedCount:  0,
		Energy:         0,
		UUID:           newUUID,
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
	return http.IsStatusCodeSuccess(lastOperationCase.ResponseStatusCode)
}

// IncreaseEnergyByRandom increases the energy of the test scenario by a random value (normal distribution).
func (ts *TestScenario) IncreaseEnergyByRandom() {
	added := max(0, int(utils.NormInt64(ScenarioEnergyIncrMean, ScenarioEnergyIncrStdDev)))
	ts.Energy = min(ts.Energy+added, MaxScenarioEnergy)
}

// DecreaseEnergyByRandom decreases the energy of the test scenario by a random value (normal distribution).
func (ts *TestScenario) DecreaseEnergyByRandom() {
	subtracted := max(0, int(utils.NormInt64(ScenarioEnergyDecrMean, ScenarioEnergyDecrStdDev)))
	ts.Energy = max(ts.Energy-subtracted, MinScenarioEnergy)
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
		UUID:           ts.UUID,
	}
}

// Reset resets the test scenario.
// It resets the executed count and energy (of both scenario itself and its cases) to 0, and gives the test scenario a new UUID.
func (ts *TestScenario) Reset() {
	ts.ExecutedCount = 0
	ts.Energy = 0
	for _, operationCase := range ts.OperationCases {
		operationCase.Energy = 0
	}
	newUUID, err := uuid.NewRandom()
	if err != nil {
		log.Err(err).Msg("[TestScenario.Reset] Failed to generate UUID")
	}
	ts.UUID = newUUID
}

// AppendOperationCase appends an operation case to the test scenario.
func (ts *TestScenario) AppendOperationCase(operationCase *OperationCase) {
	ts.OperationCases = append(ts.OperationCases, operationCase)
}

// NewOperationCase creates a new OperationCase.
func NewOperationCase(
	apiMethod static.SimpleAPIMethod,
	operation *openapi3.Operation,
) *OperationCase {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		log.Err(err).Msg("[NewOperationCase] Failed to generate UUID")
	}
	return &OperationCase{
		APIMethod: apiMethod,
		Operation: operation,
		Energy:    0,
		ExecutedCount: 0,
		UUID: newUUID,
	}
}

// IsExecutedSuccessfully checks whether the operation case is executed successfully.
// It only checks the response status code for now.
func (oc *OperationCase) IsExecutedSuccessfully() bool {
	return http.IsStatusCodeSuccess(oc.ResponseStatusCode)
}

// Copy creates a deep copy of the operation case.
// TODO: deep copy the request and response body. @xunzhou24
func (oc *OperationCase) Copy() *OperationCase {
	// Copy the request and response headers, path parameters, query parameters, and body.
	requestHeaders := make(map[string]string)
	maps.Copy(requestHeaders, oc.RequestHeaders)
	requestPathParams := make(map[string]string)
	maps.Copy(requestPathParams, oc.RequestPathParams)
	requestQueryParams := make(map[string]string)
	maps.Copy(requestQueryParams, oc.RequestQueryParams)
	requestBody := make([]byte, len(oc.RequestBody))
	copy(requestBody, oc.RequestBody)
	responseHeaders := make(map[string]string)
	maps.Copy(responseHeaders, oc.ResponseHeaders)
	responseBody := make([]byte, len(oc.ResponseBody))
	copy(responseBody, oc.ResponseBody)

	// Copy resources.
	requestPathParamResources := make(map[string]resource.Resource)
	for k, v := range oc.RequestPathParamResources {
		requestPathParamResources[k] = v.Copy()
	}
	requestQueryParamResources := make(map[string]resource.Resource)
	for k, v := range oc.RequestQueryParamResources {
		requestQueryParamResources[k] = v.Copy()
	}
	var requestBodyResources resource.Resource
	if oc.RequestBodyResource != nil {
		requestBodyResources = oc.RequestBodyResource.Copy()
	} else {
		requestBodyResources = nil
	}

	return &OperationCase{
		APIMethod:          oc.APIMethod,
		Operation:          oc.Operation,
		RequestHeaders:     requestHeaders,
		RequestPathParams:  requestPathParams,
		RequestQueryParams: requestQueryParams,
		RequestBody:        requestBody,
		ResponseHeaders:    responseHeaders,
		ResponseStatusCode: oc.ResponseStatusCode,
		ResponseBody:       responseBody,

		RequestPathParamResources:  requestPathParamResources,
		RequestQueryParamResources: requestQueryParamResources,
		RequestBodyResource:        requestBodyResources,

		Energy:                   oc.Energy,
		ExecutedCount:            oc.ExecutedCount,
		UUID:                     oc.UUID,
	}
}

// Reset resets the test operation case.
// It resets the executed count and energy to 0, and gives the test operation case a new UUID.
func (oc *OperationCase) Reset() {
	oc.ExecutedCount = 0
	oc.Energy = 0
	newUUID, err := uuid.NewRandom()
	if err != nil {
		log.Err(err).Msg("[OperationCase.Reset] Failed to generate UUID")
	}
	oc.UUID = newUUID
}

// SetRequestPathParamsByResources sets the request path parameters by the given resources.
// It stores the resources in the RequestPathParamResources field,
// and sets the RequestPathParams field to the string representation of the resources.
func (oc *OperationCase) SetRequestPathParamsByResources(resources map[string]resource.Resource) {
	requestPathParams := make(map[string]string)
	for key, resrc := range resources {
		var valueStr string
		// For array type, we need to convert the array to string.
		// For example, if the array is [1, 2, 3], we convert it to "1,2,3", instead of json-style (e.g., "[1,2,3]").
		if resrc.Typ() == static.SimpleAPIPropertyTypeArray {
			valueList := make([]string, 0)
			for _, v := range resrc.(*resource.ResourceArray).Value {
				vStr := v.String()
				valueList = append(valueList, vStr)
			}
			valueStr = strings.Join(valueList, ",")
		} else {
			valueStr = resrc.String()
		}
		requestPathParams[key] = valueStr
	}
	oc.RequestPathParams = requestPathParams
	oc.RequestPathParamResources = resources
}

// SetRequestQueryParamsByResources sets the request query parameters by the given resources.
// It stores the resources in the RequestQueryParamResources field,
// and sets the RequestQueryParams field to the string representation of the resources.
func (oc *OperationCase) SetRequestQueryParamsByResources(resources map[string]resource.Resource) {
	requestQueryParams := make(map[string]string)
	for key, resrc := range resources {
		var valueStr string
		// For array type, we need to convert the array to string.
		// For example, if the array is [1, 2, 3], we convert it to "1,2,3", instead of json-style (e.g., "[1,2,3]").
		if resrc.Typ() == static.SimpleAPIPropertyTypeArray {
			valueList := make([]string, 0)
			for _, v := range resrc.(*resource.ResourceArray).Value {
				vStr := v.String()
				valueList = append(valueList, vStr)
			}
			valueStr = strings.Join(valueList, ",")
		} else {
			valueStr = resrc.String()
		}
		requestQueryParams[key] = valueStr
	}
	oc.RequestQueryParams = requestQueryParams
	oc.RequestQueryParamResources = resources
}

// SetRequestBodyByResource sets the request body by the given resource.
// It stores the resource in the RequestBodyResources field,
// and sets the RequestBody field to the string representation of the resource.
func (oc *OperationCase) SetRequestBodyByResource(resource resource.Resource) {
	oc.RequestBodyResource = resource
	if resource == nil {
		return
	}
	// Convert the resource to json string.
	jsonStr := resource.String()
	oc.RequestBody = []byte(jsonStr)
}

// IncreaseEnergyByRandom increases the energy of the test operation case by a random value (normal distribution).
func (oc *OperationCase) IncreaseEnergyByRandom() {
	added := max(0, int(utils.NormInt64(OperationCaseEnergyIncrMean, OperationCaseEnergyIncrStdDev)))
	oc.Energy = min(oc.Energy+added, MaxOperationCaseEnergy)
}

// DecreaseEnergyByRandom decreases the energy of the test operation case by a random value (normal distribution).
func (oc *OperationCase) DecreaseEnergyByRandom() {
	subtracted := max(0, int(utils.NormInt64(OperationCaseEnergyDecrMean, OperationCaseEnergyDecrStdDev)))
	oc.Energy = max(oc.Energy-subtracted, MinOperationCaseEnergy)
}
