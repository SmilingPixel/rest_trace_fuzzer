package fuzzer

import "resttracefuzzer/pkg/feedback"

// Fuzzer is the interface that defines the basic methods of a fuzzer.
type Fuzzer interface {
	// Start starts the fuzzer.
	Start() error

	// GetRuntimeGraph gets the runtime graph.
	GetRuntimeGraph() *feedback.RuntimeGraph
}
