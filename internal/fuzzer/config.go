package fuzzer

import "time"

// FuzzerConfig is the configuration of a fuzzer.
type FuzzerConfig struct {

	// Budget is the budget of the fuzzer, which is the maximum time the fuzzer can run, in milliseconds.
	Budget time.Duration

	// BaseURL is the base URL of the API.
	BaseURL string
}
