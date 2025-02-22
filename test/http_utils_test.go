package test

import (
	"testing"

	"resttracefuzzer/pkg/utils"

	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/stretchr/testify/assert"
)

// TestNewHTTPClient tests the creation of a new HTTP client.
func TestNewHTTPClient(t *testing.T) {
	baseURL := "http://example.com"
	client := utils.NewHTTPClient(baseURL)
	assert.NotNil(t, client)
	assert.Equal(t, baseURL, client.BaseURL)
}

// TestPerformRequest tests performing a POST request with the HTTP client.
func TestPerformRequest(t *testing.T) {
	baseURL := "http://example.com"
	client := utils.NewHTTPClient(baseURL)

	headers := map[string]string{"Content-Type": "application/json"}
	pathParams := map[string]string{}
	queryParams := map[string]string{}
	body := map[string]string{"key": "value"}

	statusCode, headers, respBody, err := client.PerformRequest("/test", "POST", headers, pathParams, queryParams, body)
	assert.NoError(t, err)
	assert.Equal(t, consts.StatusOK, statusCode)
	assert.NotNil(t, headers)
	assert.NotNil(t, respBody)
}

// TestPerformHTTPSRequest tests performing a POST request with the HTTPS client.
func TestPerformHTTPSRequest(t *testing.T) {
	baseURL := "https://example.com"
	client := utils.NewHTTPClient(baseURL)

	headers := map[string]string{"Content-Type": "application/json"}
	pathParams := map[string]string{}
	queryParams := map[string]string{}
	body := map[string]string{"key": "value"}

	statusCode, headers, respBody, err := client.PerformRequest("/test", "POST", headers, pathParams, queryParams, body)
	assert.NoError(t, err)
	assert.Equal(t, consts.StatusOK, statusCode)
	assert.NotNil(t, headers)
	assert.NotNil(t, respBody)
}

// TestPerformRequestWithRetry tests performing a POST request with retries using the HTTP client.
func TestPerformRequestWithRetry(t *testing.T) {
	baseURL := "http://example.com"
	client := utils.NewHTTPClient(baseURL)

	headers := map[string]string{"Content-Type": "application/json"}
	pathParams := map[string]string{}
	queryParams := map[string]string{}
	body := map[string]string{"key": "value"}

	statusCode, headers, respBody, err := client.PerformRequestWithRetry("/test", "POST", headers, pathParams, queryParams, body, 3)
	assert.NoError(t, err)
	assert.Equal(t, consts.StatusOK, statusCode)
	assert.NotNil(t, headers)
	assert.NotNil(t, respBody)
}

// TestPerformGet tests performing a GET request with the HTTP client.
func TestPerformGet(t *testing.T) {
	baseURL := "http://example.com"
	client := utils.NewHTTPClient(baseURL)

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
	assert.Equal(t, consts.StatusContinue, utils.GetStatusCodeClass(consts.StatusContinue))
	assert.Equal(t, consts.StatusOK, utils.GetStatusCodeClass(consts.StatusOK))
	assert.Equal(t, consts.StatusMultipleChoices, utils.GetStatusCodeClass(consts.StatusMultipleChoices))
	assert.Equal(t, consts.StatusBadRequest, utils.GetStatusCodeClass(consts.StatusBadRequest))
	assert.Equal(t, consts.StatusInternalServerError, utils.GetStatusCodeClass(consts.StatusInternalServerError))
	assert.Equal(t, -1, utils.GetStatusCodeClass(999))
}

// TestIsStatusCodeSuccess tests if the given status code is a success status code.
func TestIsStatusCodeSuccess(t *testing.T) {
	assert.True(t, utils.IsStatusCodeSuccess(consts.StatusOK))
	assert.False(t, utils.IsStatusCodeSuccess(consts.StatusBadRequest))
}
