package http

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog/log"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"maps"
)

// HTTPClientMiddleware defines the interface for HTTP client middleware.
// Middleware allows you to intercept and modify HTTP requests and responses.
// It can be used for logging, authentication, modifying headers, etc.
type HTTPClientMiddleware interface {
	// HandleRequest processes the HTTP request.
    // It takes the request path, method, headers, path parameters, query parameters, and body as input.
    // It returns the modified request path, method, headers, path parameters, query parameters, body, and an error if any.
    HandleRequest(path, method string, headers map[string]string, pathParams, queryParams map[string]string, body []byte) (resPath, resMethod string, resHeaders map[string]string, resPathParams, resQueryParams map[string]string, resBody []byte, err error)
}

// EmptyHTTPClientMiddlewareSlice returns an empty slice of HTTPClientMiddleware.
// It is used when a HTTP Client has no middleware.
func EmptyHTTPClientMiddlewareSlice() []HTTPClientMiddleware {
	return make([]HTTPClientMiddleware, 0)
}


// HTTPClientScriptMiddleware is a middleware that runs a Starlark script to handle HTTP requests.
// The script can modify the request and response by returning modified values for headers, path parameters, query parameters, and body.
// The script should define global variables "headers", "pathParams", "queryParams", and "body" to return the modified values.
// headers, pathParams, and queryParams should be a Dict, and body should be a string.
// For example:
//  # Example Starlark script
//  headers = {"Authorization": "Bearer new_token"}
//  pathParams = {"id": "123"}
//  queryParams = {"search": "new_query"}
//  body = "[1, 2, 3]"
// It can be used for logging, authentication, modifying headers, etc.
// For how to write Starlark scripts, see: https://github.com/google/starlark-go/blob/master/doc/spec.md
type HTTPClientScriptMiddleware struct {

	// ScriptPath is the path to the Starlark script.
	// Path is actually used during initialization to load the script content.
	// When executing the script, the content is used, and the path is used only for logging.
	ScriptPath string `json:"scriptPath"`

	// Script is the content of the Starlark script.
	Script []byte `json:"script"`
}

// NewHTTPClientMiddleware creates a new HTTPClientScriptMiddleware.
// It takes script path as a parameter and returns an instance of HTTPClientScriptMiddleware.
func NewHTTPClientScriptMiddleware(scriptPath string) *HTTPClientScriptMiddleware {
	// Load the script
	file, err := os.Open(scriptPath)
	if err != nil {
		log.Err(err).Msgf("[NewHTTPClientScriptMiddleware] Failed to open file: %s", scriptPath)
		return nil
	}
	defer file.Close()

	script, err := io.ReadAll(file)
	if err != nil {
		log.Err(err).Msgf("[NewHTTPClientScriptMiddleware] Failed to load script from path: %s", scriptPath)
		return nil
	}

	return &HTTPClientScriptMiddleware{
		ScriptPath: scriptPath,
		Script: script,
	}
}

// HandleRequest runs the Starlark script to handle the request.
// The script can modify the request by returning modified values for headers, path parameters, query parameters, and body.
// The script should define global variables "headers", "pathParams", "queryParams", and "body" to return the modified values.
// It returns the modified request path, method, headers, path parameters, query parameters, body, and an error if any.
func (m *HTTPClientScriptMiddleware) HandleRequest(path, method string, headers map[string]string, pathParams, queryParams map[string]string, body []byte) (resPath, resMethod string, resHeaders map[string]string, resPathParams, resQueryParams map[string]string, resBody []byte, err error) {
	// Try to run the script
	thread := &starlark.Thread{Name: "http_middleware_script"}
	fileOptions := syntax.LegacyFileOptions()
	globals, err := starlark.ExecFileOptions(fileOptions, thread, m.ScriptPath, m.Script, nil)
	if err != nil {
		log.Err(err).Msg("[HTTPClientScriptMiddleware.HandleRequest] Failed to execute script")
		return path, method, headers, pathParams, queryParams, body, err
	}

	// Extract the results
	if res, ok := globals["headers"]; ok {
		if headersMap, isMap := res.(*starlark.Dict); isMap {
			extraHeaders, err := convertStarlarkMapToStringMap(headersMap)
			if err != nil {
				log.Err(err).Msg("[HTTPClientScriptMiddleware.HandleRequest] Failed to convert headers map")
				return path, method, headers, pathParams, queryParams, body, err
			}
			log.Debug().Msgf("[HTTPClientScriptMiddleware.HandleRequest] Got extra headers: %v", extraHeaders)
			maps.Copy(headers, extraHeaders)
		} else {
			log.Warn().Msg("[HTTPClientScriptMiddleware.HandleRequest] headers is not a map")
		}
	}

	if res, ok := globals["pathParams"]; ok {
		if pathParamsMap, isMap := res.(*starlark.Dict); isMap {
			extraPathParams, err := convertStarlarkMapToStringMap(pathParamsMap)
			if err != nil {
				log.Err(err).Msg("[HTTPClientScriptMiddleware.HandleRequest] Failed to convert pathParams map")
				return path, method, headers, pathParams, queryParams, body, err
			}
			log.Debug().Msgf("[HTTPClientScriptMiddleware.HandleRequest] Got extra path params: %v", extraPathParams)
			maps.Copy(pathParams, extraPathParams)
		} else {
			log.Warn().Msg("[HTTPClientScriptMiddleware.HandleRequest] pathParams is not a map")
		}
	}

	if res, ok := globals["queryParams"]; ok {
		if queryParamsMap, isMap := res.(*starlark.Dict); isMap {
			extraQueryParams, err := convertStarlarkMapToStringMap(queryParamsMap)
			if err != nil {
				log.Err(err).Msg("[HTTPClientScriptMiddleware.HandleRequest] Failed to convert queryParams map")
				return path, method, headers, pathParams, queryParams, body, err
			}
			log.Debug().Msgf("[HTTPClientScriptMiddleware.HandleRequest] Got extra query params: %v", extraQueryParams)
			maps.Copy(queryParams, extraQueryParams)
		} else {
			log.Warn().Msg("[HTTPClientScriptMiddleware.HandleRequest] queryParams is not a map")
		}
	}
	if res, ok := globals["body"]; ok {
		if str, isStr := res.(starlark.String); isStr {
			// Use GoString() to get the raw string value without extra quotes
			rawBody := string(str.GoString())
			log.Debug().Msgf("[HTTPClientScriptMiddleware.HandleRequest] Got body: %s", rawBody)
			body = []byte(rawBody)
		} else {
			// Handle non-string body values (e.g., numbers, lists, etc.)
			log.Warn().Msgf("[HTTPClientScriptMiddleware.HandleRequest] Body is not a string: %s", res.String())
			body = []byte(res.String())
		}
	}
	return path, method, headers, pathParams, queryParams, body, nil
}

// Helper function to convert a Starlark map to a Go map[string]string
func convertStarlarkMapToStringMap(starlarkMap *starlark.Dict) (map[string]string, error) {
	goMap := make(map[string]string)
	for _, key := range starlarkMap.Keys() {
		value, found, err := starlarkMap.Get(key)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		keyStr, ok := key.(starlark.String)
		if !ok {
			return nil, fmt.Errorf("map key is not a string: %v", key)
		}
		valueStr, valueIsStr := value.(starlark.String)
		if valueIsStr {
			goMap[keyStr.GoString()] = valueStr.GoString()
		} else {
			goMap[keyStr.GoString()] = value.String()
		}
	}
	return goMap, nil
}
