package parser;

import (
	"github.com/getkin/kin-openapi/openapi3"
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

// ParseFromPath parses an OpenAPI spec file from the given path.
// It returns the OpenAPI spec and an error if any.
func (p *OpenAPIParser) ParseFromPath(path string) (*openapi3.T, error) {
	return p.loader.LoadFromFile(path)
}
