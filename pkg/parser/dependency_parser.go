package parser

import (
	"resttracefuzzer/pkg/apimanager"

	// "github.com/bytedance/sonic"
)

type APIDependencyParser interface {
	ParseFromPath(path string) (*apimanager.APIDependencyGraph, error)
}
