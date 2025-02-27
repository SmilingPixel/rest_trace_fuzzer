package trace

import (
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// Time to wait before fetching the trace, as the trace may not be
	// available immediately after the request.
	TraceFetchWaitTime = 500 * time.Millisecond
)

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
	traces, err := m.TraceFetcher.FetchAllFromRemote()
	if err != nil {
		log.Err(err).Msg("[TraceManager.PullTraces] Failed to fetch traces from remote")
		return err
	}
	err = m.TraceDB.BatchUpsert(traces)
	if err != nil {
		log.Err(err).Msg("[TraceManager.PullTraces] Failed to upsert traces")
		return err
	}
	return nil
}

// PullTracesAndReturn pulls traces from the trace source(e.g., Jaeger), and return the traces.
func (m *TraceManager) PullTracesAndReturn() ([]*SimplifiedTrace, error) {
	// Fetch traces from the trace source.
	traces, err := m.TraceFetcher.FetchAllFromRemote()
	if err != nil {
		log.Err(err).Msg("[TraceManager.PullTracesAndReturn] Failed to fetch traces from remote")
		return nil, err
	}
	newTraces, err := m.TraceDB.BatchInsertAndReturn(traces)
	if err != nil {
		log.Err(err).Msg("[TraceManager.PullTracesAndReturn] Failed to insert traces")
		return nil, err
	}
	return newTraces, nil
}

// PullTraceByIDAndReturn pulls a trace by ID from the trace source(e.g., Jaeger), and return the trace.
func (m *TraceManager) PullTraceByIDAndReturn(traceID string) (*SimplifiedTrace, error) {
	// Wait a short time before fetching the trace, as the trace may not be
	// available immediately after the request.
	// TODO: a more sufficient way to wait for the trace to be available. @xunzhou24
	time.Sleep(TraceFetchWaitTime)
	trace, err := m.TraceFetcher.FetchOneByIDFromRemote(traceID)
	if err != nil || trace == nil {
		log.Err(err).Msgf("[TraceManager.PullTraceByIDAndReturn] Failed to fetch trace from remote, traceID: %s", traceID)
		return nil, err
	}
	newTrace, err := m.TraceDB.InsertAndReturn(trace)
	if err != nil {
		log.Err(err).Msgf("[TraceManager.PullTraceByIDAndReturn] Failed to insert trace, traceID: %s", traceID)
		return nil, err
	}
	return newTrace, nil
}

// BatchConvertTrace2CallInfos returns the call information (list) between services.
func (m *TraceManager) BatchConvertTrace2CallInfos(traces []*SimplifiedTrace) ([]*CallInfo, error) {
	res := make([]*CallInfo, 0)
	if len(traces) == 0 {
		log.Warn().Msg("[TraceManager.BatchConvertTrace2CallInfos] No trace available")
		return res, nil
	}
	for _, trace := range traces {
		callInfoList, err := m.convertTrace2CallInfos(trace)
		if err != nil {
			log.Err(err).Msg("[TraceManager.BatchConvertTrace2CallInfos] Failed to convert single trace to call infos")
			return nil, err
		}
		res = append(res, callInfoList...)
	}
	return res, nil
}

// convertTrace2CallInfos returns the call information (list) between services.
func (m *TraceManager) convertTrace2CallInfos(trace *SimplifiedTrace) ([]*CallInfo, error) {
	res := make([]*CallInfo, 0)
	if trace == nil || len(trace.SpanMap) == 0 {
		log.Warn().Msg("[TraceManager.convertTrace2CallInfos] Invalid trace")
		return res, nil
	}
	for _, span := range trace.SpanMap {
		// Spans of kind 'internal' would be ignored, as we only care about the calls between services.
		if span.SpanKind == INTERNAL {
			continue
		}
		var parentSpan *SimplifiedTraceSpan
		for _, ref := range span.References {
			if ref["refType"] == "CHILD_OF" {
				parentSpanID := ref["spanID"]
				parentSpan = trace.SpanMap[parentSpanID]
				break
			}
		}
		if parentSpan == nil {
			continue
		}
		parentSpanID := parentSpan.SpanID

		// retrieve method trace name
		// Failure to retrieve for some reason would lead to the call being ignored.
		sourceMethodTraceName, ok := parentSpan.RetrieveCalledMethod()
		if !ok || sourceMethodTraceName == "" {
			log.Warn().Msgf("[TraceManager.convertSingleTrace2CallInfos] Failed to retrieve source method trace name, spanID: %s", parentSpanID)
			continue
		}
		targetMethodTraceName, ok := span.RetrieveCalledMethod()
		if !ok || targetMethodTraceName == "" {
			log.Warn().Msgf("[TraceManager.convertSingleTrace2CallInfos] Failed to retrieve target method trace name, spanID: %s", span.SpanID)
			continue
		}

		callInfo := &CallInfo{
			SourceService:         parentSpan.Process.ServiceName,
			TargetService:         span.Process.ServiceName,
			SourceMethodTraceName: sourceMethodTraceName,
			TargetMethodTraceName: targetMethodTraceName,
		}
		res = append(res, callInfo)
	}
	return res, nil
}
