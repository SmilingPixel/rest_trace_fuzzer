package report

import (
	"fmt"
	"os"
	"resttracefuzzer/pkg/feedback"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils/http"
	"strconv"

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
	allStatusCodeClassList := http.GetAllStatusCodeClasses()
	statusCodeClass2TotalCnt := make(map[int]int)
	statusCodeClass2Cnt := make(map[int]int)
	for _, statusCodeClass := range allStatusCodeClassList {
		statusCodeClass2TotalCnt[statusCodeClass] = 0
		statusCodeClass2Cnt[statusCodeClass] = 0
	}
	for _, operation := range r.APIManager.APIMap {
		for fieldKey := range operation.Responses.Map() {
			statusCode, err := strconv.Atoi(fieldKey)
			if err != nil { // Ignore the 'default' field.
				continue
			}
			statusCodeClass2TotalCnt[http.GetStatusCodeClass(statusCode)]++
		}
	}

	// Calculate the hit count of the status codes.
	statusHitCount := responseProcesser.StatusHitCount
	for _, statusCount := range statusHitCount {
		for statusCode, count := range statusCount {
			if count > 0 {
				statusCodeClass2Cnt[http.GetStatusCodeClass(statusCode)]++
			}
		}
	}

	// Calculate the coverage of the status codes.
	systemTestReport.StatusCoverage = make(map[int]float64)
	for statusCodeClass, totalCnt := range statusCodeClass2TotalCnt {
		if totalCnt == 0 {
			continue
		}
		systemTestReport.StatusCoverage[statusCodeClass] = float64(statusCodeClass2Cnt[statusCodeClass]) / float64(totalCnt)
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
