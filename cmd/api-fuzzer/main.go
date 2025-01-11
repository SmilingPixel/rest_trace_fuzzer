package main

import (
	"resttracefuzzer/internal/config"
	"resttracefuzzer/internal/fuzzer"
	"resttracefuzzer/pkg/casemanager"
	"resttracefuzzer/pkg/feedback"
	"resttracefuzzer/pkg/parser"
	"resttracefuzzer/pkg/static"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	// Initialize logger with default log level (Info)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Info().Msg("Hello, World!")

	// Parse command line arguments and environment variables
	config.InitConfig()
	config.ParseCmdArgs()

	// Override log level if specified in the command line arguments
	logLevels := map[string]zerolog.Level{
		"": 	 zerolog.InfoLevel,
		"info":  zerolog.InfoLevel,
		"debug": zerolog.DebugLevel,
		"warn":  zerolog.WarnLevel,
		"error": zerolog.ErrorLevel,
		"fatal": zerolog.FatalLevel,
		"panic": zerolog.PanicLevel,
	}

	if level, exists := logLevels[config.GlobalConfig.LogLevel]; exists {
		zerolog.SetGlobalLevel(level)
	} else {
		log.Error().Msgf("[main] Unsupported log level: %s", config.GlobalConfig.LogLevel)
		return
	}

	apiManager := static.NewAPIManager()

	// read OpenAPI spec and parse it
	apiParser := parser.NewOpenAPIParser()
	doc, err := apiParser.ParseFromPath(config.GlobalConfig.OpenAPISpecPath)
	if err != nil {
		log.Error().Msgf("[main] Failed to parse OpenAPI spec: %v", err)
		return
	}
	apiManager.InitFromSystemDoc(doc)

	// Initialize case manager and response checker
	caseManager := casemanager.NewCaseManager(apiManager)
	responseChecker := feedback.NewResponseChecker(apiManager)

	// Read API dependency files
	// You can generate the dependency files by running Restler
	// We only parse Restler's output for now
	// TODO: parse other dependency files @xunzhou24
	var dependencyFileParser parser.APIDependencyParser
	if config.GlobalConfig.DependencyFileType != "" {
		if config.GlobalConfig.DependencyFileType == "Restler" {
			dependencyFileParser = parser.NewAPIDependencyRestlerParser()
		} else {
			log.Error().Msgf("[main] Unsupported dependency file type: %s", config.GlobalConfig.DependencyFileType)
			return
		}
		dependecyGraph, err := dependencyFileParser.ParseFromPath(config.GlobalConfig.DependencyFilePath)
		if err != nil {
			log.Error().Msgf("Failed to parse dependency file: %v", err)
			return
		}
		apiManager.APIDependencyGraph = dependecyGraph
	}

	// start fuzzing loop
	var mainFuzzer fuzzer.Fuzzer
	if config.GlobalConfig.FuzzerType == "Basic" {
		mainFuzzer = fuzzer.NewBasicFuzzer(
			apiManager,
			caseManager,
			responseChecker,
			feedback.NewTraceManager(),
		)
	} else {
		log.Error().Msgf("[main] Unsupported fuzzer type: %s", config.GlobalConfig.FuzzerType)
		return
	}
	err = mainFuzzer.Start()
	if err != nil {
		log.Error().Msgf("[main] Fuzzer failed: %v", err)
		return
	}

	// generate result report
	// TODO @xunzhou24

	log.Info().Msg("[main] Fuzzing completed")
}
