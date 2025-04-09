package main

import (
	"fmt"
	"os"
	"resttracefuzzer/internal/config"
	"resttracefuzzer/internal/fuzzer"
	"resttracefuzzer/pkg/casemanager"
	"resttracefuzzer/pkg/feedback"
	"resttracefuzzer/pkg/feedback/trace"
	"resttracefuzzer/pkg/parser"
	"resttracefuzzer/pkg/report"
	"resttracefuzzer/pkg/resource"
	fuzzruntime "resttracefuzzer/pkg/runtime"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/strategy"
	"time"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// The ASCII art of "HELLO" is generated by https://patorjk.com/software/taag/
const HELLO = "" +
	" .----------------.  .----------------.  .----------------.  .----------------.  .----------------. \n" +
	"| .--------------. || .--------------. || .--------------. || .--------------. || .--------------. |\n" +
	"| |  ____  ____  | || |  _________   | || |   _____      | || |   _____      | || |     ____     | |\n" +
	"| | |_   ||   _| | || | |_   ___  |  | || |  |_   _|     | || |  |_   _|     | || |   .'    `.   | |\n" +
	"| |   | |__| |   | || |   | |_  \\_|  | || |    | |       | || |    | |       | || |  /  .--.  \\  | |\n" +
	"| |   |  __  |   | || |   |  _|  _   | || |    | |   _   | || |    | |   _   | || |  | |    | |  | |\n" +
	"| |  _| |  | |_  | || |  _| |___/ |  | || |   _| |__/ |  | || |   _| |__/ |  | || |  \\  `--'  /  | |\n" +
	"| | |____||____| | || | |_________|  | || |  |________|  | || |  |________|  | || |   `.____.'   | |\n" +
	"| |              | || |              | || |              | || |              | || |              | |\n" +
	"| '--------------' || '--------------' || '--------------' || '--------------' || '--------------' |\n" +
	" '----------------'  '----------------'  '----------------'  '----------------'  '----------------' \n"

func main() {

	// Record the start time
	t := time.Now()

	// Initialize logger with default log level (Info)
	// If user specifies a log level in the command line arguments, we will override it later
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	fmt.Print(HELLO)

	// Parse command line arguments and environment variables
	config.InitConfig()
	config.ParseCmdArgs()

	// Override log level if specified in the command line arguments
	logLevels := map[string]zerolog.Level{
		"":      zerolog.InfoLevel, // Default log level
		"trace": zerolog.TraceLevel,
		"debug": zerolog.DebugLevel,
		"info":  zerolog.InfoLevel,
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

	// used to format the time in the output file name
	// We do not use RFC3339 format because it contains colons, which are not allowed in Windows file names.
	outputFileTimeFormat := "20060102150405"

	// Log to file if specified
	if config.GlobalConfig.LogToFile {
		logFilePath := fmt.Sprintf("%s/log_%s.log", config.GlobalConfig.OutputDir, t.Format(outputFileTimeFormat))
		fileWriter, err := os.Create(logFilePath)
		if err != nil {
			log.Err(err).Msgf("[main] Failed to create log file: %s", logFilePath)
			return
		}
		log.Info().Msgf("[main] Log to file is enabled, I will write logs to %s", logFilePath)
		log.Logger = log.Output(fileWriter)
	}

	APIManager := static.NewAPIManager()

	// read OpenAPI spec and parse it
	APIParser := parser.NewOpenAPIParser()
	doc, err := APIParser.ParseSystemDocFromPath(config.GlobalConfig.OpenAPISpecPath)
	if err != nil {
		log.Err(err).Msgf("[main] Failed to parse OpenAPI spec")
		return
	}
	APIManager.InitFromSystemDoc(doc)

	// Parse doc of internal services
	serviceDoc, err := APIParser.ParseServiceDocFromPath(config.GlobalConfig.InternalServiceOpenAPIPath)
	if err != nil {
		log.Err(err).Msgf("[main] Failed to parse internal service OpenAPI spec")
		return
	}
	APIManager.InitFromServiceDoc(serviceDoc)

	// Parse extra headers
	extraHeaders := make(map[string]string)
	if config.GlobalConfig.ExtraHeaders != "" {
		err = sonic.UnmarshalString(config.GlobalConfig.ExtraHeaders, &extraHeaders)
		if err != nil {
			log.Err(err).Msgf("[main] Failed to parse extra headers")
			return
		}
	}

	// Initialize necessary components
	resourceManager := resource.NewResourceManager()
	if config.GlobalConfig.FuzzValueDictFilePath != "" {
		err = resourceManager.LoadFromExternalDictFile(config.GlobalConfig.FuzzValueDictFilePath)
		// If failed to load resources from external dictionary file, log the error;
		// but continue the fuzzing process
		if err != nil {
			log.Err(err).Msgf("[main] Failed to load resources from external dictionary file")
		}
	}
	fuzzStrategist := strategy.NewFuzzStrategist(resourceManager)
	resourceMutateStrategist := strategy.NewResourceMutateStrategy()
	caseManager := casemanager.NewCaseManager(APIManager, resourceManager, fuzzStrategist, resourceMutateStrategist, extraHeaders)
	responseProcesser := feedback.NewResponseProcesser(APIManager, resourceManager)
	traceManager := trace.NewTraceManager()
	callInfoGraph := fuzzruntime.NewCallInfoGraph(APIManager.APIDataflowGraph)
	reachabilityMap := fuzzruntime.NewRuntimeReachabilityMapFromStaticMap(APIManager.StaticReachabilityMap)

	// Read API dependency files
	// You can generate the dependency files by running Restler
	// We only parse Restler's output for now
	// TODO: parse other dependency files @xunzhou24
	var dependencyFileParser parser.APIDependencyParser
	if config.GlobalConfig.DependencyFileType != "" {
		if config.GlobalConfig.DependencyFileType == "Restler" {
			dependencyFileParser = parser.NewAPIDependencyRestlerParser()
		} else {
			log.Err(err).Msgf("[main] Unsupported dependency file type: %s", config.GlobalConfig.DependencyFileType)
			return
		}
		dependecyGraph, err := dependencyFileParser.ParseFromPath(config.GlobalConfig.DependencyFilePath)
		if err != nil {
			log.Err(err).Msgf("Failed to parse dependency file")
			return
		}
		APIManager.APIDependencyGraph = dependecyGraph
	}

	// testLogReporter logs the tested operations
	testLogReporter := report.NewTestLogReporter()

	// start fuzzing loop
	var mainFuzzer fuzzer.Fuzzer
	if config.GlobalConfig.FuzzerType == "Basic" {
		mainFuzzer = fuzzer.NewBasicFuzzer(
			APIManager,
			caseManager,
			responseProcesser,
			traceManager,
			callInfoGraph,
			reachabilityMap,
			testLogReporter,
		)
	} else {
		log.Err(err).Msgf("[main] Unsupported fuzzer type: %s", config.GlobalConfig.FuzzerType)
		return
	}
	err = mainFuzzer.Start()
	if err != nil {
		log.Err(err).Msgf("[main] Fuzzer failed")
		return
	}

	// generate result report
	// Reports are named using current timestamp, in yyyyMMddHHmmss format,
	// with prefix "system_report_", "internal_service_report_", etc.
	// The reports are saved in the output directory
	// Create the output directory if it does not exist.
	err = os.MkdirAll(config.GlobalConfig.OutputDir, os.ModePerm)
	if err != nil {
		log.Err(err).Msgf("[main] Failed to create the output directory")
		return
	}
	systemReporter := report.NewSystemReporter(APIManager)
	systemReportPath := fmt.Sprintf("%s/system_report_%s.json", config.GlobalConfig.OutputDir, t.Format(outputFileTimeFormat))
	err = systemReporter.GenerateSystemReport(responseProcesser, systemReportPath)
	if err != nil {
		log.Err(err).Msgf("[main] Failed to generate system report")
		return
	}
	internalServiceReporter := report.NewInternalServiceReporter()
	internalServiceReportPath := fmt.Sprintf("%s/internal_service_report_%s.json", config.GlobalConfig.OutputDir, t.Format(outputFileTimeFormat))
	err = internalServiceReporter.GenerateInternalServiceReport(mainFuzzer.GetCallInfoGraph(), internalServiceReportPath)
	if err != nil {
		log.Err(err).Msgf("[main] Failed to generate internal service report")
		return
	}
	fuzzerStateReporter := report.NewFuzzerStateReporter()
	fuzzerStateReportPath := fmt.Sprintf("%s/fuzzer_state_report_%s.json", config.GlobalConfig.OutputDir, t.Format(outputFileTimeFormat))
	err = fuzzerStateReporter.GenerateFuzzerStateReport(resourceManager, fuzzerStateReportPath)
	if err != nil {
		log.Err(err).Msgf("[main] Failed to generate fuzzer state report")
		return
	}
	testLogReportPath := fmt.Sprintf("%s/test_log_report_%s.json", config.GlobalConfig.OutputDir, t.Format(outputFileTimeFormat))
	err = testLogReporter.GenerateTestLogReport(testLogReportPath)
	if err != nil {
		log.Err(err).Msgf("[main] Failed to generate test log report")
		return
	}

	log.Info().Msg("[main] Fuzzing completed")
}
