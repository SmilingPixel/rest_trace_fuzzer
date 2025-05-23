// Code generated by arg_config_generate.py. DO NOT EDIT.
package config

import "flag"
import "os"
import "strconv"
import "github.com/bytedance/sonic"
import "github.com/joho/godotenv"
import "github.com/rs/zerolog/log"

func ParseCmdArgs() {
	flag.StringVar(&GlobalConfig.ConfigFilePath, "config-file", "", "Path to the config file. If a argument is provided in both the config file and command line, the config file argument will be used")
	flag.StringVar(&GlobalConfig.DependencyFilePath, "dependency-file", "", "Path to the dependency file generated by other tools or manually")
	flag.StringVar(&GlobalConfig.DependencyFileType, "dependency-file-type", "", "Type of the dependency file. Currently only support 'Restler'")
	flag.StringVar(&GlobalConfig.ExtraHeaders, "extra-headers", "", "Extra headers to be added to the request, in the format of stringified JSON, e.g., '{\"header1\": \"value1\", \"header2\": \"value2\"}'")
	flag.StringVar(&GlobalConfig.FuzzValueDictFilePath, "fuzz-value-dict-file", "", "Path to the file containing the dictionary of fuzz values, in the format of a JSON list. Each element in the list is a dictionary with two key-value pairs, one is `name` (value is of type string) and the other is `value` (value can be any json).")
	flag.IntVar(&GlobalConfig.FuzzerBudget, "fuzzer-budget", 5, "The maximum time the fuzzer can run, in seconds")
	flag.StringVar(&GlobalConfig.FuzzerType, "fuzzer-type", "Basic", "Type of the fuzzer. Currently only support 'Basic'")
	flag.StringVar(&GlobalConfig.HTTPMiddlewareScriptPath, "http-middleware-script", "", "Script for HTTP middleware handling.")
	flag.StringVar(&GlobalConfig.InternalServiceOpenAPIPath, "internal-service-openapi-spec", "", "Path to internal service openapi spec file, json format")
	flag.StringVar(&GlobalConfig.LogLevel, "log-level", "info", "Log level: debug, info, warn, error, fatal, panic")
	flag.BoolVar(&GlobalConfig.LogToFile, "log-to-file", false, "Should log to file")
	flag.StringVar(&GlobalConfig.OpenAPISpecPath, "openapi-spec", "", "Path to the OpenAPI spec file")
	flag.StringVar(&GlobalConfig.OutputDir, "output-dir", "./output", "Output directory, e.g., ./output")
	flag.StringVar(&GlobalConfig.ServerBaseURL, "server-base-url", "https://www.example.com", "Base URL of the API, e.g., https://www.example.com")
	flag.StringVar(&GlobalConfig.TraceBackendType, "trace-backend-type", "Jaeger", "Type of the trace backend. Currently only support 'Jaeger'")
	flag.StringVar(&GlobalConfig.TraceBackendURL, "trace-backend-url", "", "URL of the trace backend")
	flag.StringVar(&GlobalConfig.TraceIDHeaderKey, "trace-id-header-key", "X-Trace-Id", "The key of the trace ID header to be included in the response. By default, it is 'X-Trace-Id'.")
	flag.Parse()

	// If config file is provided, load the config from the file
	if GlobalConfig.ConfigFilePath != "" {
		configData, err := os.ReadFile(GlobalConfig.ConfigFilePath)
		if err != nil {
			log.Err(err).Msgf("[ParseCmdArgs] Failed to read config file: %s", err)
		}
		err = sonic.Unmarshal(configData, GlobalConfig)
		if err != nil {
			log.Err(err).Msgf("[ParseCmdArgs] Failed to parse config file: %s", err)
		}
	}

	// If environment variables are provided, override the config
	err := godotenv.Load()
	if err != nil {
		log.Err(err).Msgf("[ParseCmdArgs] Failed to load environment variables: %s", err)
	}
	if envVal, ok := os.LookupEnv("CONFIG_FILE_PATH"); ok && envVal != "" {
		GlobalConfig.ConfigFilePath = envVal
	}
	if envVal, ok := os.LookupEnv("DEPENDENCY_FILE_PATH"); ok && envVal != "" {
		GlobalConfig.DependencyFilePath = envVal
	}
	if envVal, ok := os.LookupEnv("DEPENDENCY_FILE_TYPE"); ok && envVal != "" {
		GlobalConfig.DependencyFileType = envVal
	}
	if envVal, ok := os.LookupEnv("EXTRA_HEADERS"); ok && envVal != "" {
		GlobalConfig.ExtraHeaders = envVal
	}
	if envVal, ok := os.LookupEnv("FUZZ_VALUE_DICT_FILE_PATH"); ok && envVal != "" {
		GlobalConfig.FuzzValueDictFilePath = envVal
	}
	if envVal, ok := os.LookupEnv("FUZZER_BUDGET"); ok && envVal != "" {
		envValInt, err := strconv.Atoi(envVal)
		if err != nil {
			log.Err(err).Msgf("[ParseCmdArgs] Failed to parse int: %s", err)
		}
		GlobalConfig.FuzzerBudget = envValInt
	}
	if envVal, ok := os.LookupEnv("FUZZER_TYPE"); ok && envVal != "" {
		GlobalConfig.FuzzerType = envVal
	}
	if envVal, ok := os.LookupEnv("HTTP_MIDDLEWARE_SCRIPT_PATH"); ok && envVal != "" {
		GlobalConfig.HTTPMiddlewareScriptPath = envVal
	}
	if envVal, ok := os.LookupEnv("INTERNAL_SERVICE_OPENAPI_PATH"); ok && envVal != "" {
		GlobalConfig.InternalServiceOpenAPIPath = envVal
	}
	if envVal, ok := os.LookupEnv("LOG_LEVEL"); ok && envVal != "" {
		GlobalConfig.LogLevel = envVal
	}
	if envVal, ok := os.LookupEnv("LOG_TO_FILE"); ok && envVal != "" {
		GlobalConfig.LogToFile = true
	}
	if envVal, ok := os.LookupEnv("OPENAPI_SPEC_PATH"); ok && envVal != "" {
		GlobalConfig.OpenAPISpecPath = envVal
	}
	if envVal, ok := os.LookupEnv("OUTPUT_DIR"); ok && envVal != "" {
		GlobalConfig.OutputDir = envVal
	}
	if envVal, ok := os.LookupEnv("SERVER_BASE_URL"); ok && envVal != "" {
		GlobalConfig.ServerBaseURL = envVal
	}
	if envVal, ok := os.LookupEnv("TRACE_BACKEND_TYPE"); ok && envVal != "" {
		GlobalConfig.TraceBackendType = envVal
	}
	if envVal, ok := os.LookupEnv("TRACE_BACKEND_URL"); ok && envVal != "" {
		GlobalConfig.TraceBackendURL = envVal
	}
	if envVal, ok := os.LookupEnv("TRACE_ID_HEADER_KEY"); ok && envVal != "" {
		GlobalConfig.TraceIDHeaderKey = envVal
	}

	jsonStr, _ := sonic.Marshal(GlobalConfig)
	log.Info().Msgf("[ParseCmdArgs] Parsed arguments: %s", jsonStr)
}
