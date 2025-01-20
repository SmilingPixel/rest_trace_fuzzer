package report

import (
	"fmt"
	"math"
	"os"
	"resttracefuzzer/pkg/feedback"
	"resttracefuzzer/pkg/static"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

// SystemReporter analyses and reports the results of system-level fuzzing.
// It supports the following features:
// 1. Report the coverage of the Endpoints, i.e., number of (path, method) pairs that have been visited.
// 2. TODO: to implement the rest of the features. @xunzhou24
type SystemReporter struct {
	APIManager *static.APIManager
}

// NewSystemReporter creates a new SystemReporter.
func NewSystemReporter(apiManager *static.APIManager) *SystemReporter {
	return &SystemReporter{
		APIManager: apiManager,
	}
}

// GenerateSystemReport generates the system-level report.
// The report includes the coverage of the Endpoints.
func (r *SystemReporter) GenerateSystemReport(responseChecker *feedback.ResponseChecker, outputPath string) error {
	if responseChecker == nil {
		log.Error().Msg("[SystemReporter.GenerateSystemReport] The response checker is nil.")
		return fmt.Errorf("responseChecker is nil")
	}

	systemTestReport := SystemTestReport{}

	// TODO: Find reponse status codes that are not defined in the OpenAPI document. @xunzhou24

	// Calculate the total number of status codes in the OpenAPI document.
	totalStatusCount := 0
	for _, operation := range r.APIManager.APIMap {
		totalStatusCount += len(operation.Responses.Map())
	}

	// Calculate the hit count of the status codes.
	statusHitCount := responseChecker.StatusHitCount
	totalHitStatusCount := 0
	for _, statusCount := range statusHitCount {
		for _, count := range statusCount {
			if count > 0 {
				totalHitStatusCount++
			}
		}
	}

	// Calculate the coverage of the status codes.
	systemTestReport.StatusCoverage = float64(totalHitStatusCount) / float64(totalStatusCount)
	if math.IsInf(systemTestReport.StatusCoverage, 0) || math.IsNaN(systemTestReport.StatusCoverage) {
		return fmt.Errorf("invalid status coverage: %f", systemTestReport.StatusCoverage)
	}
	systemTestReport.SetStatusHitCountReport(statusHitCount)

	// marshal the report to a JSON file.
	reportBytes, err := sonic.Marshal(systemTestReport)
	if err != nil {
		log.Error().Err(err).Msgf("[SystemReporter.GenerateSystemReport] Failed to marshal the system test report: %v", err)
	}

	// Write the JSON string to the output file.
	err = os.WriteFile(outputPath, reportBytes, 0644)
	if err != nil {
		log.Error().Err(err).Msgf("[SystemReporter.GenerateSystemReport] Failed to write the system test report to file: %v", err)
		return err
	}
	log.Info().Msgf("[SystemReporter.GenerateSystemReport] System test report has been written to %s", outputPath)
	return nil
}
