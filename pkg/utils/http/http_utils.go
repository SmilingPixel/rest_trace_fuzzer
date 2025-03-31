package http

import (
	"context"
	"crypto/tls"
	"net/url"
	"strings"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/rs/zerolog/log"
)

// HTTPClient is an HTTP client.
// It has a base URL and a client based on Hertz.
type HTTPClient struct {
	// BaseURL is the base URL for the HTTP client.
    BaseURL                  string

	// HeadersToCapture are the headers that should be captured from the response.
    HeadersToCapture         []string

	// Client is the underlying Hertz client used to make HTTP requests.
    Client                   *client.Client

	// Middlewares are the middlewares used to process the request and response.
	Middlewares              []HTTPClientMiddleware
}

// NewHTTPClient creates a new HTTPClient.
// It takes a baseURL and headersToCapture, and middlewares as parameters and returns an instance of HTTPClient.
func NewHTTPClient(baseURL string, headersToCapture []string, middlewares []HTTPClientMiddleware) *HTTPClient {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	c, err := client.NewClient(
		client.WithTLSConfig(tlsConfig),
		client.WithDialer(standard.NewDialer()),
	)
	if err != nil {
		panic(err)
	}

	return &HTTPClient{
		Client:          c,
		BaseURL:        baseURL,
		HeadersToCapture: headersToCapture,
		Middlewares:     middlewares,
	}
}

// PerformRequestWithRetry performs an HTTP request with retry logic.
// It retries the request up to maxRetry times if a timeout error occurs.
// If the request fails for any other reason, it returns the error immediately.
// If all retry attempts fail due to timeout, it logs an error and returns the last error encountered.
func (c *HTTPClient) PerformRequestWithRetry(path, method string, headers map[string]string, pathParams, queryParams map[string]string, body []byte, maxRetry int) (int, map[string]string, []byte, error) {
	// If maxRetry is invalid, fallback to 1
	if maxRetry <= 0 {
		log.Warn().Msgf("[HTTPClient.PerformRequestWithRetry] Invalid max retry: %d, fallback to 1", maxRetry)
		maxRetry = 1
	}

	// Retry only when timeout
	var err error
	for i := range maxRetry {
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
// You do not have to encode the path params and query params, just pass them as a map. The function will do the encoding for you.
// It returns the status code, headers that we care about, the response body in bytes, and an error if any.
func (c *HTTPClient) PerformRequest(path, method string, headers map[string]string, pathParams, queryParams map[string]string, body []byte) (int, map[string]string, []byte, error) {
	// In case of nil values, initialize them
	if headers == nil {
		headers = make(map[string]string)
	}
	if pathParams == nil {
		pathParams = make(map[string]string)
	}
	if queryParams == nil {
		queryParams = make(map[string]string)
	}

	// Apply middlewares on request
	for _, middleware := range c.Middlewares {
		// errors are ignored here, as we do not want to stop the request if a middleware fails
		// You can see logs for errors in the middleware itself
		path, method, headers, pathParams, queryParams, body, _ = middleware.HandleRequest(path, method, headers, pathParams, queryParams, body)
	}
	
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	defer func() {
		protocol.ReleaseRequest(req)
		protocol.ReleaseResponse(resp)
	}()
	requestURL := c.BaseURL + path
	
	// Set path params
	if len(queryParams) > 0 {
		req.SetQueryString(paramDict2QueryStr(queryParams))
	}
	
	// Set path params, replacing the path params in the URL
	for k, v := range pathParams {
		requestURL = strings.ReplaceAll(requestURL, "{"+k+"}", url.PathEscape(v))
	}
	
	req.SetRequestURI(requestURL)
	req.SetHeaders(headers)
	req.SetMethod(method)
	req.SetBody(body)

	log.Debug().Msgf("[HTTPClient.PerformRequest] Perform request, URL: %s, method: %s, headers: %v, query params: %v, body: %s", requestURL, method, headers, queryParams, string(body))
	err := c.Client.Do(context.Background(), req, resp)
	if err != nil {
		log.Err(err).Msgf("[HTTPClient.PerformRequest] Failed to perform request, URL: %s, method: %s", requestURL, method)
		return 0, nil, nil, err
	}
	respBodyBytes, err := resp.BodyE()
	if err != nil {
		log.Err(err).Msgf("[HTTPClient.PerformRequest] Failed to get response body, URL: %s, method: %s", requestURL, method)
		return 0, nil, nil, err
	}
	// we do not log whole response body, for some responses may be too large
	statusCode := resp.StatusCode()
	log.Debug().Msgf("[HTTPClient.PerformRequest] Response, status code: %d, response body (64 bytes at most): %s", statusCode, string(respBodyBytes[:min(64, len(respBodyBytes))]))
	// retrieve headers that we care about
	retrievedHeaders := make(map[string]string)
	for _, headerKey := range c.HeadersToCapture {
		retrievedHeaders[headerKey] = resp.Header.Get(headerKey)
	}
	return statusCode, retrievedHeaders, respBodyBytes, nil
}

// PerformGet performs an HTTP GET request.
func (c *HTTPClient) PerformGet(path string, headers map[string]string, pathParams, queryParams map[string]string) (int, map[string]string, []byte, error) {
	return c.PerformRequest(path, "GET", headers, pathParams, queryParams, nil)
}

// paramDict2QueryStr converts a map of parameters to a query string.
// It returns the query string.
//
// For example, if the input is {"a": "1", "b": "2"}, the output is "a=1&b=2".
func paramDict2QueryStr(paramDict map[string]string) string {
	parameters := url.Values{}
	for k, v := range paramDict {
		parameters.Add(k, v)
	}
	return parameters.Encode()
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

// GetAllStatusCodeClasses returns all status code classes.
func GetAllStatusCodeClasses() []int {
	return []int{
		consts.StatusContinue,
		consts.StatusOK,
		consts.StatusMultipleChoices,
		consts.StatusBadRequest,
		consts.StatusInternalServerError,
	}
}
