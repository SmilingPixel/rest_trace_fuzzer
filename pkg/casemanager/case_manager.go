package casemanager

import (
	"fmt"
	"resttracefuzzer/pkg/resource"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/strategy"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

const (
	// MaxExecutedTimes is the default maximum executed times of a test scenario.
	MaxExecutedTimes = 3
)

// CaseManager manages the test cases.
type CaseManager struct {
	// TestScenarios is a list of test scenarios.
	TestScenarios []*TestScenario

	// The API manager.
	APIManager *static.APIManager

	// The resource manager.
	ResourceManager *resource.ResourceManager

	// The fuzz strategist.
	FuzzStrategist *strategy.FuzzStrategist

	// GlobalExtraHeaders is the global extra headers, which will be added to each request.
	// It is a map of header name to header value.
	// It can be used for simple cases, e.g., adding an authorization header.
	GlobalExtraHeaders map[string]string
}

// NewCaseManager creates a new CaseManager.
func NewCaseManager(APIManager *static.APIManager, resourceManager *resource.ResourceManager, fuzzStrategist *strategy.FuzzStrategist, globalExtraHeaders map[string]string) *CaseManager {
	testScenarios := make([]*TestScenario, 0)
	m := &CaseManager{
		APIManager:      APIManager,
		ResourceManager: resourceManager,
		FuzzStrategist:  fuzzStrategist,
		TestScenarios:   testScenarios,
		GlobalExtraHeaders: globalExtraHeaders,
	}
	m.initTestcasesFromDoc()
	return m
}

// Pop pops a test scenario of highest priority from the queue.
func (m *CaseManager) Pop() (*TestScenario, error) {
	// TODO: Implement this method. @xunzhou24
	// We select the first test scenario for now.
	if len(m.TestScenarios) == 0 {
		log.Error().Msg("[CaseManager.Pop] No test scenario available")
		return nil, fmt.Errorf("no test scenario available")
	}
	testScenario := m.TestScenarios[0]
	m.TestScenarios = m.TestScenarios[1:]
	return testScenario, nil
}

// PopAndPopulate pops a test scenario of highest priority from the case manager
// and populates the request part, including the headers, params and request body.
func (m *CaseManager) PopAndPopulate() (*TestScenario, error) {
	testScenario, err := m.Pop()
	if err != nil {
		log.Err(err).Msg("[CaseManager.PopAndFillRequest] Failed to pop a test scenario")
		return nil, err
	}

	for _, operationCase := range testScenario.OperationCases {
		log.Debug().Msgf("[CaseManager.PopAndPopulate] Start to populate request for operation %v", operationCase.APIMethod)
		// fill the request path and query params
		requestParamsDef := operationCase.Operation.Parameters
		if requestParamsDef != nil {
			requestPathParams, requestQueryParams, err := m.generateRequestParamsFromSchema(requestParamsDef)
			if err != nil {
				log.Err(err).Msg("[CaseManager.PopAndFillRequest] Failed to generate request params")
				return nil, err
			}
			operationCase.RequestPathParams = requestPathParams
			operationCase.RequestQueryParams = requestQueryParams
		}

		// fill the request headers, including global extra headers and operation specific headers
		requestHeaders := make(map[string]string)
		// Add global extra headers
		for k, v := range m.GlobalExtraHeaders {
			requestHeaders[k] = v
		}
		// Add operation specific headers
		// TODO: Implement this. @xunzhou24
		operationCase.RequestHeaders = requestHeaders

		// fill the request body
		requestBodySchema := operationCase.Operation.RequestBody
		if requestBodySchema != nil {
			requestBody, err := m.generateRequestBodyFromSchema(requestBodySchema)
			if err != nil {
				log.Err(err).Msg("[CaseManager.PopAndFillRequest] Failed to generate request body")
				return nil, err
			}
			operationCase.RequestBody = requestBody
		}
	}
	return testScenario, nil
}

// Push adds a test case to the case manager.
func (m *CaseManager) Push(testcase *TestScenario) {
	m.TestScenarios = append(m.TestScenarios, testcase)
}

// EvaluateScenarioAndTryUpdate evaluates the given metrics for the given test scenario that has been executed,
// determines whether to put the scenario back to the queue, and processes the scenario to generate a new one if needed.
// It returns an error if any.
func (m *CaseManager) EvaluateScenarioAndTryUpdate(hasAchieveNewCoverage bool, executedScenario *TestScenario) error {
	// Update the executed times
	executedScenario.ExecutedTimes++

	// Put the scenario back to the queue if it has achieved new coverage or has not been executed for enough times
	if hasAchieveNewCoverage || executedScenario.ExecutedTimes < MaxExecutedTimes {
		// Put the scenario back to the queue
		m.Push(executedScenario)
	}

	// Process the scenario to generate a new one
	newScenario, err := m.extendScenarioIfExecSuccess(executedScenario)
	if err != nil {
		log.Err(err).Msg("[CaseManager.evaluateScenarioAndTryUpdate] Failed to process scenario")
		return err
	}
	if newScenario != nil {
		m.Push(newScenario)
	}

	return nil
}

// extendScenarioIfExecSuccess processes a test scenario to extend it, if needed, and if all proceeding operations are executed successfully.
func (m *CaseManager) extendScenarioIfExecSuccess(existingScenario *TestScenario) (*TestScenario, error) {
	// This might involve modifying request parameters, headers, body, etc.
	if !existingScenario.IsExecutedSuccessfully() {
		log.Warn().Msg("[CaseManager.extendScenarioIfExecSuccess] The existing scenario is not executed successfully")
		return nil, nil
	}

	// copy the existing scenario
	newScenario := existingScenario.Copy()
	newScenario.Reset()
	// TODO: append a new operation to the scenario @xunzhou24
	return newScenario, nil
}

// Init initializes the case queue.
func (m *CaseManager) initTestcasesFromDoc() error {
	// TODO: Implement this method. @xunzhou24
	// At the beginning, each testcase is a simple request to each API.
	for method, operation := range m.APIManager.APIMap {
		operationCase := OperationCase{
			Operation: operation,
			APIMethod: method,
		}
		testcase := NewTestScenario([]*OperationCase{&operationCase})
		m.Push(testcase)
	}
	return nil
}

// generateRequestBodyFromSchema generates a request body from a schema.
// It returns a json object, and error if any.
func (m *CaseManager) generateRequestBodyFromSchema(requestBodyRef *openapi3.RequestBodyRef) (interface{}, error) {
	if requestBodyRef == nil || requestBodyRef.Value == nil {
		return nil, fmt.Errorf("request body is nil")
	}
	return m.FuzzStrategist.GenerateValueForSchema(requestBodyRef.Value.Content.Get("application/json").Schema)
}

// generateRequestParamsFromSchema generates request params from a schema.
// It returns a map of request (path) params, a map of query params, and an error if any.
func (m *CaseManager) generateRequestParamsFromSchema(params []*openapi3.ParameterRef) (map[string]string, map[string]string, error) {
	pathParams := make(map[string]string)
	queryParams := make(map[string]string)
	for _, param := range params {
		if param == nil || param.Value == nil {
			return nil, nil, fmt.Errorf("request param is nil")
		}

		generatedValue, err := m.FuzzStrategist.GenerateValueForSchema(param.Value.Schema)
		if err != nil {
			log.Err(err).Msgf("[CaseManager.generateRequestParamsFromSchema] Failed to generate object from schema %v", param.Value.Schema)
			return nil, nil, err
		}

		var valueStr string
		// For array type, we need to convert the array to string.
		// For example, if the array is [1, 2, 3], we convert it to "1,2,3", instead of json-style (e.g., "[1,2,3]").
		if generatedArray, ok := generatedValue.([]interface{}); ok {
			valueList := make([]string, 0)
			for _, v := range generatedArray {
				vStr, err := sonic.MarshalString(v)
				if err != nil {
					log.Err(err).Msgf("[CaseManager.generateRequestParamsFromSchema] Failed to marshal object to string %v", v)
					return nil, nil, err
				}
				valueList = append(valueList, vStr)
			}
			valueStr = strings.Join(valueList, ",")
		} else {
			valueStr, err = sonic.MarshalString(generatedValue)
			if err != nil {
				log.Err(err).Msgf("[CaseManager.generateRequestParamsFromSchema] Failed to marshal object to string %v", generatedValue)
				return nil, nil, err
			}
		}
		
		if param.Value.In == "path" {
			pathParams[param.Value.Name] = valueStr
		} else if param.Value.In == "query" {
			queryParams[param.Value.Name] = valueStr
		} else {
			// TODO: support other param locations (e.g., header) @xunzhou24
			log.Warn().Msgf("[CaseManager.generateRequestParamsFromSchema] Unsupported param location %v", param.Value.In)
		}
	}
	return pathParams, queryParams, nil
}
