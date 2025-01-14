package report

import (
	"os"
	"resttracefuzzer/pkg/feedback"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

// InternalServiceReporter analyses and reports the results of internal service (API) states.
type InternalServiceReporter struct {
}

// NewInternalServiceReporter creates a new InternalServiceReporter.
func NewInternalServiceReporter() *InternalServiceReporter {
	return &InternalServiceReporter{}
}

// GenerateInternalServiceReport generates the internal service report.
func (r *InternalServiceReporter) GenerateInternalServiceReport(runtimeGraph *feedback.RuntimeGraph, outputPath string) error {
	// At present, we only report the edge coverage.
	coveredEdges := 0
	for _, edge := range runtimeGraph.Edges {
		if edge.HitCount > 0 {
			coveredEdges++
		}
	}
	// Calculate the coverage of the edges.
	edgeCoverage := float64(coveredEdges) / float64(len(runtimeGraph.Edges))
	
	// Generate the report and marshal it to JSON.
	report := InternalServiceTestReport{
		EdgeCoverage: edgeCoverage,
	}
	reportJSON, err := sonic.Marshal(report)
	if err != nil {
		log.Error().Msgf("[InternalServiceReporter.GenerateInternalServiceReport] Failed to marshal the internal service report: %v", err)
		return err
	}

	// Write the report to a file.
	err = os.WriteFile(outputPath, reportJSON, 0644)
	if err != nil {
		log.Error().Msgf("[InternalServiceReporter.GenerateInternalServiceReport] Failed to write the internal service report to file: %v", err)
		return err
	}
	return nil
}


