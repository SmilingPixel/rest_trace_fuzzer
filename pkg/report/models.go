package report

import "resttracefuzzer/pkg/static"

// SystemTestReport is the report of the system-level test.
type SystemTestReport struct {

	StatusCoverage float64 `json:"statusCoverage"`

	// StatusHitCount is the hit count of the status code.
	StatusHitCount map[static.SimpleAPIMethod]map[int]int `json:"statusHitCount"`

}

