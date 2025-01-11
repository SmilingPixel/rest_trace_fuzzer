package feedback

import (
	"time"

	"github.com/rs/zerolog/log"
)

// SimplifiedJaegerTrace represents a simplified version of a Jaeger trace.
type SimplifiedJaegerTrace struct {
    TraceID string
    SpanMap map[string]*SimplifiedJaegerTraceSpan
}

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
    ChildrenIDs   []string                                 // Children spans' IDs
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
   Traces    []*SimplifiedJaegerTrace
}

// NewTraceManager creates a new TraceManager.
func NewTraceManager() *TraceManager {
    return &TraceManager{}
}

// QueryTraces pulls traces from the trace source(e.g., Jaeger), and update local data.
func (m *TraceManager) PullTraces() error {
    // TODO: Implement this method @xunzhou24
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
                    SourceService: parentSpan.Process.ServiceName,
                    TargetService: span.Process.ServiceName,
                    SourceMethod:  parentSpan.OperationName,
                    TargetMethod:  span.OperationName,
                }
                res = append(res, callInfo)
            }
        }
    }
    return res, nil
}


