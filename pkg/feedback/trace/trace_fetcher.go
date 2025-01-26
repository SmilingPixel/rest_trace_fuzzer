package trace

import (
	"io"
	"os"

	"resttracefuzzer/internal/config"
	"resttracefuzzer/pkg/utils"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

// TraceFetcher fetches traces from trace banckend and parses them into Jaeger-style spans.
type TraceFetcher interface {
	// FetchFromPath fetches traces from a local file.
	FetchFromPath(path string) ([]*SimplifiedJaegerTraceSpan, error)

	// FetchFromRemote fetches traces from a remote source.
	FetchFromRemote() ([]*SimplifiedJaegerTrace, error)
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
func (p *JaegerTraceFetcher) FetchFromPath(filePath string) ([]*SimplifiedJaegerTraceSpan, error) {
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
		Spans []*SimplifiedJaegerTraceSpan `json:"spans"`
	}
	if err := sonic.Unmarshal(bytes, &result); err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchFromPath] Failed to unmarshal file: %s", filePath)
		return nil, err
	}

	return result.Spans, nil
}

// FetchFromRemote fetches Jaeger traces from remote source.
// It returns a list of traces, or an error if failed.
func (p *JaegerTraceFetcher) FetchFromRemote() ([]*SimplifiedJaegerTrace, error) {
	serviceNames, err := p.fetchAllServicesFromRemote()
	if err != nil {
		log.Err(err).Msg("[JaegerTraceFetcher.FetchFromRemote] Failed to fetch services")
		return nil, err
	}
	if len(serviceNames) == 0 {
		log.Warn().Msg("[JaegerTraceFetcher.FetchFromRemote] No services found")
		return nil, nil
	}
	traces := make([]*SimplifiedJaegerTrace, 0)
	for _, serviceName := range serviceNames {
		serviceTraces, err := p.fetchServiceTracesFromRemote(serviceName)
		if err != nil {
			log.Err(err).Msg("[JaegerTraceFetcher.FetchFromRemote] Failed to fetch traces")
			return nil, err
		}
		traces = append(traces, serviceTraces...)
	}
	return traces, nil
}

// fetchAllServicesFromRemote fetches all services from remote source.
// It returns a list of service names, or an error if failed.
func (p *JaegerTraceFetcher) fetchAllServicesFromRemote() ([]string, error) {
	headers := map[string]string{}
	params := map[string]string{}
	statusCode, respBytes, err := p.FetcherClient.PerformGet("/api/services", headers, params)
	if err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchAllServicesFromRemote] Failed to fetch services: %v", err)
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchAllServicesFromRemote] Failed to fetch services: %d", statusCode)
		return nil, err
	}
	var serviceNames []string
	if err := sonic.Unmarshal(respBytes, &serviceNames); err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchAllServicesFromRemote] Failed to unmarshal services: %v", err)
		return nil, err
	}
	return serviceNames, nil
}

// fetchServiceTracesFromRemote fetches traces of a service from remote source.
// It returns a list of traces, or an error if failed.
func (p *JaegerTraceFetcher) fetchServiceTracesFromRemote(serviceName string) ([]*SimplifiedJaegerTrace, error) {
	url := "/api/traces" + "?limit=2000service=" + serviceName
	headers := map[string]string{}
	params := map[string]string{}
	statusCode, respBytes, err := p.FetcherClient.PerformGet(url, headers, params)
	if err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchServiceTracesFromRemote] Failed to fetch traces: %v", err)
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchServiceTracesFromRemote] Failed to fetch traces: %d", statusCode)
		return nil, err
	}
	var traces []*SimplifiedJaegerTrace
	if err := sonic.Unmarshal(respBytes, &traces); err != nil {
		log.Err(err).Msgf("[JaegerTraceFetcher.FetchServiceTracesFromRemote] Failed to unmarshal traces: %v", err)
		return nil, err
	}
	return traces, nil
}
