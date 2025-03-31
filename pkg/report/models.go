package report

import (
	"resttracefuzzer/pkg/casemanager"
	"resttracefuzzer/pkg/feedback"
	"resttracefuzzer/pkg/resource"
	"resttracefuzzer/pkg/static"
	"time"

	"github.com/google/uuid"
)

type StatusHitCountReport struct {
	APIMethod static.SimpleAPIMethod `json:"APIMethod"`
	Status    int                    `json:"status"`
	HitCount  int                    `json:"hitCount"`
}

// SystemTestReport is the report of the system-level test.
type SystemTestReport struct {

	// StatusCoverage is the coverage of the status code.
	// It maps from status code class (2xx, 3xx, 4xx, 5xx) to coverage.
	StatusCoverage map[int]float64 `json:"statusCoverage"`

	// statusHitCount is the hit count of the status code.
	// Ignore this field when marshalling to JSON.
	// This field has same information as statusHitCountReport.
	// You should set it using SetStatusHitCountReport.
	StatusHitCount map[static.SimpleAPIMethod]map[int]int `json:"-"`

	// statusHitCountReport is the report of the status hit count.
	// This field is used when marshalling to JSON.
	// It is generated from statusHitCount.
	// You should set statusHitCount using SetStatusHitCountReport.
	StatusHitCountReport []StatusHitCountReport `json:"statusHitCountReport"`
}

// SetStatusHitCountReport sets the status hit count report.
func (r *SystemTestReport) SetStatusHitCountReport(statusHitCount map[static.SimpleAPIMethod]map[int]int) {
	r.StatusHitCount = statusHitCount
	r.StatusHitCountReport = make([]StatusHitCountReport, 0)
	for APIMethod, statusCount := range statusHitCount {
		for status, hitCount := range statusCount {
			r.StatusHitCountReport = append(r.StatusHitCountReport, StatusHitCountReport{
				APIMethod: APIMethod,
				Status:    status,
				HitCount:  hitCount,
			})
		}
	}
}

// InternalServiceTestReport is the report of the internal service test.
type InternalServiceTestReport struct {

	// EdgeCoverage is the coverage of the edge.
	EdgeCoverage float64 `json:"edgeCoverage"`

	// FinalRuntimeGraph is the final runtime graph.
	FinalRuntimeGraph *feedback.RuntimeGraph `json:"finalRuntimeGraph"`
}

// FuzzerStateReport is the report of the fuzzer state.
type FuzzerStateReport struct {

	// ResourcePool is the resource pool.
	ResourceNameMap map[string][]resource.Resource `json:"resourceNameMap"`

}

// OperationCaseForReport stores info of an operation tested during fuzzing.
// Simplified version of [resttracefuzzer/pkg/casemanager.OperationCase]
type OperationCaseForReport struct {
	// APIMethod is the API method.
	APIMethod          static.SimpleAPIMethod `json:"APIMethod"`

	// RequestHeaders contains the headers to be sent with the request.
	RequestHeaders map[string]string `json:"requestHeaders"`

	// RequestPathParams contains the path parameters to be sent with the request.
	RequestPathParams map[string]string `json:"requestPathParams"`

	// RequestQueryParams contains the query parameters to be sent with the request.
	RequestQueryParams map[string]string `json:"requestQueryParams"`

	// RequestBody contains the body to be sent with the request.
	// It is a json object as a string.
	RequestBody string `json:"requestBody"`

	// ResponseStatusCode is the expected status code of the response.
	ResponseStatusCode int `json:"responseStatusCode"`
}

// NewReportFromOperationCase creates a new OperationCaseForReport from an OperationCase.
func NewReportFromOperationCase(operationCase *casemanager.OperationCase) *OperationCaseForReport {
	return &OperationCaseForReport{
		APIMethod:          operationCase.APIMethod,
		RequestHeaders:     operationCase.RequestHeaders,
		RequestPathParams:  operationCase.RequestPathParams,
		RequestQueryParams: operationCase.RequestQueryParams,
		RequestBody:        string(operationCase.RequestBody),
		ResponseStatusCode: operationCase.ResponseStatusCode,
	}
}

// TestScenarioForReport stores info of a test scenario tested during fuzzing.
// Simplified version of [resttracefuzzer/pkg/casemanager.TestScenario]
type TestScenarioForReport struct {

	// OperationCases is a sequence of tested operation cases.
	OperationCases []*OperationCaseForReport `json:"operationCases"`

	// OperationCaseLength is the length of the operation cases.
	// It is used to improve the readability of the report.
	OperationCaseLength int `json:"operationCaseLength"`
	
	// EndTime is the end time of the test scenario.
	EndTime time.Time `json:"endTime"`

	// TestScenarioUUID is the UUID of the test scenario.
	TestScenarioUUID uuid.UUID `json:"testScenarioUUID"`
}

// NewReportFromTestScenario creates a new TestScenarioForReport from a TestScenario.
func NewReportFromTestScenario(testScenario *casemanager.TestScenario) *TestScenarioForReport {
	operationCases := make([]*OperationCaseForReport, 0)
	for _, operationCase := range testScenario.OperationCases {
		operationCases = append(operationCases, NewReportFromOperationCase(operationCase))
	}
	return &TestScenarioForReport{
		OperationCases: operationCases,
		OperationCaseLength: len(operationCases),
		EndTime:          time.Now(),
		TestScenarioUUID: testScenario.UUID,
	}
}


// TestLogReport is the report of the test log.
// It contains the history of testing.
// To reduce size of the report, it uses a simplified version of the tested scenario.
type TestLogReport struct {
	TestedScenarios []*TestScenarioForReport `json:"testedScenarios"`
}

// NewTestLogReport creates a new TestLogReport.
func NewTestLogReport() *TestLogReport {
	return &TestLogReport{
		TestedScenarios: make([]*TestScenarioForReport, 0),
	}
}
