package trace

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"resttracefuzzer/internal/config"
	"resttracefuzzer/pkg/utils/http"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/rs/zerolog/log"
)

const (
	// Threshold for trace age.
	TRACE_FILTER_OUT_AGE = 3 * time.Minute

	// Maximum number of traces in a fetch request. (You can set it via query param "limit")
	MAX_TRACE_FETCH_NUM = 100
)

// TraceFetcher fetches traces from trace backend and parses them into Jaeger-style spans.
type TraceFetcher interface {
	// FetchFromPath fetches traces from a local file.
	//
	// Deprecated: Use FetchFromRemote instead.
	FetchFromPath(path string) ([]*SimplifiedTraceSpan, error)

	// FetchAllFromRemote fetches all traces from a remote source.
	FetchAllFromRemote() ([]*SimplifiedTrace, error)

	// FetchOneByIDFromRemote fetches a trace by its ID from a remote source.
	FetchOneByIDFromRemote(traceID string) (*SimplifiedTrace, error)
}

// JaegerTraceFetcher represents a fetcher for Jaeger traces.
type JaegerTraceFetcher struct {
	// FetcherClient is the HTTP client for fetching traces.
	FetcherClient *http.HTTPClient
}

// NewJaegerTraceFetcher creates a new JaegerTraceFetcher.
// See [official Jaeger API doc](https://www.jaegertracing.io/docs/2.3/apis/#query-json-over-http)
func NewJaegerTraceFetcher() *JaegerTraceFetcher {
	jaegerBackendURL := config.GlobalConfig.TraceBackendURL
	httpClient := http.NewHTTPClient(jaegerBackendURL, []string{}, http.EmptyHTTPClientMiddlewareSlice())
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

// FetchAllFromRemote fetches all Jaeger traces from remote source.
// It returns a list of traces, or an error if failed.
func (p *JaegerTraceFetcher) FetchAllFromRemote() ([]*SimplifiedTrace, error) {
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

// FetchOneByIDFromRemote fetches a Jaeger trace by its ID from remote source.
// It returns a SimplifiedTrace or an error if failed.
func (p *JaegerTraceFetcher) FetchOneByIDFromRemote(traceID string) (*SimplifiedTrace, error) {
	return p.fetchTraceByIDFromRemote(traceID)
}

// fetchAllServicesFromRemote fetches all services from remote source.
// It returns a list of service names, or an error if failed.
func (p *JaegerTraceFetcher) fetchAllServicesFromRemote() ([]string, error) {
	headers := map[string]string{}
	statusCode, _, respBytes, err := p.FetcherClient.PerformGet("/api/services", headers, nil, nil)
	if err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchAllServicesFromRemote] Failed to fetch services")
		return nil, err
	}
	if http.GetStatusCodeClass(statusCode) != consts.StatusOK {
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
	statusCode, _, respBytes, err := p.FetcherClient.PerformGet(path, headers, nil, queryParams)
	if err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchServiceTracesFromRemote] Failed to fetch traces, path: %s, query params: %v", path, queryParams)
		return nil, err
	}
	if http.GetStatusCodeClass(statusCode) != consts.StatusOK {
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

// fetchTraceByIDFromRemote fetches a Jaeger trace by its ID from a remote source.
// It returns a SimplifiedTrace or an error if the fetch operation fails.
func (p *JaegerTraceFetcher) fetchTraceByIDFromRemote(traceID string) (*SimplifiedTrace, error) {
	path := fmt.Sprintf("/api/traces/%s", traceID)
	headers := map[string]string{}
	statusCode, _, respBytes, err := p.FetcherClient.PerformGet(path, headers, nil, nil)
	if err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchTraceByIDFromRemote] Failed to fetch trace, path: %s", path)
		return nil, err
	}
	if http.GetStatusCodeClass(statusCode) != consts.StatusOK {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchTraceByIDFromRemote] Failed to fetch trace, statusCode: %d, path: %s", statusCode, path)
		return nil, err
	}

	var jaegerTraceResp struct {
		Data []JaegerTrace `json:"data"`
	}
	if err := sonic.Unmarshal(respBytes, &jaegerTraceResp); err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchTraceByIDFromRemote] Failed to unmarshal Jaeger trace response")
		return nil, err
	}
	if len(jaegerTraceResp.Data) == 0 {
		err := fmt.Errorf("trace not found: %s", traceID)
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchTraceByIDFromRemote] Failed to fetch trace")
		return nil, err
	}
	return jaegerTraceResp.Data[0].ToSimplifiedTrace(), nil
}


// TempoTraceFetcher represents a fetcher for Tempo traces.
type TempoTraceFetcher struct {
	// FetcherClient is the HTTP client for fetching traces.
	FetcherClient *http.HTTPClient
}

// NewTempoTraceFetcher creates a new TempoTraceFetcher.
// See [official Tempo API doc](https://grafana.com/docs/tempo/latest/api_docs/)
func NewTempoTraceFetcher() *TempoTraceFetcher {
	tempoBackendURL := config.GlobalConfig.TraceBackendURL
	httpClient := http.NewHTTPClient(tempoBackendURL, []string{}, http.EmptyHTTPClientMiddlewareSlice())
	return &TempoTraceFetcher{
		FetcherClient: httpClient,
	}
}

// FetchFromPath fetches Tempo traces from given path.
// The method is not implemented, and will not be, as the interface marks the method as deprecated.
func (p *TempoTraceFetcher) FetchFromPath(filePath string) ([]*SimplifiedTraceSpan, error) {
	return nil, fmt.Errorf("TempoTraceFetcher.FetchFromPath is not implemented")
}

// FetchAllFromRemote fetches all Tempo traces from remote source.
// It returns a list of traces, or an error if failed.
func (p *TempoTraceFetcher) FetchAllFromRemote() ([]*SimplifiedTrace, error) {
	// TODO: Implement this method @xunzhou24
	return nil, fmt.Errorf("TempoTraceFetcher.FetchAllFromRemote is not implemented")
}

// FetchOneByIDFromRemote fetches a Tempo trace by its ID from remote source.
// It returns a SimplifiedTrace or an error if failed.
func (p *TempoTraceFetcher) FetchOneByIDFromRemote(traceID string) (*SimplifiedTrace, error) {
	path := fmt.Sprintf("/api/v2/traces/%s", traceID)
	headers := map[string]string{}
	statusCode, _, respBytes, err := p.FetcherClient.PerformGet(path, headers, nil, nil)
	if err != nil {
		log.Err(err).Msgf("[TempoTraceFetcher.FetchOneByIDFromRemote] Failed to fetch trace, path: %s", path)
		return nil, err
	}
	if http.GetStatusCodeClass(statusCode) != consts.StatusOK {
		log.Err(err).Msgf("[TempoTraceFetcher.FetchOneByIDFromRemote] Failed to fetch trace, statusCode: %d, path: %s", statusCode, path)
		return nil, err
	}

	var tempoTraceResp struct {
		Data []TempoTrace `json:"data"`
	}
	if err := sonic.Unmarshal(respBytes, &tempoTraceResp); err != nil {
		log.Err(err).Msgf("[TempoTraceFetcher.FetchOneByIDFromRemote] Failed to unmarshal Tempo trace response")
		return nil, err
	}
	if len(tempoTraceResp.Data) == 0 {
		err := fmt.Errorf("trace not found: %s", traceID)
		log.Err(err).Msgf("[TempoTraceFetcher.FetchOneByIDFromRemote] Failed to fetch trace")
		return nil, err
	}
	return tempoTraceResp.Data[0].ToSimplifiedTrace(), nil
}
