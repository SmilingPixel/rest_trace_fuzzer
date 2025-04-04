package casemanager

import (
	"fmt"
	"resttracefuzzer/internal/config"
	"resttracefuzzer/pkg/resource"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/strategy"
	"sort"

	"maps"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	"slices"
)

const (
	// MaxAllowedScenarioExecutedCount is the default maximum executed times of a test scenario.
	MaxAllowedScenarioExecutedCount = 3

	// MaxAllowedOperationCaseExecutedCount is the default maximum executed times of a test operation.
	MaxAllowedOperationCaseExecutedCount = 7

	// MaxAllowedScenarios is the default maximum number of test scenarios in the queue.
	MaxAllowedScenarios = 114

	// MaxAllowedOperationCases is the default maximum number of test operation cases in the queue of the API method.
	MaxAllowedOperationCases = 7
)

// CaseManager manages the test cases.
type CaseManager struct {
	// TestScenarios is a list of test scenarios.
	// In each loop of testing, a test scenario will be popped from the queue and executed.
	// The test scenario is a combination of one or multiple operations.
	TestScenarios []*TestScenario

	// TestOperationCaseQueueMap is a map of test operation case queue.
	// The key is the API method, and the value is a list of test operation cases.
	// Only cases that have already been executed will be put into its queue.
	// So each operation case should have a non-empty request parameters, and energy as well.
	// It is used to get a new operation case to execute when extending a test scenario.
	TestOperationCaseQueueMap map[static.SimpleAPIMethod][]*OperationCase

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
	testOperationCaseQueueMap := make(map[static.SimpleAPIMethod][]*OperationCase)
	m := &CaseManager{
		APIManager:             APIManager,
		ResourceManager:        resourceManager,
		FuzzStrategist:         fuzzStrategist,
		ResourceMutateStrategy: ResourceMutateStrategy,
		TestScenarios:          testScenarios,
		GlobalExtraHeaders:     globalExtraHeaders,
		TestOperationCaseQueueMap: testOperationCaseQueueMap,
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
				log.Err(err).Msgf("[CaseManager.PopAndFillRequest] Failed to generate request body resource, scenario UUID: %s", testScenario.UUID.String())
				return nil, err
			}
			operationCase.SetRequestBodyByResource(requestBodyResrc)
		}
	}
	return testScenario, nil
}

// pushAndSort pushes a test scenario to the case manager and sorts the test scenarios by energy (if energy function is enabled in config).
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
// If energy function is not enabled in config, it only culls the test scenarios.
func (m *CaseManager) sortAndCullByEnergy() {
	if config.GlobalConfig.EnableEnergyScenario {
		sort.Slice(m.TestScenarios, func(i, j int) bool {
			return m.TestScenarios[i].Energy > m.TestScenarios[j].Energy
		})
	}

	if len(m.TestScenarios) > MaxAllowedScenarios {
		m.TestScenarios = m.TestScenarios[:MaxAllowedScenarios]
	}
}

// pushAndSortOperationCase pushes a test operation case to the case manager and sorts the test operation cases by energy (if energy function is enabled in config).
// It also culls the test operation cases if there are too many.
func (m *CaseManager) pushAndSortOperationCase(operationCase *OperationCase) {
	m.pushOperationCase(operationCase)
	m.sortAndCullOperationCaseByEnergy()
}

// pushOperationCase adds a test operation case to the case manager.
func (m *CaseManager) pushOperationCase(testcase *OperationCase) {
	// Get the API method of the operation case.
	apiMethod := testcase.APIMethod
	// Get the queue of the API method.
	operationCaseQueue, exist := m.TestOperationCaseQueueMap[apiMethod]
	if !exist {
		operationCaseQueue = make([]*OperationCase, 0)
	}
	// Add the operation case to the queue.
	// Check if the operation case is already in the queue before adding it.
	for _, operationCase := range operationCaseQueue {
		if operationCase.UUID == testcase.UUID {
			return
		}
	}
	operationCaseQueue = append(operationCaseQueue, testcase)
	// Update the queue in the map.
	m.TestOperationCaseQueueMap[apiMethod] = operationCaseQueue
}

