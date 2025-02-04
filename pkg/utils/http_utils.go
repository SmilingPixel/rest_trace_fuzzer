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

// PerformRequest performs an HTTP request.
// It returns the status code, the response body in bytes, and an error if any.
// TODO: support authentication. @xunzhou24
func (c *HTTPClient) PerformRequest(path, method string, headers map[string]string, pathParams, queryParams map[string]string, body interface{}) (int, []byte, error) {
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	requestURL := c.BaseURL + path
	req.SetRequestURI(requestURL)
	req.SetHeaders(headers)
	req.SetMethod(method)

	// Set path params
	if len(queryParams) > 0 {
		queryParamsStrList := make([]string, 0)
		for k, v := range queryParams {
			queryParamsStrList = append(queryParamsStrList, k+"="+v)
		}
		req.SetQueryString(strings.Join(queryParamsStrList, "&"))
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

	log.Debug().Msgf("[HTTPClient.PerformRequest] Perform request, path: %s, method: %s, headers: %v, query params: %v, body: %v", path, method, headers, queryParams, body)
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
