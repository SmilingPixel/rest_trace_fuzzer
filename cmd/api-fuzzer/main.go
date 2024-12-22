package main

import (
	"fmt"
	"resttracefuzzer/internal/config"
	"resttracefuzzer/pkg/logger"
	"resttracefuzzer/pkg/parser"

    "github.com/rs/zerolog/log"
)

func main() {

	logger.ConfigLogger()
    log.Info().Msg("Hello, World!")

	// Parse command line arguments and environment variables
	config.InitConfig()
	config.ParseCmdArgs()

	// read OpenAPI spec and parse it
	apiParser := parser.NewOpenAPIParser()
	_, err := apiParser.ParseFromPath(config.GlobalConfig.OpenAPISpecPath)
	if err != nil {
		fmt.Println("Failed to parse OpenAPI spec")
		return
	}

	// Read API dependency files
	// You can generate the dependency files by running Restler
	// TODO @xunzhou24

	// set hyperparameters
	// TODO @xunzhou24

	// start fuzzing loop
	// TODO @xunzhou24

	// generate result report
	// TODO @xunzhou24

	log.Info().Msg("Fuzzing completed")
}
