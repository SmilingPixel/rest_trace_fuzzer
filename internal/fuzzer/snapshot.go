package fuzzer

// FuzzingSnapshot represents a snapshot of the fuzzing process.
// It includes metrics such as runtime call info graph edge coverage and the count of covered status codes.
// TODO: Add more metrics. @xunzhou24
type FuzzingSnapshot struct {
	// CallInfoGraphEdgeCoverage is the percentage of edges covered in the runtime call info graph.
	CallInfoGraphEdgeCoverage float64 `json:"callInfoGraphEdgeCoverage"`

	// CoveredStatusCodeCount is the number of unique status codes covered during fuzzing.
	CoveredStatusCodeCount int `json:"coveredStatusCodeCount"`
}

// NewFuzzingSnapshot creates a new FuzzingSnapshot.
func NewFuzzingSnapshot() *FuzzingSnapshot {
	return &FuzzingSnapshot{
		CallInfoGraphEdgeCoverage: 0.0,
		CoveredStatusCodeCount:   0,
	}
}

// Update updates the snapshot with the edge coverage and the count of covered status codes.
// It returns whether the update is successful and a higher coverage is achieved.
func (s *FuzzingSnapshot) Update(edgeCoverage float64, statusCodeCount int) bool {
	ret := false
	if edgeCoverage > s.CallInfoGraphEdgeCoverage {
		ret = true
		s.CallInfoGraphEdgeCoverage = edgeCoverage
	}
	if statusCodeCount > s.CoveredStatusCodeCount {
		ret = true
		s.CoveredStatusCodeCount = statusCodeCount
	}
	return ret
}
