package feedback

import (
	"resttracefuzzer/pkg/resource"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"
	"resttracefuzzer/pkg/utils/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/rs/zerolog/log"
)

// ResponseProcesser process the response.
type ResponseProcesser struct {
	// StatusHitCount is the hit count of the status code.
	// It maps the status code to the hit count.
	StatusHitCount map[static.SimpleAPIMethod]map[int]int

	// ResponseProcesser requires an APIManager to initialize.
	APIManager *static.APIManager

	// The Resource Manager. ResponseProcesser will extract resource from response, and store it in the resource manager.
	ResourceManager *resource.ResourceManager
}

// NewResponseProcesser creates a new ResponseProcesser.
// An OpenAPI document is required to initialize the ResponseProcesser.
func NewResponseProcesser(APIManager *static.APIManager, resourceManager *resource.ResourceManager) *ResponseProcesser {
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
				log.Debug().Msgf("[NewResponseProcesser] Failed to parse field key %s as int", fieldKey)
				continue
			}
			counter[method][statusCode] = 0
		}
	}
	return &ResponseProcesser{
		StatusHitCount:  counter,
		APIManager:      APIManager,
		ResourceManager: resourceManager,
	}
}

// ProcessResponse checks and processes the response status code and response body.
// If the status exists in the OpenAPI document, the hit count will be increased.
// Otherwise, it will log a warning.
// If a successful response is received, the resource will be extracted and stored in the resource manager.
func (rc *ResponseProcesser) ProcessResponse(method static.SimpleAPIMethod, statusCode int, responseBody []byte) error {
	// handle status code
	if _, ok := rc.StatusHitCount[method]; !ok {
		log.Warn().Msgf("[ResponseProcesser.ProcessResponse] Method %s %s is not in the OpenAPI document", method.Method, method.Endpoint)
		return nil
	}
	rc.StatusHitCount[method][statusCode]++

	// handle response body
	if http.GetStatusCodeClass(statusCode) == consts.StatusOK {
		// when storing resources, we use the API method as the root resource name.
		// For example, if the API method is "GET /api/v1/user", the root resource name will be "user".
		resourceName := utils.ExtractLastSegment(method.Endpoint, "/")
		err := rc.ResourceManager.StoreResourcesFromRawObjectBytes(responseBody, resourceName, true)
		if err != nil {
			log.Err(err).Msg("[ResponseProcesser.ProcessResponse] Failed to store resources")
			return err
		}
	}
	return nil
}

// GetCoveredStatusCodeCount returns the covered status codes.
func (rc *ResponseProcesser) GetCoveredStatusCodeCount() int {
	count := 0
	for _, statusMap := range rc.StatusHitCount {
		for _, hit := range statusMap {
			if hit > 0 {
				count++
			}
		}
	}
	return count
}
