package utils

import (
	"context"
	"strings"

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
// It returns the status code, the response body, and an error.
// TODO: support authentication. @xunzhou24
func (c *HTTPClient) PerformRequest(path, method string, headers map[string]string, params map[string]string, body []byte) (int, []byte, error) {
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	req.SetRequestURI(c.BaseURL)
	req.SetHeaders(headers)
	if len(params) > 0 {
		queryParams := make([]string, 0)
		for k, v := range params {
			queryParams = append(queryParams, k+"="+v)
		}
		req.SetQueryString(strings.Join(queryParams, "&"))
	}
	req.SetBody(body)

	req.SetMethod(method)
	err := c.Client.Do(context.Background(), req, resp)
	if err != nil {
		log.Err(err).Msgf("[HTTPClient.PerformRequest] Failed to perform request: %v", err)
		return 0, nil, err
	}
	bodyBytes, err := resp.BodyE()
	if err != nil {
		log.Err(err).Msgf("[HTTPClient.PerformRequest] Failed to get response body: %v", err)
		return 0, nil, err
	}
	statusCode := resp.StatusCode()
	return statusCode, bodyBytes, nil
}
