package casemanager

import (
	"fmt"
	"resttracefuzzer/pkg/resource"
	"resttracefuzzer/pkg/static"

	"github.com/rs/zerolog/log"
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
		TestScenarios:       testScenarios,
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
	// for _, operationCase := range testScenario.OperationCases {
	// 	// fill the request params
	// 	requestParams := make(map[string]string)
	// }
	return testScenario, nil
}

// Push adds a test case to the case manager.
func (m *CaseManager) Push(testcase *TestScenario) {
	m.TestScenarios = append(m.TestScenarios, testcase)
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
		testcase := &TestScenario{
			OperationCases: []*OperationCase{&operationCase},
		}
		m.Push(testcase)
	}
	return nil
}
