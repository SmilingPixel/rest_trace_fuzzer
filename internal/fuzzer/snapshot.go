package fuzzer

// FuzzingSnapshot represents a snapshot of the fuzzing process.
// It includes metrics such as runtime call info graph edge coverage and the count of covered status codes.
// TODO: Add more metrics. @xunzhou24
type FuzzingSnapshot struct {
	// CallInfoGraphEdgeCoveredCount is the number of edges covered in the runtime call info graph.
	CallInfoGraphEdgeCoveredCount int `json:"callInfoGraphEdgeCoveredCount"`

	// CoveredStatusCodeCount is the number of unique status codes covered during fuzzing.
	CoveredStatusCodeCount int `json:"coveredStatusCodeCount"`
}

// NewFuzzingSnapshot creates a new FuzzingSnapshot.
func NewFuzzingSnapshot() *FuzzingSnapshot {
	return &FuzzingSnapshot{
		CallInfoGraphEdgeCoveredCount: 0,
		CoveredStatusCodeCount:   0,
	}
}

// Update updates the snapshot with the edge coverage and the count of covered status codes.
// It returns whether the update is successful and a higher coverage is achieved.
func (s *FuzzingSnapshot) Update(edgeCoveredCount int, statusCodeCount int) bool {
	ret := false
	if edgeCoveredCount > s.CallInfoGraphEdgeCoveredCount {
		ret = true
		s.CallInfoGraphEdgeCoveredCount = edgeCoveredCount
	}
	if statusCodeCount > s.CoveredStatusCodeCount {
		ret = true
		s.CoveredStatusCodeCount = statusCodeCount
	}
	return ret
}
