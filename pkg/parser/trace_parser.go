package parser

import (
	"io"
	"os"
	"resttracefuzzer/pkg/feedback"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

// TraceParser represents a parser for traces.
type TraceParser interface {
	ParseFromPath(path string) (*feedback.SimplifiedJaegerTraceSpan, error)
}

// JaegerTraceParser represents a parser for Jaeger traces.
type JaegerTraceParser struct {
}

// NewJaegerTraceParser creates a new JaegerTraceParser.
func NewJaegerTraceParser() *JaegerTraceParser {
	return &JaegerTraceParser{}
}

// ParseFromPath parses a Jaeger trace from a given path.
func (p *JaegerTraceParser) ParseFromPath(filePath string) ([]*feedback.SimplifiedJaegerTraceSpan, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error().Err(err).Msgf("[JaegerTraceParser.ParseFromPath]Failed to open file: %s", filePath)
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Error().Err(err).Msgf("[JaegerTraceParser.ParseFromPath]Failed to read file: %s", filePath)
		return nil, err
	}

	var result struct {
		Spans []*feedback.SimplifiedJaegerTraceSpan `json:"spans"`
	}
	if err := sonic.Unmarshal(bytes, &result); err != nil {
		log.Error().Err(err).Msgf("[JaegerTraceParser.ParseFromPath]Failed to unmarshal file: %s", filePath)
		return nil, err
	}

	return result.Spans, nil
}
