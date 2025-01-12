package config

import "time"

var GlobalConfig *RuntimeConfig

type RuntimeConfig struct {
	// Path to the OpenAPI spec file
	OpenAPISpecPath string

	// Internal service openapi spec, multiple files, map from service name to file path
	InternalServiceOpenAPISpecs map[string]string

	// Path to the dependency file generated by other tools or manually
	DependencyFilePath string

	// Type of the dependency file
	// Currently only support 'Restler'
	DependencyFileType string

	// Type of the fuzzer
	FuzzerType string

	// Budget is the budget of the fuzzer, which is the maximum time the fuzzer can run, in milliseconds.
	FuzzerBudget time.Duration

	// BaseURL is the base URL of the API.
	ServerBaseURL string

	// log level
	LogLevel string
}

func InitConfig() {
	GlobalConfig = &RuntimeConfig{}
}
