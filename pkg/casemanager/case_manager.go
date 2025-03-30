package casemanager

import (
	"fmt"
	"resttracefuzzer/pkg/resource"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/strategy"
	"sort"

	"maps"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

const (
	// MaxAllowedExecutedCount is the default maximum executed times of a test scenario.
	MaxAllowedExecutedCount = 3

	// MaxAllowedScenarios is the default maximum number of test scenarios in the queue.
	MaxAllowedScenarios = 114
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

	// The mutation strategist.
	ResourceMutateStrategy *strategy.ResourceMutateStrategy

	// GlobalExtraHeaders is the global extra headers, which will be added to each request.
	// It is a map of header name to header value.
	// It can be used for simple cases, e.g., adding an authorization header.
	GlobalExtraHeaders map[string]string
}

// NewCaseManager creates a new CaseManager.
func NewCaseManager(
	APIManager *static.APIManager,
	resourceManager *resource.ResourceManager,
	fuzzStrategist *strategy.FuzzStrategist,
	ResourceMutateStrategy *strategy.ResourceMutateStrategy,
	globalExtraHeaders map[string]string,
) *CaseManager {
	testScenarios := make([]*TestScenario, 0)
	m := &CaseManager{
		APIManager:             APIManager,
		ResourceManager:        resourceManager,
		FuzzStrategist:         fuzzStrategist,
		ResourceMutateStrategy: ResourceMutateStrategy,
		TestScenarios:          testScenarios,
		GlobalExtraHeaders:     globalExtraHeaders,
	}
	m.initTestcasesFromDoc()
	return m
}

// Pop pops a test scenario of highest priority from the queue.
func (m *CaseManager) Pop() (*TestScenario, error) {
	// TODO: Implement this method with different strategies. @xunzhou24
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
		requestPathParamResources, requestQueryParamResources, err := m.generateRequestParamResourcesFromSchema(requestParamsDef)
		if err != nil {
			log.Err(err).Msg("[CaseManager.PopAndFillRequest] Failed to generate request param resources")
			return nil, err
		}
		operationCase.SetRequestPathParamsByResources(requestPathParamResources)
		operationCase.SetRequestQueryParamsByResources(requestQueryParamResources)

		// fill the request headers, including global extra headers and operation specific headers
		requestHeaders := make(map[string]string)
		// Add global extra headers
		maps.Copy(requestHeaders, m.GlobalExtraHeaders)
		// Add operation specific headers
		operationCase.RequestHeaders = requestHeaders

		// fill the request body
		requestBodySchema := operationCase.Operation.RequestBody
		if requestBodySchema != nil {
			requestBodyResrc, err := m.generateRequestBodyResourceFromSchema(requestBodySchema)
			if err != nil {
				log.Err(err).Msg("[CaseManager.PopAndFillRequest] Failed to generate request body resource")
				return nil, err
			}
			operationCase.SetRequestBodyByResource(requestBodyResrc)
		}
	}
	return testScenario, nil
}

// pushAndSort pushes a test scenario to the case manager and sorts the test scenarios by energy.
// It also culls the test scenarios if there are too many.
func (m *CaseManager) pushAndSort(testcase *TestScenario) {
	m.push(testcase)
	m.sortAndCullByEnergy()
}

// push adds a test case to the case manager.
func (m *CaseManager) push(testcase *TestScenario) {
	m.TestScenarios = append(m.TestScenarios, testcase)
}

// sortAndCullByEnergy sorts the test scenarios by energy and culls the test scenarios if there are too many.
func (m *CaseManager) sortAndCullByEnergy() {
	sort.Slice(m.TestScenarios, func(i, j int) bool {
		return m.TestScenarios[i].Energy > m.TestScenarios[j].Energy
	})

	if len(m.TestScenarios) > MaxAllowedScenarios {
		m.TestScenarios = m.TestScenarios[:MaxAllowedScenarios]
	}
}

