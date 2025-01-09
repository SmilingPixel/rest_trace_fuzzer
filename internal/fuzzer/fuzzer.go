package fuzzer

import (

)


// Fuzzer is the interface that defines the basic methods of a fuzzer.
type Fuzzer interface {
	// Start starts the fuzzer.
	Start() error
}


