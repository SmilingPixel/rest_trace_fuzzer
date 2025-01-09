package casemanager

import (
	"resttracefuzzer/pkg/static"

	"github.com/getkin/kin-openapi/openapi3"
)

// An OperationCase is a pair of an API method and an operation.
type OperationCase struct {
	APIMethod *static.SimpleAPIMethod
	Operation *openapi3.Operation
}

// A Testcase is a sequence of [resttracefuzzer/pkg/casemanager/OperationCase].
type Testcase struct {
	OperationCases []*OperationCase
}
