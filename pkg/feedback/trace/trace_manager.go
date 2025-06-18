package trace

import (
	"resttracefuzzer/internal/config"
	"time"

	"github.com/rs/zerolog/log"
)

// TraceManager manages traces.
type TraceManager struct {

	// TraceFetcher fetches traces from the trace source(e.g., Jaeger, Tempo).
	// It is a interface, and the implementation can be decided based on the trace backend.
	TraceFetcher TraceFetcher

	// TraceDBs is the databases for traces.
	// TraceDB is a interface, and the implementation can be decided based on your needs.
	TraceDBs []TraceDB
}

// NewTraceManager creates a new TraceManager.
func NewTraceManager(
	traceDBs []TraceDB,
) *TraceManager {
	var traceFetcher TraceFetcher

	// By default, we use InMemoryTraceDB.
	if config.GlobalConfig.TraceBackendType == "Jaeger" {
		traceFetcher = NewJaegerTraceFetcher()
	} else if config.GlobalConfig.TraceBackendType == "Tempo" {
		traceFetcher = NewTempoTraceFetcher()
	} else {
		log.Error().Msgf("[NewTraceManager] Unsupported trace backend type: %s", config.GlobalConfig.TraceBackendType)
		return nil
	}
	
	return &TraceManager{
		TraceFetcher: traceFetcher,
		TraceDBs:      traceDBs,
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
	for _, traceDB := range m.TraceDBs {
		err = traceDB.BatchUpsert(traces)
		if err != nil {
			log.Err(err).Msg("[TraceManager.PullTraces] Failed to upsert traces")
			return err
		}
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

	for _, traceDB := range m.TraceDBs {
		err = traceDB.BatchUpsert(traces)
		if err != nil {
			log.Err(err).Msg("[TraceManager.PullTracesAndReturn] Failed to insert traces")
			return nil, err
		}
	}
	return traces, nil
}

// PullTraceByIDAndReturn pulls a trace by ID from the trace source(e.g., Jaeger), and return the trace.
func (m *TraceManager) PullTraceByIDAndReturn(traceID string) (*SimplifiedTrace, error) {
	// Wait a short time before fetching the trace, as the trace may not be
	// available immediately after the request.
	// TODO: a more sufficient way to wait for the trace to be available. @xunzhou24
	time.Sleep(time.Duration(config.GlobalConfig.TraceFetchWaitTime) * time.Millisecond)
	trace, err := m.TraceFetcher.FetchOneByIDFromRemote(traceID)
	if err != nil || trace == nil {
		log.Err(err).Msgf("[TraceManager.PullTraceByIDAndReturn] Failed to fetch trace from remote, traceID: %s", traceID)
		return nil, err
	}

	for _, traceDB := range m.TraceDBs {
		err = traceDB.Upsert(trace)
		if err != nil {
			log.Err(err).Msgf("[TraceManager.PullTraceByIDAndReturn] Failed to upsert trace, traceID: %s", traceID)
			return nil, err
		}
	}
	return trace, nil
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
		log.Warn().Msg("[TraceManager.convertTrace2CallInfos] Invalid trace, trace is nil or has no spans")
		return res, nil
	}
	for _, span := range trace.SpanMap {
		// Spans of kind 'internal' would be ignored, as we only care about the calls between services.
		if span.SpanKind == INTERNAL {
			continue
		}
		var parentSpan *SimplifiedTraceSpan
		if span.ParentID != "" {
			parentSpan = trace.SpanMap[span.ParentID]
		}
		// Invalid parent span, or the parent span is of kind 'internal'.
		if parentSpan == nil || parentSpan.SpanKind == INTERNAL {
			continue
		}
		// If parent span and span are from the same service, ignore the call.
		if parentSpan.ServiceName == span.ServiceName {
			continue
		}
		parentSpanID := parentSpan.SpanID

		// retrieve method trace name
		// Failure to retrieve for some reason would lead to the call being ignored.
		// At least one of the method trace names from the parent span and the span should be available.
		sourceMethodTraceName, sourceOk := parentSpan.RetrieveCalledMethod()
		targetMethodTraceName, targetOk := span.RetrieveCalledMethod()
		if (!sourceOk && !targetOk) || (sourceMethodTraceName == "" && targetMethodTraceName == "") {
			log.Warn().Msgf("[TraceManager.convertTrace2CallInfos] Failed to retrieve method trace name, parentSpanID: %s, spanID: %s", parentSpanID, span.SpanID)
			continue
		}
		var methodTraceName string
		if sourceMethodTraceName != "" {
			methodTraceName = sourceMethodTraceName
		} else {
			methodTraceName = targetMethodTraceName
		}

		callInfo := NewCallInfo(
			parentSpan.ServiceName,
			span.ServiceName,
			methodTraceName,
		)
		res = append(res, callInfo)
	}
	return res, nil
}
