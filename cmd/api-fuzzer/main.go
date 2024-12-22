package main

import (
	"fmt"
	"resttracefuzzer/pkg/logger"
	"resttracefuzzer/pkg/parser"

    "github.com/rs/zerolog/log"
)

func main() {

	logger.ConfigLogger()
    log.Info().Msg("Hello, World!")

	// read OpenAPI spec and parse it
	apiParser := parser.NewOpenAPIParser()
	// TODO: we hard code the path here, we should make it configurable @xunzhou24
	_, err := apiParser.ParseFromPath("path/to/openapi/spec")
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