// sortAndCullOperationCaseByEnergy sorts the test operation cases by energy and culls the test operation cases if there are too many.
func (m *CaseManager) sortAndCullOperationCaseByEnergy() {
	for apiMethod, operationCaseQueue := range m.TestOperationCaseQueueMap {
		if config.GlobalConfig.EnableEnergyScenario {
			sort.Slice(operationCaseQueue, func(i, j int) bool {
				return operationCaseQueue[i].Energy > operationCaseQueue[j].Energy
			})
		}
		if len(operationCaseQueue) > MaxAllowedOperationCases {
			m.TestOperationCaseQueueMap[apiMethod] = operationCaseQueue[:MaxAllowedOperationCases]
		}
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
	if hasAchieveNewCoverage || executedScenario.ExecutedCount < MaxAllowedScenarioExecutedCount {
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

// EvaluateOperationCaseAndTryUpdate evaluates the given metrics for the given test operation case that has been executed,
// determines whether to put the operation to the queue.
// It returns an error if any.
func (m *CaseManager) EvaluateOperationCaseAndTryUpdate(hasAchieveNewCoverage bool, executedOperationCase *OperationCase) error {
	// Update the executed count and energy
	executedOperationCase.ExecutedCount++
	if hasAchieveNewCoverage {
		executedOperationCase.IncreaseEnergyByRandom()
	} else {
		executedOperationCase.DecreaseEnergyByRandom()
	}

	// If it has achieved new coverage or has not been executed for enough times,
	// put it to the queue.
	if hasAchieveNewCoverage || executedOperationCase.ExecutedCount < MaxAllowedOperationCaseExecutedCount {
		m.pushAndSortOperationCase(executedOperationCase)
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

	// Check if the existing scenario has reached the maximum number of operations.
	if len(existingScenario.OperationCases) >= config.GlobalConfig.MaxOpsPerScenario {
		log.Debug().Msgf("[CaseManager.extendScenarioIfExecSuccess] The existing scenario (UUID: %s) has reached the maximum number of operations", existingScenario.UUID.String())
		return nil, nil
	}

	// copy the existing scenario
	newScenario := existingScenario.Copy()
	newScenario.Reset()
	// The new one will inherit the energy from the existing one.
	newScenario.Energy = existingScenario.Energy / 2
	
	// Append a new operation.
	// We will try to get a new API method based on producer-consumer relationship. If there is no consumer, we will randomly select an API method.
	// When generating a new operation case, we will try to get a operation from operation case queue (which is sorted by energy in advance).
	var candidateAPIMethods []static.SimpleAPIMethod
	producers := make([]static.SimpleAPIMethod, 0)
	for _, operationCase := range newScenario.OperationCases {
		producers = append(producers, operationCase.APIMethod)
	}
	// Get the consumers of the producers in the existing scenario.
	consumers := m.APIManager.GetConsumerAPIMethodsByProducers(producers)
	// If there are no consumers, we can randomly select an API method.
	// Otherwise, we select the consumers.
	if len(consumers) > 0 {
		candidateAPIMethods = consumers
	} else {
		candidateAPIMethods = append(candidateAPIMethods, m.APIManager.GetRandomAPIMethod())
	}

	if len(candidateAPIMethods) == 0 {
		log.Warn().Msg("[CaseManager.extendScenarioIfExecSuccess] No candidates available for extending the scenario")
		return nil, nil
	}

	// Generate operation cases and select one from the candidates.
	candidateOperationCases := make([]*OperationCase, 0)
	for _, apiMethod := range candidateAPIMethods {
		var operationCase *OperationCase
		// First try to get the operation case from the queue.
		operationCaseQueue, exist := m.TestOperationCaseQueueMap[apiMethod]
		if exist && len(operationCaseQueue) > 0 {
			// Get the first operation case (whose energy is the highest) from the queue.
			// As it is picked from the queue only as a candidate, we do not remove it from the queue right now.
			operationCase = operationCaseQueue[0]
		} else {
			// If the queue is empty, we need to create a new operation case.
			operation, exist := m.APIManager.GetOperationByMethod(apiMethod)
			if !exist {
				log.Warn().Msgf("[CaseManager.extendScenarioIfExecSuccess] The API method %v does not exist in the API manager", apiMethod)
				continue
			}
			operationCase = NewOperationCase(apiMethod, operation)
		}
		// Add the operation case to the candidate operation cases.
		candidateOperationCases = append(candidateOperationCases, operationCase)
	}

	if len(candidateOperationCases) == 0 {
		log.Warn().Msg("[CaseManager.extendScenarioIfExecSuccess] No candidates available for extending the scenario")
		return nil, nil
	}
	// Select the operation case with the highest energy from the candidate operation cases.
	sort.Slice(candidateOperationCases, func(i, j int) bool {
		return candidateOperationCases[i].Energy > candidateOperationCases[j].Energy
	})
	newOperationCase := candidateOperationCases[0]
	
	// If the operation is selected from the queue, we need to remove it from the queue (We can check it by checking its UUID).
	// In addition, considering that the operations in queue have all been executed before, we should do some mutation.
	selectedAPIMethod := newOperationCase.APIMethod
	operationCaseQueue, exist := m.TestOperationCaseQueueMap[selectedAPIMethod]
	if exist {
		for i, operationCase := range operationCaseQueue {
			if operationCase.UUID == newOperationCase.UUID {
				// Remove the operation case from the queue.
				operationCaseQueue = slices.Delete(operationCaseQueue, i, i+1)
				// Mutate the operation case.
				mutatedOperationCase, err := m.mutateOperationCase(newOperationCase)
				if err != nil {
					log.Err(err).Msgf("[CaseManager.extendScenarioIfExecSuccess] Failed to mutate operation case %v", newOperationCase.APIMethod)
					return nil, err
				}
				newOperationCase = mutatedOperationCase
				
				break
			}
		}
		m.TestOperationCaseQueueMap[selectedAPIMethod] = operationCaseQueue
	}

	newScenario.OperationCases = append(newScenario.OperationCases, newOperationCase)
	return newScenario, nil
}

// Init initializes the case queue.
func (m *CaseManager) initTestcasesFromDoc() error { 
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

	// For each operation in the scenario, we mutate the request params, headers, and body.
	for i, operationCase := range newScenario.OperationCases {
		mutatedOperationCase, err := m.mutateOperationCase(operationCase)
		if err != nil {
			log.Err(err).Msgf("[CaseManager.mutateScenario] Failed to mutate operation case %v", operationCase.APIMethod)
			return nil, err
		}
		newScenario.OperationCases[i] = mutatedOperationCase
	}
	return newScenario, nil
}

// mutateOperationCase mutates the given test operation case and returns it.
// Mutation would not reset the operation case, i.e., the executed count and energy will be inherited from the existing one.
func (m *CaseManager) mutateOperationCase(operationCase *OperationCase) (*OperationCase, error) {
	// copy the given operation case
	newOperationCase := operationCase.Copy()

	requestPathParamResrc := newOperationCase.RequestPathParamResources
	requestQueryParamResrc := newOperationCase.RequestQueryParamResources
	requestBodyResrc := newOperationCase.RequestBodyResource
	// mutate the request path params
	for key, resrc := range requestPathParamResrc {
		mutatedResrc, err := m.ResourceMutateStrategy.MutateResource(resrc)
		if err != nil {
			log.Err(err).Msgf("[CaseManager.mutateOperationCase] Failed to mutate request path param %v", key)
			return nil, err
		}
		requestPathParamResrc[key] = mutatedResrc
	}
	// mutate the request query params
	for key, resrc := range requestQueryParamResrc {
		mutatedResrc, err := m.ResourceMutateStrategy.MutateResource(resrc)
		if err != nil {
			log.Err(err).Msgf("[CaseManager.mutateOperationCase] Failed to mutate request query param %v", key)
			return nil, err
		}
		requestQueryParamResrc[key] = mutatedResrc
	}
	// mutate the request body
	if requestBodyResrc != nil {
		mutatedResrc, err := m.ResourceMutateStrategy.MutateResource(requestBodyResrc)
		if err != nil {
			log.Err(err).Msgf("[CaseManager.mutateOperationCase] Failed to mutate request body")
			return nil, err
		}
		requestBodyResrc = mutatedResrc
	}
	// set the mutated resources back to the operation case
	// use `Set...ByResource` to set the actual request params at the same time
	newOperationCase.SetRequestPathParamsByResources(requestPathParamResrc)
	newOperationCase.SetRequestQueryParamsByResources(requestQueryParamResrc)
	newOperationCase.SetRequestBodyByResource(requestBodyResrc)

	return newOperationCase, nil
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
