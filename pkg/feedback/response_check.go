package feedback

import (
	"resttracefuzzer/pkg/static"

	"github.com/rs/zerolog/log"
)

// ResponseChecker checks the response status code.
type ResponseChecker struct {
	// StatusHitCount is the hit count of the status code.
	// It maps the status code to the hit count.
	StatusHitCount map[static.SimpleAPIMethod]map[int]int

	// ResponseChecker requires an APIManager to initialize.
	APIManager *static.APIManager
}

// NewResponseChecker creates a new ResponseChecker.
// An OpenAPI document is required to initialize the ResponseChecker.
func NewResponseChecker(APIManager *static.APIManager) *ResponseChecker {
	counter := make(map[static.SimpleAPIMethod]map[int]int)
	// TODO: initialize the counter with the status code in the OpenAPI document. @xunzhou24
	return &ResponseChecker{
		StatusHitCount: counter,
		APIManager:     APIManager,
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

// GetCoveredStatusCodeCount returns the covered status codes.
func (rc *ResponseChecker) GetCoveredStatusCodeCount() int {
	count := 0
	for _, statusMap := range rc.StatusHitCount {
		count += len(statusMap)
	}
	return count
}

