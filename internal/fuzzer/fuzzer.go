package fuzzer

import fuzzruntime "resttracefuzzer/pkg/runtime"

// Fuzzer is the interface that defines the basic methods of a fuzzer.
type Fuzzer interface {
	// Start starts the fuzzer.
	Start() error

	// GetCallInfoGraph gets the runtime call info graph.
	GetCallInfoGraph() *fuzzruntime.CallInfoGraph
}
