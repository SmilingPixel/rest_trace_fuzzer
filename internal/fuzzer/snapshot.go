package fuzzer

// FuzzingSnapshot represents a snapshot of the fuzzing process.
// It includes metrics such as runtime graph edge coverage and the count of covered status codes.
// TODO: Add more metrics. @xunzhou24
type FuzzingSnapshot struct {
    // RuntimeGraphEdgeCoverage is the percentage of edges covered in the runtime graph.
    RuntimeGraphEdgeCoverage float64 `json:"runtime_graph_edge_coverage"`

    // CoveredStatusCodeCount is the number of unique status codes covered during fuzzing.
    CoveredStatusCodeCount int `json:"covered_status_code_count"`
}

// NewFuzzingSnapshot creates a new FuzzingSnapshot.
func NewFuzzingSnapshot() *FuzzingSnapshot {
	return &FuzzingSnapshot{
		RuntimeGraphEdgeCoverage: 0.0,
		CoveredStatusCodeCount:   0,
	}
}

// Update updates the snapshot with the edge coverage and the count of covered status codes.
// It returns whether the update is successful and a higher coverage is achieved.
func (s *FuzzingSnapshot) Update(edgeCoverage float64, statusCodeCount int) bool {
	ret := false
	if edgeCoverage > s.RuntimeGraphEdgeCoverage {
		ret = true
		s.RuntimeGraphEdgeCoverage = edgeCoverage
	}
	if statusCodeCount > s.CoveredStatusCodeCount {
		ret = true
		s.CoveredStatusCodeCount = statusCodeCount
	}
	return ret
}
