package config

import (
	"flag"
	// "fmt"
	// "os"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

func ParseCmdArgs() {
	flag.StringVar(&GlobalConfig.OpenAPISpecPath, "openapi-spec", "", "Path to the OpenAPI spec file")
	flag.Parse()

	jsonStr, _ := sonic.Marshal(GlobalConfig)
	log.Info().Msgf("Parsed command line arguments: %s", jsonStr)

}
	
