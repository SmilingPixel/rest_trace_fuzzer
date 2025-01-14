package report

import "resttracefuzzer/pkg/static"

// SystemTestReport is the report of the system-level test.
type SystemTestReport struct {

	// StatusCoverage is the coverage of the status code.
	StatusCoverage float64 `json:"statusCoverage"`

	// StatusHitCount is the hit count of the status code.
	StatusHitCount map[static.SimpleAPIMethod]map[int]int `json:"statusHitCount"`

}

// InternalServiceTestReport is the report of the internal service test.
type InternalServiceTestReport struct {

	// EdgeCoverage is the coverage of the edge.
	EdgeCoverage float64 `json:"edgeCoverage"`

}

