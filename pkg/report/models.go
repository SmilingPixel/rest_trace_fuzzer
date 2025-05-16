package report

import (
	"fmt"
	"resttracefuzzer/pkg/casemanager"
	"resttracefuzzer/pkg/resource"
	fuzzruntime "resttracefuzzer/pkg/runtime"
	"resttracefuzzer/pkg/static"
	"time"

	"github.com/google/uuid"
)

type APIMethodStatusHitCountReport struct {
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
	APIMethodStatusHitCount map[static.SimpleAPIMethod]map[int]int `json:"-"`

	// statusHitCountReport is the report of the status hit count.
	// This field is used when marshalling to JSON.
	// It is generated from statusHitCount.
	// You should set statusHitCount using SetStatusHitCountReport.
	APIMethodStatusHitCountReport []APIMethodStatusHitCountReport `json:"statusHitCountReport"`
}

// SetStatusHitCountReport sets the status hit count report.
func (r *SystemTestReport) SetStatusHitCountReport(statusHitCount map[static.SimpleAPIMethod]map[int]int) {
	r.APIMethodStatusHitCount = statusHitCount
	r.APIMethodStatusHitCountReport = make([]APIMethodStatusHitCountReport, 0)
	for APIMethod, statusCount := range statusHitCount {
		for status, hitCount := range statusCount {
			r.APIMethodStatusHitCountReport = append(r.APIMethodStatusHitCountReport, APIMethodStatusHitCountReport{
				APIMethod: APIMethod,
				Status:    status,
				HitCount:  hitCount,
			})
		}
	}
}

// InternalServiceTestReport is the report of states of the internal service after fuzzing.
type InternalServiceTestReport struct {

	// EdgeCoverage is the coverage of the edge.
	EdgeCoverage float64 `json:"edgeCoverage"`

	// FinalCallInfoGraph is the final runtime call info graph.
	FinalCallInfoGraph *fuzzruntime.CallInfoGraph `json:"finalCallInfoGraph"`

	// RuntimeHighConfidenceReachabilityMap is the runtime reachability map.
	// By default, it only includes high confidence reachability map.
	RuntimeHighConfidenceReachabilityMap *ReachabilityMapForReport `json:"runtimeHighConfidenceReachabilityMap"`
}

// FuzzerStateReport is the report of the fuzzer state.
type FuzzerStateReport struct {

	// ResourceNameMap is the map of resource name to resource.
	// It is not jsonified, as we would call its custom method to jsonified it.
	// ResourceNameMapJsonObject is the jsonified (for resources) version of ResourceNameMap, and would be set when ResourceNameMap is set.
	ResourceNameMap map[string][]resource.Resource `json:"-"`

	// ResourceJSONObjectNameMap is the jsonified version of ResourceNameMap.
	ResourceJSONObjectNameMap map[string][]interface{} `json:"resourceNameMap"`
}

// OperationCaseForReport stores info of an operation tested during fuzzing.
// Simplified version of [resttracefuzzer/pkg/casemanager.OperationCase]
type OperationCaseForReport struct {
	// APIMethod is the API method.
	APIMethod static.SimpleAPIMethod `json:"APIMethod"`

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
		OperationCases:      operationCases,
		OperationCaseLength: len(operationCases),
		EndTime:             time.Now(),
		TestScenarioUUID:    testScenario.UUID,
	}
}

// TestLogReport is the report of the test log.
// It contains the history of testing, and other information as well.
// To reduce size of the report, it uses a simplified version of the tested scenario.
type TestLogReport struct {
	// TestedScenarios is the list of tested scenarios.
	TestedScenarios []*TestScenarioForReport `json:"testedScenarios"`

	// TestedScenariosLengthCount records the number of tested scenarios of each length.
	// It maps from length of the tested scenarios to the number of tested scenarios.
	TestedScenariosLengthCount map[int]int `json:"testedScenariosLengthCount"`
}

// NewTestLogReport creates a new TestLogReport.
func NewTestLogReport() *TestLogReport {
	return &TestLogReport{
		TestedScenarios: make([]*TestScenarioForReport, 0),
		TestedScenariosLengthCount: make(map[int]int),
	}
}

// ReachabilityMapForReport is the report of the reachability map.
// It contains the reachability information of the system.
// It is a simplified version of [resttracefuzzer/pkg/static.ReachabilityMap].
type ReachabilityMapForReport struct {
	// M maps from external API to internal APIs.
	// external API is string representation of [resttracefuzzer/pkg/static.SimpleAPIMethod], as json key.
	M map[string][]static.InternalServiceEndpoint `json:"m"`
}

// NewReachabilityMapForReport creates a new ReachabilityMapForReport from a ReachabilityMap.
func NewReachabilityMapForReport(reachabilityMap *static.ReachabilityMap) *ReachabilityMapForReport {
	reachabilityMapForReport := &ReachabilityMapForReport{
		M: make(map[string][]static.InternalServiceEndpoint),
	}
	for external, internals := range reachabilityMap.External2Internal {
		reachabilityMapForReport.M[fmt.Sprintf("%v", external)] = internals
	}
	return reachabilityMapForReport
}
