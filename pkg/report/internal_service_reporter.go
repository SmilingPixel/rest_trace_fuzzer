package report

import (
	"os"
	fuzzruntime "resttracefuzzer/pkg/runtime"
	"resttracefuzzer/pkg/static"
	"slices"

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
// The report includes the edge coverage.
func (r *InternalServiceReporter) GenerateInternalServiceReport(
	callInfoGraph *fuzzruntime.CallInfoGraph,
	runtimeReachabilityMap *fuzzruntime.RuntimeReachabilityMap,
	outputPath string,
) error {
	// At present, we only report the edge coverage.
	coveredEdges := 0
	for _, edge := range callInfoGraph.Edges {
		if edge.HitCount > 0 {
			coveredEdges++
		}
	}
	// Calculate the coverage of the edges.
	edgeCoverage := float64(coveredEdges) / float64(len(callInfoGraph.Edges))

	slices.SortFunc(callInfoGraph.Edges, func(a, b *fuzzruntime.CallInfoEdge) int {
		return static.CompareInternalServiceEndpoint(a.Source, b.Source)
	})

	// Generate the report and marshal it to JSON.
	report := InternalServiceTestReport{
		EdgeCoverage:       edgeCoverage,
		RuntimeHighConfidenceReachabilityMap: NewReachabilityMapForReport(runtimeReachabilityMap.HighConfidenceMap),
		FinalCallInfoGraph: callInfoGraph,
	}
	reportJSON, err := sonic.Marshal(report)
	if err != nil {
		log.Err(err).Msgf("[InternalServiceReporter.GenerateInternalServiceReport] Failed to marshal the internal service report")
		return err
	}

	// Write the report to a file.
	err = os.WriteFile(outputPath, reportJSON, 0644)
	if err != nil {
		log.Err(err).Msgf("[InternalServiceReporter.GenerateInternalServiceReport] Failed to write the internal service report to file")
		return err
	}
	return nil
}
