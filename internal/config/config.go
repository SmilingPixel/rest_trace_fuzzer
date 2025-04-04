// Code generated by arg_config_generate.py. DO NOT EDIT.
package config

var GlobalConfig *RuntimeConfig

type RuntimeConfig struct {
	// Path to the config file. If a argument is provided in both the config file and command line, the config file argument will be used
	ConfigFilePath string `json:"configFilePath"`

	// Path to the dependency file generated by other tools or manually
	DependencyFilePath string `json:"dependencyFilePath"`

	// Type of the dependency file. Currently only support 'Restler'. Required if dependency-file is provided.
	DependencyFileType string `json:"dependencyFileType"`

	// Enable energy (priority) of test operation. If true, energy would affect the test operation selection when extending the test scenario (sequence of test operations).
	EnableEnergyOperation bool `json:"enableEnergyOperation"`

	// Enable energy (priority) of test scenario. If true, energy would affect the test scenario selection when starting a new test loop
	EnableEnergyScenario bool `json:"enableEnergyScenario"`

	// Extra headers to be added to the request, in the format of stringified JSON, e.g., '{\"header1\": \"value1\", \"header2\": \"value2\"}'
	ExtraHeaders string `json:"extraHeaders"`

	// Path to the file containing the dictionary of fuzz values, in the format of a JSON list. Each element in the list is a dictionary with two key-value pairs, one is `name` (value is of type string) and the other is `value` (value can be any json).
	FuzzValueDictFilePath string `json:"fuzzValueDictFilePath"`

	// The maximum time the fuzzer can run, in seconds
	FuzzerBudget int `json:"fuzzerBudget"`

	// Type of the fuzzer. Currently only support 'Basic'
	FuzzerType string `json:"fuzzerType"`

	// Path to the script file that contains the HTTP middleware functions, see [HTTP Middleware Script](#about-http-middleware-script).
	HTTPMiddlewareScriptPath string `json:"HTTPMiddlewareScriptPath"`

	// Path to internal service openapi spec file, json format
	InternalServiceOpenAPIPath string `json:"internalServiceOpenAPIPath"`

	// Log level: debug, info, warn, error, fatal, panic
	LogLevel string `json:"logLevel"`

	// Should log to file
	LogToFile bool `json:"logToFile"`

	// Maximum number of operations to execute in each scenario
	MaxOpsPerScenario int `json:"maxOpsPerScenario"`

	// The maximum executed times of a test operation case.
	MaxAllowedOperationCaseExecutedCount int `json:"maxAllowedOperationCaseExecutedCount"`

	// The maximum number of test operation cases in the queue of an API method.
	MaxAllowedOperationCases int `json:"maxAllowedOperationCases"`

	// The maximum executed times of a test scenario.
	MaxAllowedScenarioExecutedCount int `json:"maxAllowedScenarioExecutedCount"`

	// The maximum number of test scenarios in the queue.
	MaxAllowedScenarios int `json:"maxAllowedScenarios"`

	// Path to the OpenAPI spec file
	OpenAPISpecPath string `json:"OpenAPISpecPath"`

	// Output directory, e.g., ./output
	OutputDir string `json:"outputDir"`

	// Base URL of the API, e.g., https://www.example.com
	ServerBaseURL string `json:"serverBaseURL"`

	// Type of the trace backend. Currently only support 'Jaeger'
	TraceBackendType string `json:"traceBackendType"`

	// URL of the trace backend
	TraceBackendURL string `json:"traceBackendURL"`

	// The key of the trace ID header to be included in the response. By default, it is 'X-Trace-Id'.
	TraceIDHeaderKey string `json:"traceIDHeaderKey"`
}

func InitConfig() {
	GlobalConfig = &RuntimeConfig{}
}
