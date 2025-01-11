package feedback

import (
	"resttracefuzzer/pkg/static"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)


type ResponseChecker struct {
	// StatusHitCount is the hit count of the status code.
	// It maps the status code to the hit count.
	StatusHitCount map[static.SimpleAPIMethod]map[int]int
}

// NewResponseChecker creates a new ResponseChecker.
// An OpenAPI document is required to initialize the ResponseChecker.
func NewResponseChecker(doc openapi3.T) *ResponseChecker {
	counter := make(map[static.SimpleAPIMethod]map[int]int)
	// TODO: initialize the counter with the status code in the OpenAPI document.
	return &ResponseChecker{
		StatusHitCount: counter,
	}
}

// CheckResponse checks the response status code.
// If the status exists in the OpenAPI document, the hit count will be increased.
// Otherwise, it will log a warning.
func (rc *ResponseChecker) CheckResponse(method static.SimpleAPIMethod, statusCode int) error {
	// TODO: implement the CheckResponse method. @xunzhou24
	if _, ok := rc.StatusHitCount[method]; !ok {
		log.Warn().Msgf("Method %s %s is not in the OpenAPI document", method.Method, method.Endpoint)
		return nil
	}
	rc.StatusHitCount[method][statusCode]++
	return nil
}
