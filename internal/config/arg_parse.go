package config

import (
	"flag"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

func ParseCmdArgs() {
	// Required arguments
	flag.StringVar(&GlobalConfig.OpenAPISpecPath, "openapi-spec", "", "Path to the OpenAPI spec file")
	flag.StringVar(&GlobalConfig.ServerBaseURL, "server-base-url", "https://www.example.com", "Base URL of the API, e.g., https://www.example.com")
	// Internal service openapi spec, multiple files, json format
	var intenalServiceOpenAPISpecsJSON string
	flag.StringVar(&intenalServiceOpenAPISpecsJSON, "internal-service-openapi-spec", "", "JSON string of service name and path to the OpenAPI spec files in the format {\"service_name\":\"oas_spec_file\"}")


	// Deprecated, used to parse dependency files from Restler
	flag.StringVar(&GlobalConfig.DependencyFilePath, "dependency-file", "", "Path to the dependency file generated by other tools or manually")
	flag.StringVar(&GlobalConfig.DependencyFileType, "dependency-file-type", "", "Type of the dependency file. Currently only support 'Restler'")

	// Optional arguments
	flag.StringVar(&GlobalConfig.FuzzerType, "fuzzer-type", "Basic", "Type of the fuzzer. Currently only support 'Basic'")
	// flag.StringVar(&GlobalConfig.DependencyFileType, "dependency-file-type", "", "Type of the dependency file. Currently only support 'Restler'")
	flag.DurationVar(&GlobalConfig.FuzzerBudget, "fuzzer-budget", 30, "The maximum time the fuzzer can run, in seconds")
	flag.StringVar(&GlobalConfig.LogLevel, "log-level", "info", "Log level: debug, info, warn, error, fatal, panic")

	
	flag.Parse()

	// Parse the JSON string into the map
    if intenalServiceOpenAPISpecsJSON != "" {
		log.Info().Msgf("[ParseCmdArgs] Internal service openapi spec provided: %s", intenalServiceOpenAPISpecsJSON)
		intenalServiceOpenAPISpecsJSON = strings.ReplaceAll(intenalServiceOpenAPISpecsJSON, "\n", "") // Remove newlines, otherwise it will cause unmarshal error
		intenalServiceOpenAPISpecsJSON = strings.ReplaceAll(intenalServiceOpenAPISpecsJSON, "\t", "") // Remove newlines, otherwise it will cause unmarshal error
		log.Info().Msgf("[ParseCmdArgs] Internal service openapi spec provided: %s", intenalServiceOpenAPISpecsJSON)
        err := sonic.UnmarshalString(intenalServiceOpenAPISpecsJSON, &GlobalConfig.InternalServiceOpenAPISpecs)
        if err != nil {
            log.Error().Err(err).Msg("[ParseCmdArgs] Failed to parse openapi-specs-json")
        }
    } else {
		log.Warn().Msg("[ParseCmdArgs] No internal service openapi spec provided")
		GlobalConfig.InternalServiceOpenAPISpecs = make(map[string]string)
	}

	jsonStr, _ := sonic.Marshal(GlobalConfig)
	log.Info().Msgf("Parsed command line arguments: %s", jsonStr)


}
	
