package trace

import (
	"time"

	"github.com/bytedance/sonic"
)

// SimplifiedTrace represents a simplified version of a trace model.
type SimplifiedTrace struct {
	// TraceID is the unique identifier for the trace.
	TraceID string `json:"traceID"`
	// Spans maps span IDs to SimplifiedTraceSpan.
	SpanMap map[string]*SimplifiedTraceSpan `json:"spanMap"`
}

// SimplifiedTraceSpan represents a simplified version of a trace span model.
type SimplifiedTraceSpan struct {
	TraceID       string        `json:"traceID"`       // Unique identifier for the trace
	SpanID        string        `json:"spanID"`        // Unique identifier for the span
	ParentID      string        `json:"parentID"`      // Unique identifier for the parent span
	OperationName string        `json:"operationName"` // Name of the operation
	SpanKind      SpanKindType  `json:"spanKind"`      // Kind of the span
	References    []map[string]string `json:"references"`    // References to other spans
	StartTime     time.Time     `json:"startTime"`     // Start time of the span
	Duration      int64        `json:"duration"`      // Duration of the span, in microseconds
	Tags          []map[string]string `json:"tags"`          // Tags associated with the span
	Process       Process       `json:"process"`       // Process information
	Logs          []LogEntry    `json:"logs"`          // Log entries associated with the span
	ChildrenIDs   []string      `json:"childrenIDs"`    // Children spans' IDs
}

// Process represents the process information in a trace.
type Process struct {
	ServiceName string        `json:"serviceName"` // Name of the service
	Tags        []interface{} `json:"tags"`        // Tags associated with the process
}

// LogEntry represents a log entry in a trace.
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

func (s SpanKindType) MarshalJSON() ([]byte, error) {
	return sonic.Marshal(string(s))
}

func (s *SpanKindType) UnmarshalJSON(data []byte) error {
	*s = SpanKindType(string(data))
	return nil
}

type JaegerTrace struct {
	TraceID string `json:"traceID"`
	Spans   []JaegerTraceSpan `json:"spans"`
}

// JaegerTraceSpan represents a span in a Jaeger trace.
type JaegerTraceSpan struct {
	TraceID       string                 `json:"traceID"`       // Unique identifier for the trace
	SpanID        string                 `json:"spanID"`        // Unique identifier for the span
	ParentID      string                 `json:"parentID"`      // Unique identifier for the parent span
	OperationName string                 `json:"operationName"` // Name of the operation
	References    []map[string]string    `json:"references"`    // References to other spans
	StartTime     int64                  `json:"startTime"`     // Start time of the span
	Duration      int64                  `json:"duration"`      // Duration of the span
	Tags          []map[string]string    `json:"tags"`          // Tags associated with the span
	Logs          []map[string]interface{} `json:"logs"`          // Log entries associated with the span
	ProcessID     string                 `json:"processID"`     // Process ID
	Warnings      interface{}            `json:"warnings"`      // Warnings associated with the span
}

// ToSimplifiedTrace converts a JaegerTrace to a SimplifiedTrace.
func (j *JaegerTrace) ToSimplifiedTrace() *SimplifiedTrace {
	spanMap := make(map[string]*SimplifiedTraceSpan)
	for _, span := range j.Spans {
		spanMap[span.SpanID] = span.ToSimplifiedTraceSpan()
	}
	return &SimplifiedTrace{
		TraceID: j.TraceID,
		SpanMap: spanMap,
	}
}

// ToSimplifiedTraceSpan converts a JaegerTraceSpan to a SimplifiedTraceSpan.
func (j *JaegerTraceSpan) ToSimplifiedTraceSpan() *SimplifiedTraceSpan {
	span := &SimplifiedTraceSpan{
		TraceID:       j.TraceID,
		SpanID:        j.SpanID,
		ParentID:      j.ParentID,
		OperationName: j.OperationName,
		StartTime:     time.Unix(0, j.StartTime*int64(time.Microsecond)),
		Duration:      j.Duration,
		Process: Process{
			ServiceName: j.ProcessID,
		},
	}

	span.Tags = append(span.Tags, j.Tags...)
	span.References = append(span.References, j.References...)
	for _, log := range j.Logs {
		span.Logs = append(span.Logs, LogEntry{
			Timestamp: time.Unix(0, int64(log["timestamp"].(float64))*int64(time.Microsecond)),
			Fields:    log["fields"].([]interface{}),
		})
	}
	return span
}
