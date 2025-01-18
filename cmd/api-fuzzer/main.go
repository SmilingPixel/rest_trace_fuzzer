package main

import (
	"fmt"
	"resttracefuzzer/internal/config"
	"resttracefuzzer/internal/fuzzer"
	"resttracefuzzer/pkg/casemanager"
	"resttracefuzzer/pkg/feedback"
	"resttracefuzzer/pkg/parser"
	"resttracefuzzer/pkg/report"
	"resttracefuzzer/pkg/static"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const HELLO = `
 x                                                
 x                                                
 x            x                x     x            
 x           xx                x     x            
 x           x                 x    xx            
xx           x                x     x             
x           xx     xxxxxx     x    xx             
x   xxxxxxxxx     xx    x     x    x      xxxxx   
x           x    xx   xx     xx    x     xx   xxx 
x          xx    xxxxxx      x     x     x      xx
x          x     x           x     x     x       x
x          x      xx         x     x     x      xx
x          x       xxxxxxx   xx    xx    xxx   xx 
 x         x                               xxxxx  

`

func main() {

	// Initialize logger with default log level (Info)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	fmt.Print(HELLO)

	// Parse command line arguments and environment variables
	config.InitConfig()
	config.ParseCmdArgs()

	// Override log level if specified in the command line arguments
	logLevels := map[string]zerolog.Level{
		"":      zerolog.InfoLevel,
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
	doc, err := apiParser.ParseSystemDocFromPath(config.GlobalConfig.OpenAPISpecPath)
	if err != nil {
		log.Error().Err(err).Msgf("[main] Failed to parse OpenAPI spec: %v", err)
		return
	}
	apiManager.InitFromSystemDoc(doc)

	// Parse doc of internal services
	serviceDoc, err := apiParser.ParseServiceDocFromPath(config.GlobalConfig.InternalServiceOpenAPIPath)
	if err != nil {
		log.Error().Err(err).Msgf("[main] Failed to parse internal service OpenAPI spec: %v", err)
		return
	}
	apiManager.InitFromServiceDoc(serviceDoc)

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
			log.Error().Err(err).Msgf("[main] Unsupported dependency file type: %s", config.GlobalConfig.DependencyFileType)
			return
		}
		dependecyGraph, err := dependencyFileParser.ParseFromPath(config.GlobalConfig.DependencyFilePath)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to parse dependency file: %v", err)
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
		log.Error().Err(err).Msgf("[main] Unsupported fuzzer type: %s", config.GlobalConfig.FuzzerType)
		return
	}
	err = mainFuzzer.Start()
	if err != nil {
		log.Error().Err(err).Msgf("[main] Fuzzer failed: %v", err)
		return
	}

	// generate result report
	// Reports are named using current timestamp, in RFC3339 format,
	// with prefix "system_report_", "internal_service_report_", etc.
	// The reports are saved in the output directory
	// TODO @xunzhou24
	t := time.Now()
	systemReporter := report.NewSystemReporter(apiManager)
	systemReportPath := fmt.Sprintf("%s/system_report_%s.json", config.GlobalConfig.OutputDir, t.Format(time.RFC3339))
	err = systemReporter.GenerateSystemReport(responseChecker, systemReportPath)
	if err != nil {
		log.Error().Err(err).Msgf("[main] Failed to generate system report: %v", err)
		return
	}
	internalServiceReporter := report.NewInternalServiceReporter()
	internalServiceReportPath := fmt.Sprintf("%s/internal_service_report_%s.json", config.GlobalConfig.OutputDir, t.Format(time.RFC3339))
	err = internalServiceReporter.GenerateInternalServiceReport(mainFuzzer.GetRuntimeGraph(), internalServiceReportPath)
	if err != nil {
		log.Error().Err(err).Msgf("[main] Failed to generate internal service report: %v", err)
		return
	}

	log.Info().Msg("[main] Fuzzing completed")
}
