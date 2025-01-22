package casemanager

import (
	"fmt"
	"resttracefuzzer/pkg/static"

	"github.com/rs/zerolog/log"
)

// CaseManager manages the test cases.
type CaseManager struct {
	// Testcases is a list of test cases.
	Testcases []*Testcase

	// The API manager.
	APIManager *static.APIManager
}

// NewCaseManager creates a new CaseManager.
func NewCaseManager(APIManager *static.APIManager) *CaseManager {
	testcases := make([]*Testcase, 0)
	m := &CaseManager{
		APIManager: APIManager,
		Testcases:  testcases,
	}
	m.initTestcasesFromDoc()
	return m
}

// Pop pops a test case of highest priority from the case manager.
func (m *CaseManager) Pop() (*Testcase, error) {
	// TODO: Implement this method. @xunzhou24
	// We select the first test case for now.
	if len(m.Testcases) == 0 {
		log.Error().Msg("[CaseManager.Pop] No test case available")
		return nil, fmt.Errorf("no test case available")
	}
	testcase := m.Testcases[0]
	m.Testcases = m.Testcases[1:]
	return testcase, nil
}

// Push adds a test case to the case manager.
func (m *CaseManager) Push(testcase *Testcase) {
	m.Testcases = append(m.Testcases, testcase)
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
		testcase := &Testcase{
			OperationCases: []*OperationCase{&operationCase},
		}
		m.Push(testcase)
	}
	return nil
}
