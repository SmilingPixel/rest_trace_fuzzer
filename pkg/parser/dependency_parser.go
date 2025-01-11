package parser

import "resttracefuzzer/pkg/static"

// "github.com/bytedance/sonic"

type APIDependencyParser interface {
	ParseFromPath(path string) (*static.APIDependencyGraph, error)
}
