package parser

import (
	"os"

	"github.com/bytedance/sonic"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

// OpenAPIParser is an OpenAPI parser that parses OpenAPI spec files.
type OpenAPIParser struct {
	loader *openapi3.Loader
}

// NewOpenAPIParser creates a new OpenAPIParser.
func NewOpenAPIParser() *OpenAPIParser {
	parser := &OpenAPIParser{}
	parser.init()
	return parser
}

// init initializes the OpenAPIParser.
func (p *OpenAPIParser) init() {
	p.loader = openapi3.NewLoader()
}

// ParseSystemDocFromPath parses an OpenAPI spec file from the given path.
// It returns the OpenAPI spec and an error if any.
func (p *OpenAPIParser) ParseSystemDocFromPath(path string) (*openapi3.T, error) {
	return p.loader.LoadFromFile(path)
}

// ParseServiceDocFromMapPath parses OpenAPI spec files from the given map path.
// The map path is a JSON file that contains a map of service names to OpenAPI spec file paths.
// It returns a map of service names to OpenAPI specs and an error if any.
func (p *OpenAPIParser) ParseServiceDocFromMapPath(path string) (map[string]*openapi3.T, error) {
	// Read the config file
	fileContent, err := os.ReadFile(path)
	if err != nil {
		log.Error().Msgf("[OpenAPIParser.ParseServiceDocFromMapPath] Error reading file: %v", err)
		return nil, err
	}

	// Parse the config file to a map
	var serviceMap map[string]string
	err = sonic.Unmarshal(fileContent, &serviceMap)
	if err != nil {
		log.Error().Msgf("[OpenAPIParser.ParseServiceDocFromMapPath] Error unmarshalling file: %v", err)
		return nil, err
	}

	// Initialize the result map
	result := make(map[string]*openapi3.T)

	// Iterate over the service map and load each file
	for serviceName, filePath := range serviceMap {
		parsedDoc, err := p.loader.LoadFromFile(filePath)
		if err != nil {
			log.Error().Msgf("[OpenAPIParser.ParseServiceDocFromMapPath] Error loading file: %v", err)
			continue
		}
		log.Info().Msgf("[OpenAPIParser.ParseServiceDocFromMapPath] Loaded OpenAPI spec for service %s", serviceName)
		result[serviceName] = parsedDoc
	}
	
	return result, nil
}
