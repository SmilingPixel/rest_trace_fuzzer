package parser

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

// ParseSystemDocFromPath parses an OpenAPI spec file from the given path.
// It returns the OpenAPI spec and an error if any.
func (p *OpenAPIParser) ParseSystemDocFromPath(path string) (*openapi3.T, error) {
	return p.loader.LoadFromFile(path)
}

// ParseServiceDocFromMapPath parses OpenAPI spec file from the given path.
// It returns a map of service names to OpenAPI specs and an error if any.
func (p *OpenAPIParser) ParseServiceDocFromPath(path string) (*openapi3.T, error) {
	return p.loader.LoadFromFile(path)
}
