package casemanager

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// CaseManager manages the test cases.
type CaseManager struct {
	// Testcases is a list of test cases.
	Testcases []*Testcase
}


// NewCaseManager creates a new CaseManager.
func NewCaseManager() *CaseManager {
	testcases := make([]*Testcase, 0)
	return &CaseManager{
		Testcases: testcases,
	}
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
