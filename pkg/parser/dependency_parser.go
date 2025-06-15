package parser

import (
	"fmt"
	"resttracefuzzer/pkg/static"
)

// APIDependencyParser is an interface for parsing API dependencies.
type APIDependencyParser interface {
	// ParseFromFile parses the API dependency graph from the given file path.
	ParseFromFile(path string) (*static.APIDependencyGraph, error)

	// ParseFromBytes parses the API dependency graph from the given byte slice.
	ParseFromBytes(data []byte) (*static.APIDependencyGraph, error)

	// ParseFromServiceMapFile parses the API dependency graph from the given service map file.
	// The file is a JSON file that contains a map of service names to their corresponding API dependencies.
	ParseFromServiceMapFile(path string) (map[string]*static.APIDependencyGraph, error)
}

// NewAPIDependencyParserByType creates a new APIDependencyParser instance based on the given parser type.
func NewAPIDependencyParserByType(parserType string) (APIDependencyParser, error) {
	// We support Restler parser for now
	// You can contact us if you want to add support for other parsers
	switch parserType {
	case "Restler":
		return NewAPIDependencyRestlerParser(), nil
	default:
		return nil, fmt.Errorf("unsupported parser type: %s", parserType)
	}
}
