package feedback

import (
	"time"

	"github.com/rs/zerolog/log"
)

// SimplifiedJaegerTraceSpan represents a simplified version of a Jaeger trace.
type SimplifiedJaegerTraceSpan struct {
    TraceID       string            `json:"traceId"`       // Unique identifier for the trace
    SpanID        string            `json:"spanId"`        // Unique identifier for the span
    ParentID      string            `json:"parentId"`      // Unique identifier for the parent span
    OperationName string            `json:"operationName"` // Name of the operation
    SpanKind      SpanKindType      `json:"spanKind"`      // Kind of the span
    References    []interface{}     `json:"references"`    // References to other spans
    StartTime     time.Time         `json:"startTime"`     // Start time of the span
    Duration      string            `json:"duration"`      // Duration of the span
    Tags          []interface{}     `json:"tags"`          // Tags associated with the span
    Process       Process           `json:"process"`       // Process information
    Logs          []LogEntry        `json:"logs"`          // Log entries associated with the span
}

// Process represents the process information in a Jaeger trace.
type Process struct {
    ServiceName string        `json:"serviceName"` // Name of the service
    Tags        []interface{} `json:"tags"`        // Tags associated with the process
}

// LogEntry represents a log entry in a Jaeger trace.
type LogEntry struct {
    Timestamp time.Time       `json:"timestamp"` // Timestamp of the log entry
    Fields    []interface{}   `json:"fields"`    // Fields associated with the log entry
}

// CallInfo represents the information of a call between two services.
type CallInfo struct {
    // SourceService is the name of the source service.
    SourceService string
    // TargetService is the name of the target service.
    TargetService string
    // SourceMethod is the name of the source method.
    SourceMethod string
    // TargetMethod is the name of the target method.
    TargetMethod string
}

type SpanKindType string

// TraceManager manages traces.
type TraceManager struct {
    // maps trace IDs to spans
    TraceIDMap map[string][]*SimplifiedJaegerTraceSpan

    // maps span IDs to spans
    SpanIDMap map[string]*SimplifiedJaegerTraceSpan
}

// NewTraceManager creates a new TraceManager.
func NewTraceManager() *TraceManager {
    return &TraceManager{}
}

// QueryTraces pulls traces from the trace source(e.g., Jaeger).
func (m *TraceManager) PullTraces() ([]*SimplifiedJaegerTraceSpan, error) {
    // TODO: Implement this method @xunzhou24
    return nil, nil
}

// Process processes the given spans.
func (m *TraceManager) ProcessTraces(spans []*SimplifiedJaegerTraceSpan) error {
    // Map trace IDs to spans
    for _, span := range spans {
        m.TraceIDMap[span.TraceID] = append(m.TraceIDMap[span.TraceID], span)
        m.SpanIDMap[span.SpanID] = span
    }
    return nil
}

// GetCallInfos returns the call information (list) between services.
func (m *TraceManager) GetCallInfos() ([]*CallInfo, error) {
    res := make([]*CallInfo, 0)
    for _, span := range m.SpanIDMap {
        parentID := span.ParentID
        if parentID == "" || parentID == span.SpanID {
            continue
        }
        parentSpan, ok := m.SpanIDMap[parentID]
        if !ok {
            log.Error().Msgf("[TraceManager.getCallInfos] Parent span not found: %s", parentID)
            continue
        }
        // TODO: extract call info from spans @xunzhou24
        sourceServiceName := "sourceServiceName"
        targetServiceName := "targetServiceName"
        sourceMethodName := "sourceMethodName"
        targetMethodName := "targetMethodName"
        res = append(res, &CallInfo{
            SourceService: sourceServiceName,
            TargetService: targetServiceName,
            SourceMethod:  sourceMethodName,
            TargetMethod:  targetMethodName,
        })
    }
    return res, nil
}


