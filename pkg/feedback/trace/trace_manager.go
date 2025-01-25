package trace

import "github.com/rs/zerolog/log"

// TraceManager manages traces.
type TraceManager struct {

	// TraceFetcher fetches traces from the trace source(e.g., Jaeger).
	// It is a interface, and the implementation can be decided based on the trace backend.
	TraceFetcher TraceFetcher

	// TraceDB is the database for traces.
	// It is a interface, and the implementation can be decided based on the trace backend.
	TraceDB TraceDB
}

// NewTraceManager creates a new TraceManager.
func NewTraceManager() *TraceManager {

	// TODO: Initialize the TraceFetcher and TraceDB according config. @xunzhou24
	// By default, we use JaegerTraceFetcher and InMemoryTraceDB.
	traceFetcher := NewJaegerTraceFetcher()
	traceDB := NewInMemoryTraceDB()
	return &TraceManager{
		TraceFetcher: traceFetcher,
		TraceDB:      traceDB,
	}
}

// QueryTraces pulls traces from the trace source(e.g., Jaeger), and update local data.
func (m *TraceManager) PullTraces() error {
	// Fetch traces from the trace source.
	traces, err := m.TraceFetcher.FetchFromRemote()
	if err != nil {
		log.Err(err).Msg("Failed to fetch traces from remote")
		return err
	}
	err = m.TraceDB.Upsert(traces)
	if err != nil {
		log.Err(err).Msg("Failed to upsert traces")
		return err
	}
	return nil
}

// GetCallInfos returns the call information (list) between services.
func (m *TraceManager) GetCallInfos(trace *SimplifiedJaegerTrace) ([]*CallInfo, error) {
	res := make([]*CallInfo, 0)
	if trace == nil || len(trace.SpanMap) == 0 {
		log.Warn().Msg("No trace available")
		return res, nil
	}
	for _, span := range trace.SpanMap {
		for _, ref := range span.References {
			refMap, ok := ref.(map[string]string)
			if !ok {
				log.Warn().Msg("Invalid reference type")
				continue
			}
			if refMap["refType"] == "CHILD_OF" {
				parentSpanID := refMap["spanID"] // TODO: check here @xunzhou24
				parentSpan := trace.SpanMap[parentSpanID]
				callInfo := &CallInfo{
					SourceService:         parentSpan.Process.ServiceName,
					TargetService:         span.Process.ServiceName,
					SourceMethodTraceName: parentSpan.OperationName,
					TargetMethodTraceName: span.OperationName,
				}
				res = append(res, callInfo)
			}
		}
	}
	return res, nil
}
