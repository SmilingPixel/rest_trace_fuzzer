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
func (p *JaegerTraceFetcher) FetchFromRemote() ([]*SimplifiedJaegerTrace, error) {
	// TODO: Implement this method @xunzhou24
	return nil, nil
}
