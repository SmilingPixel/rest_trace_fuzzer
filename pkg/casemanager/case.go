package casemanager

import (
	"resttracefuzzer/pkg/apimanager"

	"github.com/getkin/kin-openapi/openapi3"
)

// An OperationCase is a pair of an API method and an operation.
type OperationCase struct {
	ServiceName string
	APIMethod *apimanager.SimpleAPIMethod
	Operation *openapi3.Operation
}

// A Testcase is a sequence of operations.
type Testcase struct {
	OperationCases []*OperationCase
}