package trace

import (
	"time"

	"github.com/bytedance/sonic"
)

// SimplifiedJaegerTrace represents a simplified version of a Jaeger trace.
type SimplifiedJaegerTrace struct {
	TraceID string
	SpanMap map[string]*SimplifiedJaegerTraceSpan
}

// SimplifiedJaegerTraceSpan represents a simplified version of a Jaeger trace.
type SimplifiedJaegerTraceSpan struct {
	TraceID       string        `json:"traceId"`       // Unique identifier for the trace
	SpanID        string        `json:"spanId"`        // Unique identifier for the span
	ParentID      string        `json:"parentId"`      // Unique identifier for the parent span
	OperationName string        `json:"operationName"` // Name of the operation
	SpanKind      SpanKindType  `json:"spanKind"`      // Kind of the span
	References    []interface{} `json:"references"`    // References to other spans
	StartTime     time.Time     `json:"startTime"`     // Start time of the span
	Duration      string        `json:"duration"`      // Duration of the span
	Tags          []interface{} `json:"tags"`          // Tags associated with the span
	Process       Process       `json:"process"`       // Process information
	Logs          []LogEntry    `json:"logs"`          // Log entries associated with the span
	ChildrenIDs   []string      // Children spans' IDs
}

// Process represents the process information in a Jaeger trace.
type Process struct {
	ServiceName string        `json:"serviceName"` // Name of the service
	Tags        []interface{} `json:"tags"`        // Tags associated with the process
}

// LogEntry represents a log entry in a Jaeger trace.
type LogEntry struct {
	Timestamp time.Time     `json:"timestamp"` // Timestamp of the log entry
	Fields    []interface{} `json:"fields"`    // Fields associated with the log entry
}

// CallInfo represents the information of a call between two services.
type CallInfo struct {
	// SourceService is the name of the source service.
	SourceService string
	// TargetService is the name of the target service.
	TargetService string
	// SourceMethodTraceName is the name of the source method.
	SourceMethodTraceName string
	// TargetMethodTraceName is the name of the target method.
	TargetMethodTraceName string
}

// SpanKindType represents the type of a span.
type SpanKindType string

func (s SpanKindType) String() string {
	return string(s)
}

func (s SpanKindType)  MarshalJSON() ([]byte, error) {
	return sonic.Marshal(string(s))
}

func (s *SpanKindType) UnmarshalJSON(data []byte) error {
	*s = SpanKindType(string(data))
	return nil
}

