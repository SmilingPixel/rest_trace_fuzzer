package utils

import (
	"context"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/rs/zerolog/log"
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
func (c *HTTPClient) PerformRequestWithRetry(path, method string, headers map[string]string, pathParams, queryParams map[string]string, body interface{}, maxRetry int) (int, []byte, error) {
	// If maxRetry is invalid, fallback to 1
	if maxRetry <= 0 {
		log.Warn().Msgf("[HTTPClient.PerformRequestWithRetry] Invalid max retry: %d, fallback to 1", maxRetry)
		return c.PerformRequest(path, method, headers, pathParams, queryParams, body)
	}

	// Retry only when timeout
	var err error
	for i := 0; i < maxRetry; i++ {
		statusCode, respBodyBytes, err := c.PerformRequest(path, method, headers, pathParams, queryParams, body)
		if err != nil {
			if strings.Contains(string(err.Error()), "timeout") {
				log.Warn().Msgf("[HTTPClient.PerformRequestWithRetry] Retry %d times due to timeout, URL: %s, method: %s", i+1, c.BaseURL+path, method)
				continue
			} else {
				return statusCode, respBodyBytes, err
			}
		}
		return statusCode, respBodyBytes, nil
	}
	log.Err(err).Msgf("[HTTPClient.PerformRequestWithRetry] Retry %d times but still timeout, URL: %s, method: %s", maxRetry, c.BaseURL+path, method)
	return 0, nil, err
}

// PerformRequest performs an HTTP request.
// It returns the status code, the response body in bytes, and an error if any.
func (c *HTTPClient) PerformRequest(path, method string, headers map[string]string, pathParams, queryParams map[string]string, body interface{}) (int, []byte, error) {
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
		return 0, nil, err
	}
	req.SetBody(bodyBytes)

	log.Debug().Msgf("[HTTPClient.PerformRequest] Perform request, URL: %s, method: %s, headers: %v, query params: %v, body: %v", requestURL, method, headers, queryParams, body)
	err = c.Client.Do(context.Background(), req, resp)
	if err != nil {
		log.Err(err).Msgf("[HTTPClient.PerformRequest] Failed to perform request, URL: %s, method: %s", requestURL, method)
		return 0, nil, err
	}
	respBodyBytes, err := resp.BodyE()
	if err != nil {
		log.Err(err).Msgf("[HTTPClient.PerformRequest] Failed to get response body, URL: %s, method: %s", requestURL, method)
		return 0, nil, err
	}
	statusCode := resp.StatusCode()
	log.Debug().Msgf("[HTTPClient.PerformRequest] Response, status code: %d", statusCode) // we do not log response body, for some responses may be too large
	return statusCode, respBodyBytes, nil
}

// PerformGet performs an HTTP GET request.
func (c *HTTPClient) PerformGet(path string, headers map[string]string, pathParams, queryParams map[string]string) (int, []byte, error) {
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
