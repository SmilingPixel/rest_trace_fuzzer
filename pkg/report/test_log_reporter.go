package report

import (
	"os"
	"resttracefuzzer/pkg/casemanager"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

// TestLogReporter is responsible for logging the tested operations (with their results),
// and generating a report after the fuzzing process.
type TestLogReporter struct {
	TestLogReport *TestLogReport
}

// NewTestLogReporter creates a new TestLogReporter.
func NewTestLogReporter() *TestLogReporter {
	return &TestLogReporter{
		TestLogReport: NewTestLogReport(),
	}
}

// LogTestScenario logs the tested test scenario.
// To reduce the size of the report, it removes some info (such as response body) from origin tested operation, and uses a simplified version of the tested scenario in the report.
func (r *TestLogReporter) LogTestScenario(testScenario *casemanager.TestScenario) {
	r.TestLogReport.TestedScenarios = append(r.TestLogReport.TestedScenarios, NewReportFromTestScenario(testScenario))
	r.TestLogReport.TestedScenariosLengthCount[len(testScenario.OperationCases)]++
}

// GenerateTestLogReport generates the test log report.
func (r *TestLogReporter) GenerateTestLogReport(outputPath string) error {
	// marshal the report to a JSON file.
	reportBytes, err := sonic.Marshal(r.TestLogReport)
	if err != nil {
		log.Err(err).Msgf("[TestLogReporter.GenerateTestLogReport] Failed to marshal the test log report")
	}

	// Write the JSON string to the output file.
	err = os.WriteFile(outputPath, reportBytes, 0644)
	if err != nil {
		log.Err(err).Msgf("[TestLogReporter.GenerateTestLogReport] Failed to write the test log report")
		return err
	}
	log.Info().Msgf("[TestLogReporter.GenerateTestLogReport] Test log report has been written to %s", outputPath)
	return nil
}
