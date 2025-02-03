package casemanager

import (
	"fmt"
	"resttracefuzzer/pkg/resource"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"

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
}

// NewCaseManager creates a new CaseManager.
func NewCaseManager(APIManager *static.APIManager, resourceManager *resource.ResourceManager) *CaseManager {
	testScenarios := make([]*TestScenario, 0)
	m := &CaseManager{
		APIManager:      APIManager,
		ResourceManager: resourceManager,
		TestScenarios:   testScenarios,
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
	// TODO: Implement this method. @xunzhou24
	for _, operationCase := range testScenario.OperationCases {
		log.Debug().Msgf("[CaseManager.PopAndPopulate] Start to populate request for operation %v", operationCase.APIMethod)
		// fill the request params
		requestParamsDef := operationCase.Operation.Parameters
		if requestParamsDef != nil {
			requestParams, err := m.generateRequestParamsFromSchema(requestParamsDef)
			if err != nil {
				log.Err(err).Msg("[CaseManager.PopAndFillRequest] Failed to generate request params")
				return nil, err
			}
			operationCase.RequestParams = requestParams
		}

		// fill the request headers
		// TODO: Implement this method. @xunzhou24

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
func (m *CaseManager) generateRequestBodyFromSchema(requestBodyRef *openapi3.RequestBodyRef) (map[string]interface{}, error) {
	if requestBodyRef == nil || requestBodyRef.Value == nil {
		return nil, fmt.Errorf("request body is nil")
	}
	return m.generateObjectFromSchema(requestBodyRef.Value.Content.Get("application/json").Schema)
}

// generateRequestParamsFromSchema generates request params from a schema.
// It returns a map of request params, and error if any.
func (m *CaseManager) generateRequestParamsFromSchema(params []*openapi3.ParameterRef) (map[string]string, error) {
	result := make(map[string]string)
	for _, param := range params {
		if param == nil || param.Value == nil {
			return nil, fmt.Errorf("request param is nil")
		}
		// TODO: format string to param format @xunzhou24
		// e.g. ['a', 'b'] => 'a,b'
		generatedObject, err := m.generateObjectFromSchema(param.Value.Schema)
		if err != nil {
			log.Err(err).Msgf("[CaseManager.generateRequestParamsFromSchema] Failed to generate object from schema %v", param.Value.Schema)
			return nil, err
		}
		valueStr, err := sonic.MarshalString(generatedObject)
		if err != nil {
			log.Err(err).Msgf("[CaseManager.generateRequestParamsFromSchema] Failed to marshal object to string %v", generatedObject)
			return nil, err
		}
		result[param.Value.Name] = valueStr
	}
	return result, nil
}

// generateObjectFromSchema generates a json object from a schema.
// It returns a json object, and error if any.
//
// TODO: Implement strategies
func (m *CaseManager) generateObjectFromSchema(schema *openapi3.SchemaRef) (map[string]interface{}, error) {
	if schema == nil || schema.Value == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	result := make(map[string]interface{})

	for propName, propSchema := range schema.Value.Properties {
		switch {
		case propSchema.Value.Type.Includes("object"):
			subResult, err := m.generateObjectFromSchema(propSchema)
			if err != nil {
				return nil, err
			}
			result[propName] = subResult

		case propSchema.Value.Type.Includes("array"):
			subResult, err := m.generateObjectFromSchema(propSchema.Value.Items)
			if err != nil {
				return nil, err
			}
			// TODO: control the array size @xunzhou24
			result[propName] = []interface{}{subResult}

		default:
			// primitive types
			result[propName] = utils.GenerateDefaultValueForPrimitiveSchemaType(propSchema.Value.Type)
		}
	}
	return result, nil
}
