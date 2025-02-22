package utils

import (
	"context"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/rs/zerolog/log"
)

const (
	FUZZER_TRACE_ID_HEADER_KEY = "X-Fuzzer-Trace-ID"
)

// HTTPClient is an HTTP client.
// It has a base URL and a client based on Hertz.
type HTTPClient struct {
	BaseURL string
	Client  *client.Client
}

// NewHTTPClient creates a new HTTPClient.
func NewHTTPClient(baseURL string) *HTTPClient {
	c, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	return &HTTPClient{
		Client:  c,
		BaseURL: baseURL,
	}
}

// PerformRequestWithRetry performs an HTTP request with retry logic.
// It retries the request up to maxRetry times if a timeout error occurs.
// If the request fails for any other reason, it returns the error immediately.
// If all retry attempts fail due to timeout, it logs an error and returns the last error encountered.
func (c *HTTPClient) PerformRequestWithRetry(path, method string, headers map[string]string, pathParams, queryParams map[string]string, body interface{}, maxRetry int) (int, map[string]string, []byte, error) {
	// If maxRetry is invalid, fallback to 1
	if maxRetry <= 0 {
		log.Warn().Msgf("[HTTPClient.PerformRequestWithRetry] Invalid max retry: %d, fallback to 1", maxRetry)
		return c.PerformRequest(path, method, headers, pathParams, queryParams, body)
	}

	// Retry only when timeout
	var err error
	for i := 0; i < maxRetry; i++ {
		statusCode, headers, respBodyBytes, err := c.PerformRequest(path, method, headers, pathParams, queryParams, body)
		if err != nil {
			if strings.Contains(string(err.Error()), "timeout") {
				log.Warn().Msgf("[HTTPClient.PerformRequestWithRetry] Retry %d times due to timeout, URL: %s, method: %s", i+1, c.BaseURL+path, method)
				continue
			} else {
				return statusCode, headers, respBodyBytes, err
			}
		}
		return statusCode, headers, respBodyBytes, nil
	}
	log.Err(err).Msgf("[HTTPClient.PerformRequestWithRetry] Retry %d times but still timeout, URL: %s, method: %s", maxRetry, c.BaseURL+path, method)
	return 0, nil, nil, err
}

// PerformRequest performs an HTTP request.
// It returns the status code, headers that we care about, the response body in bytes, and an error if any.
func (c *HTTPClient) PerformRequest(path, method string, headers map[string]string, pathParams, queryParams map[string]string, body interface{}) (int, map[string]string, []byte, error) {
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	requestURL := c.BaseURL + path
	req.SetRequestURI(requestURL)
	req.SetHeaders(headers)
	req.SetMethod(method)

	// Set path params
	if len(queryParams) > 0 {
		req.SetQueryString(paramDict2QueryStr(queryParams))
	}

	// Set path params, replacing the path params in the URL
	for k, v := range pathParams {
		requestURL = strings.ReplaceAll(requestURL, "{"+k+"}", v)
	}

	bodyBytes, err := sonic.Marshal(body)
	if err != nil {
		log.Err(err).Msgf("[HTTPClient.PerformRequest] Failed to marshal request body, URL: %s, method: %s", requestURL, method)
		return 0, nil, nil, err
	}
	req.SetBody(bodyBytes)

	log.Debug().Msgf("[HTTPClient.PerformRequest] Perform request, URL: %s, method: %s, headers: %v, query params: %v, body: %v", requestURL, method, headers, queryParams, body)
	err = c.Client.Do(context.Background(), req, resp)
	if err != nil {
		log.Err(err).Msgf("[HTTPClient.PerformRequest] Failed to perform request, URL: %s, method: %s", requestURL, method)
		return 0, nil, nil, err
	}
	respBodyBytes, err := resp.BodyE()
	if err != nil {
		log.Err(err).Msgf("[HTTPClient.PerformRequest] Failed to get response body, URL: %s, method: %s", requestURL, method)
		return 0, nil, nil, err
	}
	statusCode := resp.StatusCode()
	log.Debug().Msgf("[HTTPClient.PerformRequest] Response, status code: %d", statusCode) // we do not log response body, for some responses may be too large

	// retrive headers that we care about
	retrivedHeaders := make(map[string]string)
	retrivedHeaders[FUZZER_TRACE_ID_HEADER_KEY] = resp.Header.Get(FUZZER_TRACE_ID_HEADER_KEY)
	return statusCode, retrivedHeaders, respBodyBytes, nil
}

// PerformGet performs an HTTP GET request.
func (c *HTTPClient) PerformGet(path string, headers map[string]string, pathParams, queryParams map[string]string) (int, map[string]string, []byte, error) {
	return c.PerformRequest(path, "GET", headers, pathParams, queryParams, nil)
}

// paramDict2QueryStr converts a map of parameters to a query string.
// It returns the query string.
//
// For example, if the input is {"a": "1", "b": ["2", "3"]}, the output is "a=1&b=2,3".
func paramDict2QueryStr(paramDict map[string]string) string {
	queryParamsStrList := make([]string, 0)
	for k, v := range paramDict {
		queryParamsStrList = append(queryParamsStrList, k+"="+v)
	}
	return strings.Join(queryParamsStrList, "&")
}

// GetStatusCodeClass returns the class of a status code.
// There are five classes defined by the standard:
//  - 1xx: Informational
//  - 2xx: Successful
//  - 3xx: Redirection
//  - 4xx: Client Error
//  - 5xx: Server Error
// It returns minimal code of a class to indicate the class.
// If the status code is not in the standard range, it returns -1.
func GetStatusCodeClass(statusCode int) int {
	switch {
	case statusCode >= consts.StatusContinue && statusCode < consts.StatusOK:
		return consts.StatusContinue
	case statusCode >= consts.StatusOK && statusCode < consts.StatusMultipleChoices:
		return consts.StatusOK
	case statusCode >= consts.StatusMultipleChoices && statusCode < consts.StatusBadRequest:
		return consts.StatusMultipleChoices
	case statusCode >= consts.StatusBadRequest && statusCode < consts.StatusInternalServerError:
		return consts.StatusBadRequest
	case statusCode >= consts.StatusInternalServerError && statusCode < consts.StatusHTTPVersionNotSupported:
		return consts.StatusInternalServerError
	default:
		log.Warn().Msgf("[GetStatusCodeClass] Invalid status code: %d", statusCode)
		return -1
	}
}

// IsStatusCodeSuccess checks whether a status code is successful.
// It returns true if the status code is in the 2xx range, otherwise false.
func IsStatusCodeSuccess(statusCode int) bool {
	return GetStatusCodeClass(statusCode) == consts.StatusOK
}
