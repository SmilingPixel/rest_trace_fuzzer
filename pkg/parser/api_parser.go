package parser;

import (
	"github.com/getkin/kin-openapi/openapi3"
)


type OpenAPIParser struct {
	loader *openapi3.Loader
}


func NewOpenAPIParser() *OpenAPIParser {
	parser := &OpenAPIParser{}
	parser.init()
	return parser
}

func (p *OpenAPIParser) init() {
	p.loader = openapi3.NewLoader()
}

func (p *OpenAPIParser) ParseFromPath(path string) (*openapi3.T, error) {
	return p.loader.LoadFromFile(path)
}
