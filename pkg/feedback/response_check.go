package feedback

import (
	"resttracefuzzer/pkg/static"
	"strconv"

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
	// Initialize the hit count map according to the OpenAPI document.
	// We can utilize the pre-processed APIMap in the APIManager to make it.
	counter := make(map[static.SimpleAPIMethod]map[int]int)
	APIMap := APIManager.APIMap
	for method, operation := range APIMap {
		counter[method] = make(map[int]int)
		// fieldKey is the status code or 'default'. See [OpenAPI 3.0](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#responses-object).
		// We only care about the status code, and ignore 'default'.
		for fieldKey := range operation.Responses.Map() {
			statusCode, err := strconv.Atoi(fieldKey)
			if err != nil {
				log.Debug().Msgf("[NewResponseChecker] Failed to parse field key %s as int", fieldKey)
				continue
			}
			counter[method][statusCode] = 0
		}
	}
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
