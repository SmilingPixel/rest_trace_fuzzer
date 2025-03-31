package test

import (
	"testing"

	"resttracefuzzer/pkg/utils/http"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/stretchr/testify/assert"
)

const (
	TRACE_ID_HEADER_KEY = "X-Trace-Id"
)

// TestNewHTTPClient tests the creation of a new HTTP client.
func TestNewHTTPClient(t *testing.T) {
	baseURL := "http://example.com"
	client := http.NewHTTPClient(baseURL, []string{TRACE_ID_HEADER_KEY}, http.EmptyHTTPClientMiddlewareSlice())
	assert.NotNil(t, client)
	assert.Equal(t, baseURL, client.BaseURL)
}

// TestPerformRequest tests performing a POST request with the HTTP client.
func TestPerformRequest(t *testing.T) {
	baseURL := "http://example.com"
	client := http.NewHTTPClient(baseURL, []string{TRACE_ID_HEADER_KEY}, http.EmptyHTTPClientMiddlewareSlice())

	headers := map[string]string{"Content-Type": "application/json"}
	pathParams := map[string]string{}
	queryParams := map[string]string{}
	body := map[string]string{"key": "value"}

	bodyBytes, err := sonic.Marshal(body)
	assert.NoError(t, err)

	statusCode, headers, respBody, err := client.PerformRequest("/test", "POST", headers, pathParams, queryParams, bodyBytes)
	assert.NoError(t, err)
	assert.Equal(t, consts.StatusOK, statusCode)
	assert.NotNil(t, headers)
	assert.NotNil(t, respBody)
}

// TestPerformHTTPSRequest tests performing a POST request with the HTTPS client.
func TestPerformHTTPSRequest(t *testing.T) {
	baseURL := "https://example.com"
	client := http.NewHTTPClient(baseURL, []string{TRACE_ID_HEADER_KEY}, http.EmptyHTTPClientMiddlewareSlice())

	headers := map[string]string{"Content-Type": "application/json"}
	pathParams := map[string]string{}
	queryParams := map[string]string{}
	body := map[string]string{"key": "value"}

	
	bodyBytes, err := sonic.Marshal(body)
	assert.NoError(t, err)

	statusCode, headers, respBody, err := client.PerformRequest("/test", "POST", headers, pathParams, queryParams, bodyBytes)
	assert.NoError(t, err)
	assert.Equal(t, consts.StatusOK, statusCode)
	assert.NotNil(t, headers)
	assert.NotNil(t, respBody)
}

// TestPerformRequestWithRetry tests performing a POST request with retries using the HTTP client.
func TestPerformRequestWithRetry(t *testing.T) {
	baseURL := "http://example.com"
	client := http.NewHTTPClient(baseURL, []string{TRACE_ID_HEADER_KEY}, http.EmptyHTTPClientMiddlewareSlice())

	headers := map[string]string{"Content-Type": "application/json"}
	pathParams := map[string]string{}
	queryParams := map[string]string{}
	body := map[string]string{"key": "value"}

	bodyBytes, err := sonic.Marshal(body)
	assert.NoError(t, err)

	statusCode, headers, respBody, err := client.PerformRequestWithRetry("/test", "POST", headers, pathParams, queryParams, bodyBytes, 3)
	assert.NoError(t, err)
	assert.Equal(t, consts.StatusOK, statusCode)
	assert.NotNil(t, headers)
	assert.NotNil(t, respBody)
}

// TestPerformGet tests performing a GET request with the HTTP client.
func TestPerformGet(t *testing.T) {
	baseURL := "http://example.com"
	client := http.NewHTTPClient(baseURL, []string{TRACE_ID_HEADER_KEY}, http.EmptyHTTPClientMiddlewareSlice())

	headers := map[string]string{}
	pathParams := map[string]string{}
	queryParams := map[string]string{}

	statusCode, headers, respBody, err := client.PerformGet("/test", headers, pathParams, queryParams)
	assert.NoError(t, err)
	assert.Equal(t, consts.StatusOK, statusCode)
	assert.NotNil(t, headers)
	assert.NotNil(t, respBody)
}

// TestGetStatusCodeClass tests the classification of HTTP status codes.
func TestGetStatusCodeClass(t *testing.T) {
	assert.Equal(t, consts.StatusContinue, http.GetStatusCodeClass(consts.StatusContinue))
	assert.Equal(t, consts.StatusOK, http.GetStatusCodeClass(consts.StatusOK))
	assert.Equal(t, consts.StatusMultipleChoices, http.GetStatusCodeClass(consts.StatusMultipleChoices))
	assert.Equal(t, consts.StatusBadRequest, http.GetStatusCodeClass(consts.StatusBadRequest))
	assert.Equal(t, consts.StatusInternalServerError, http.GetStatusCodeClass(consts.StatusInternalServerError))
	assert.Equal(t, -1, http.GetStatusCodeClass(999))
}

// TestIsStatusCodeSuccess tests if the given status code is a success status code.
func TestIsStatusCodeSuccess(t *testing.T) {
	assert.True(t, http.IsStatusCodeSuccess(consts.StatusOK))
	assert.False(t, http.IsStatusCodeSuccess(consts.StatusBadRequest))
}
