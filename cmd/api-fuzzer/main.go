package main

import (
	"resttracefuzzer/internal/config"
	"resttracefuzzer/pkg/logger"
	"resttracefuzzer/pkg/parser"
	"resttracefuzzer/pkg/static"

	"github.com/rs/zerolog/log"
)

func main() {

	// Initialize logger
	logger.ConfigLogger()
	log.Info().Msg("Hello, World!")

	// Parse command line arguments and environment variables
	config.InitConfig()
	config.ParseCmdArgs()

	apiManager := static.NewAPIManager()

	// read OpenAPI spec and parse it
	apiParser := parser.NewOpenAPIParser()
	doc, err := apiParser.ParseFromPath(config.GlobalConfig.OpenAPISpecPath)
	if err != nil {
		log.Error().Msgf("Failed to parse OpenAPI spec: %v", err)
		return
	}
	apiManager.APIDefinition = doc

	// Read API dependency files
	// You can generate the dependency files by running Restler
	// We only parse Restler's output for now
	// TODO: parse other dependency files @xunzhou24
	var dependencyFileParser parser.APIDependencyParser
	if config.GlobalConfig.DependencyFileType == "Restler" {
		dependencyFileParser = parser.NewAPIDependencyRestlerParser()
	} else {
		log.Error().Msgf("Unsupported dependency file type: %s", config.GlobalConfig.DependencyFileType)
		return
	}
	dependecyGraph, err := dependencyFileParser.ParseFromPath(config.GlobalConfig.DependencyFilePath)
	if err != nil {
		log.Error().Msgf("Failed to parse dependency file: %v", err)
		return
	}
	apiManager.APIDependencyGraph = dependecyGraph

	// set hyperparameters
	// TODO @xunzhou24

	// start fuzzing loop
	// TODO @xunzhou24

	// generate result report
	// TODO @xunzhou24

	log.Info().Msg("Fuzzing completed")
}
