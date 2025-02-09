package trace

import (
	"io"
	"os"
	"strconv"
	"time"

	"resttracefuzzer/internal/config"
	"resttracefuzzer/pkg/utils"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

const (
	// Threshold for trace age in seconds.
	TRACE_FILTER_OUT_AGE = 3 * time.Minute

	// Maximum number of traces in a fetch request. (You can set it via query param "limit")
	MAX_TRACE_FETCH_NUM = 100
)

// TraceFetcher fetches traces from trace backend and parses them into Jaeger-style spans.
type TraceFetcher interface {
	// FetchFromPath fetches traces from a local file.
	FetchFromPath(path string) ([]*SimplifiedTraceSpan, error)

	// FetchFromRemote fetches traces from a remote source.
	FetchFromRemote() ([]*SimplifiedTrace, error)
}

// JaegerTraceFetcher represents a fetcher for Jaeger traces.
type JaegerTraceFetcher struct {
	// FetcherClient is the HTTP client for fetching traces.
	FetcherClient *utils.HTTPClient
}

// NewJaegerTraceFetcher creates a new JaegerTraceFetcher.
func NewJaegerTraceFetcher() *JaegerTraceFetcher {
	jaegerBackendURL := config.GlobalConfig.TraceBackendURL
	httpClient := utils.NewHTTPClient(jaegerBackendURL)
	return &JaegerTraceFetcher{
		FetcherClient: httpClient,
	}
}

// FetchFromPath fetches Jaeger traces from given path.
//
// Deprecated: Use FetchFromRemote instead.
func (p *JaegerTraceFetcher) FetchFromPath(filePath string) ([]*SimplifiedTraceSpan, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchFromPath] Failed to open file: %s", filePath)
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchFromPath] Failed to read file: %s", filePath)
		return nil, err
	}

	var result struct {
		Spans []*SimplifiedTraceSpan `json:"spans"`
	}
	if err := sonic.Unmarshal(bytes, &result); err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchFromPath] Failed to unmarshal file: %s", filePath)
		return nil, err
	}

	return result.Spans, nil
}

// FetchFromRemote fetches Jaeger traces from remote source.
// It returns a list of traces, or an error if failed.
func (p *JaegerTraceFetcher) FetchFromRemote() ([]*SimplifiedTrace, error) {
	serviceNames, err := p.fetchAllServicesFromRemote()
	if err != nil {
		log.Err(err).Msg("[JaegerTraceFetcher.FetchFromRemote] Failed to fetch services")
		return nil, err
	}
	if len(serviceNames) == 0 {
		log.Warn().Msg("[JaegerTraceFetcher.FetchFromRemote] No services found")
		return nil, nil
	}
	traces := make([]*SimplifiedTrace, 0)
	for _, serviceName := range serviceNames {
		serviceTraces, err := p.fetchServiceTracesFromRemote(serviceName)
		if err != nil {
			log.Err(err).Msg("[JaegerTraceFetcher.FetchFromRemote] Failed to fetch traces")
			return nil, err
		}
		// Filter out empty and too old traces
		currentTime := time.Now()
		for _, trace := range serviceTraces {
			if trace == nil || currentTime.Sub(trace.StartTime) > TRACE_FILTER_OUT_AGE {
				continue
			}
			traces = append(traces, trace)
		}
	}
	return traces, nil
}

// fetchAllServicesFromRemote fetches all services from remote source.
// It returns a list of service names, or an error if failed.
func (p *JaegerTraceFetcher) fetchAllServicesFromRemote() ([]string, error) {
	headers := map[string]string{}
	statusCode, respBytes, err := p.FetcherClient.PerformGet("/api/services", headers, nil, nil)
	if err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchAllServicesFromRemote] Failed to fetch services")
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchAllServicesFromRemote] Failed to fetch services, statusCode: %d", statusCode)
		return nil, err
	}
	var serviceNamesResp struct {
		Data []string `json:"data"`
	}
	if err := sonic.Unmarshal(respBytes, &serviceNamesResp); err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchAllServicesFromRemote] Failed to unmarshal services")
		return nil, err
	}
	return serviceNamesResp.Data, nil
}

// fetchServiceTracesFromRemote fetches traces of a service from remote source.
// It returns a list of traces, or an error if failed.
func (p *JaegerTraceFetcher) fetchServiceTracesFromRemote(serviceName string) ([]*SimplifiedTrace, error) {
	path := "/api/traces"
	headers := map[string]string{}
	queryParams := map[string]string{
		"limit":   strconv.Itoa(MAX_TRACE_FETCH_NUM),
		"service": serviceName,
	}
	statusCode, respBytes, err := p.FetcherClient.PerformGet(path, headers, nil, queryParams)
	if err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchServiceTracesFromRemote] Failed to fetch traces, path: %s, query params: %v", path, queryParams)
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchServiceTracesFromRemote] Failed to fetch traces, statusCode: %d, path: %s, query params: %v", statusCode, path, queryParams)
		return nil, err
	}

	var jaegerTraceListResp struct {
		Data []JaegerTrace `json:"data"`
	}
	if err := sonic.Unmarshal(respBytes, &jaegerTraceListResp); err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchServiceTracesFromRemote] Failed to unmarshal Jaeger traces response")
		return nil, err
	}
	traces := make([]*SimplifiedTrace, 0)
	for _, jaegerTrace := range jaegerTraceListResp.Data {
		traces = append(traces, jaegerTrace.ToSimplifiedTrace())
	}
	return traces, nil
}
