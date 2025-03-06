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
func NewSystemReporter(APIManager *static.APIManager) *SystemReporter {
	return &SystemReporter{
		APIManager: APIManager,
	}
}

// GenerateSystemReport generates the system-level report.
// The report includes the coverage of the Endpoints and Status Codes.
func (r *SystemReporter) GenerateSystemReport(responseProcesser *feedback.ResponseProcesser, outputPath string) error {
	if responseProcesser == nil {
		log.Error().Msg("[SystemReporter.GenerateSystemReport] responseProcesser is nil.")
		return fmt.Errorf("responseProcesser is nil")
	}

	systemTestReport := SystemTestReport{}

	// TODO: Find reponse status codes that are not defined in the OpenAPI document. @xunzhou24

	// Calculate the total number of status codes in the OpenAPI document.
	totalStatusCount := 0
	for _, operation := range r.APIManager.APIMap {
		totalStatusCount += len(operation.Responses.Map())
	}

	// Calculate the hit count of the status codes.
	statusHitCount := responseProcesser.StatusHitCount
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
		log.Err(err).Msgf("[SystemReporter.GenerateSystemReport] Failed to marshal the system test report")
	}

	// Write the JSON string to the output file.
	err = os.WriteFile(outputPath, reportBytes, 0644)
	if err != nil {
		log.Err(err).Msgf("[SystemReporter.GenerateSystemReport] Failed to write the system test report to file")
		return err
	}
	log.Info().Msgf("[SystemReporter.GenerateSystemReport] System test report has been written to %s", outputPath)
	return nil
}