// EvaluateScenarioAndTryUpdate evaluates the given metrics for the given test scenario that has been executed,
// determines whether to put the scenario back to the queue, and expand the scenario with an operation to a new scenario if needed.
// It returns an error if any.
func (m *CaseManager) EvaluateScenarioAndTryUpdate(hasAchieveNewCoverage bool, executedScenario *TestScenario) error {
	// Update the executed count and energy
	executedScenario.ExecutedCount++
	if hasAchieveNewCoverage {
		executedScenario.IncreaseEnergyByRandom()
	} else {
		executedScenario.DecreaseEnergyByRandom()
	}

	// If it has achieved new coverage or has not been executed for enough times,
	// mutate it and put it back to the queue.
	if hasAchieveNewCoverage || executedScenario.ExecutedCount < MaxAllowedExecutedCount {
		mutatedScenario, err := m.mutateScenario(executedScenario)
		if err != nil {
			log.Err(err).Msg("[CaseManager.evaluateScenarioAndTryUpdate] Failed to mutate scenario")
			return err
		}
		m.pushAndSort(mutatedScenario)
	}

	// Extend the scenario to generate a new one
	newScenario, err := m.extendScenarioIfExecSuccess(executedScenario)
	if err != nil {
		log.Err(err).Msg("[CaseManager.evaluateScenarioAndTryUpdate] Failed to process scenario")
		return err
	}
	if newScenario != nil {
		m.pushAndSort(newScenario)
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
	// The new one will inherit the energy from the existing one.
	newScenario.Energy = existingScenario.Energy / 2
	// We append a random operation to the scenario for now.
	// TODO: append an operation based on API dependency @xunzhou24
	newAPIMethod, newOperation := m.APIManager.GetRandomAPIMethod()
	newOperationCase := OperationCase{
		Operation: newOperation,
		APIMethod: newAPIMethod,
	}
	newScenario.OperationCases = append(newScenario.OperationCases, &newOperationCase)
	return newScenario, nil
}

// Init initializes the case queue.
func (m *CaseManager) initTestcasesFromDoc() error {
	// TODO: Implement this method. @xunzhou24
	// At the beginning, each testcase is a simple request to each API.
	for method, operation := range m.APIManager.APIMap {
		operationCase := NewOperationCase(method, operation)
		testcase := NewTestScenario([]*OperationCase{operationCase})
		m.pushAndSort(testcase)
	}
	return nil
}

// mutateScenario mutates the given test scenario and returns it.
// Mutation would not reset the scenario, i.e., the executed count and energy will be inherited from the existing one. (This is different from extending)
func (m *CaseManager) mutateScenario(scenario *TestScenario) (*TestScenario, error) {
	// copy the given scenario
	newScenario := scenario.Copy()
	// The new one will inherit the energy from the existing one.
	newScenario.Energy = scenario.Energy / 2

	// For each operation in the scenario, we mutate the request params, headers, and body.
	for _, operationCase := range newScenario.OperationCases {
		requestPathParamResrc := operationCase.RequestPathParamResources
		requestQueryParamResrc := operationCase.RequestQueryParamResources
		requestBodyResrc := operationCase.RequestBodyResource
		// mutate the request path params
		for key, resrc := range requestPathParamResrc {
			mutatedResrc, err := m.ResourceMutateStrategy.MutateResource(resrc)
			if err != nil {
				log.Err(err).Msgf("[CaseManager.mutateScenario] Failed to mutate request path param %v", key)
				return nil, err
			}
			requestPathParamResrc[key] = mutatedResrc
		}
		// mutate the request query params
		for key, resrc := range requestQueryParamResrc {
			mutatedResrc, err := m.ResourceMutateStrategy.MutateResource(resrc)
			if err != nil {
				log.Err(err).Msgf("[CaseManager.mutateScenario] Failed to mutate request query param %v", key)
				return nil, err
			}
			requestQueryParamResrc[key] = mutatedResrc
		}
		// mutate the request body
		if requestBodyResrc != nil {
			mutatedResrc, err := m.ResourceMutateStrategy.MutateResource(requestBodyResrc)
			if err != nil {
				log.Err(err).Msgf("[CaseManager.mutateScenario] Failed to mutate request body")
				return nil, err
			}
			requestBodyResrc = mutatedResrc
		}
		// set the mutated resources back to the operation case
		// use `Set...ByResource` to set the actual request params at the same time
		operationCase.SetRequestPathParamsByResources(requestPathParamResrc)
		operationCase.SetRequestQueryParamsByResources(requestQueryParamResrc)
		operationCase.SetRequestBodyByResource(requestBodyResrc)
	}
	return newScenario, nil
}


// generateRequestBodyResourceFromSchema generates a request body resource from a schema.
// It returns a json object as a resource and error if any.
// If the schema is empty, it returns nil.
func (m *CaseManager) generateRequestBodyResourceFromSchema(requestBodyRef *openapi3.RequestBodyRef) (resource.Resource, error) {
	if requestBodyRef == nil || requestBodyRef.Value == nil {
		return nil, nil
	}
	generatedValue, err := m.FuzzStrategist.GenerateValueForSchema(requestBodyRef.Ref, requestBodyRef.Value.Content.Get("application/json").Schema)
	if err != nil {
		log.Err(err).Msgf("[CaseManager.generateRequestBodyResourceFromSchema] Failed to generate object from schema %v", requestBodyRef.Value.Content.Get("application/json").Schema)
		return nil, err
	}
	return generatedValue, nil
}

// generateRequestParamResourcesFromSchema generates request params resources (including path and query) from a schema.
// It returns a map of request path params, a map of query params, and an error if any.
func (m *CaseManager) generateRequestParamResourcesFromSchema(params []*openapi3.ParameterRef) (map[string]resource.Resource, map[string]resource.Resource, error) {
	pathParams := make(map[string]resource.Resource)
	queryParams := make(map[string]resource.Resource)
	for _, param := range params {
		if param == nil || param.Value == nil {
			return nil, nil, fmt.Errorf("request param is nil")
		}

		generatedValue, err := m.FuzzStrategist.GenerateValueForSchema(param.Value.Name, param.Value.Schema)
		if err != nil {
			log.Err(err).Msgf("[CaseManager.generateRequestParamResourcesFromSchema] Failed to generate object from schema %v", param.Value.Schema)
			return nil, nil, err
		}

		if param.Value.In == "path" {
			pathParams[param.Value.Name] = generatedValue
		} else if param.Value.In == "query" {
			queryParams[param.Value.Name] = generatedValue
		} else {
			// TODO: support other param locations (e.g., header) @xunzhou24
			log.Warn().Msgf("[CaseManager.generateRequestParamResourcesFromSchema] Unsupported param location %v", param.Value.In)
		}
	}
	return pathParams, queryParams, nil
}
