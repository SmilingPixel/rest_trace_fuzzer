package parser

import "resttracefuzzer/pkg/static"

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
