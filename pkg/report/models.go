package report

import (
	"resttracefuzzer/pkg/feedback"
	"resttracefuzzer/pkg/static"
)

type StatusHitCountReport struct {
	APIMethod static.SimpleAPIMethod `json:"APIMethod"`
	Status    int                    `json:"status"`
	HitCount  int                    `json:"hitCount"`
}

// SystemTestReport is the report of the system-level test.
type SystemTestReport struct {

	// StatusCoverage is the coverage of the status code.
	StatusCoverage float64 `json:"statusCoverage"`

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
