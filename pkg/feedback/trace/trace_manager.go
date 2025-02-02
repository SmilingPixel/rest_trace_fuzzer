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

// PullTraces pulls traces from the trace source(e.g., Jaeger), and update local data.
func (m *TraceManager) PullTraces() error {
	// Fetch traces from the trace source.
	traces, err := m.TraceFetcher.FetchFromRemote()
	if err != nil {
		log.Err(err).Msg("Failed to fetch traces from remote")
		return err
	}
	err = m.TraceDB.BatchUpsert(traces)
	if err != nil {
		log.Err(err).Msg("Failed to upsert traces")
		return err
	}
	return nil
}

// PullTracesAndReturn pulls traces from the trace source(e.g., Jaeger), and return the traces.
func (m *TraceManager) PullTracesAndReturn() ([]*SimplifiedTrace, error) {
	// Fetch traces from the trace source.
	traces, err := m.TraceFetcher.FetchFromRemote()
	if err != nil {
		log.Err(err).Msg("Failed to fetch traces from remote")
		return nil, err
	}
	newTraces, err := m.TraceDB.BatchInsertAndReturn(traces)
	if err != nil {
		log.Err(err).Msg("Failed to insert traces")
		return nil, err
	}
	return newTraces, nil
}

// ConvertTraces2CallInfos returns the call information (list) between services.
func (m *TraceManager) ConvertTraces2CallInfos(traces []*SimplifiedTrace) ([]*CallInfo, error) {
	res := make([]*CallInfo, 0)
	if len(traces) == 0 {
		log.Warn().Msg("[TraceManager.ConvertTraces2CallInfos] No trace available")
		return res, nil
	}
	for _, trace := range traces {
		callInfoList, err := m.convertSingleTrace2CallInfos(trace)
		if err != nil {
			log.Err(err).Msg("[TraceManager.ConvertTraces2CallInfos] Failed to convert single trace to call infos")
			return nil, err
		}
		res = append(res, callInfoList...)
	}
	return res, nil
}

// convertSingleTrace2CallInfos returns the call information (list) between services.
func (m *TraceManager) convertSingleTrace2CallInfos(trace *SimplifiedTrace) ([]*CallInfo, error) {
	res := make([]*CallInfo, 0)
	if trace == nil || len(trace.SpanMap) == 0 {
		log.Warn().Msg("[TraceManager.convertSingleTrace2CallInfos] Invalid trace")
		return res, nil
	}
	for _, span := range trace.SpanMap {
		for _, ref := range span.References {
			if ref["refType"] == "CHILD_OF" {
				parentSpanID := ref["spanID"] // TODO: check here @xunzhou24
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
