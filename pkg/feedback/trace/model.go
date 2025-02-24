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
	// StartTime is the start time of the trace.
	StartTime time.Time `json:"startTime"`
}

// SimplifiedTraceSpan represents a simplified version of a trace span model.
type SimplifiedTraceSpan struct {
	TraceID       string              `json:"traceID"`       // Unique identifier for the trace
	SpanID        string              `json:"spanID"`        // Unique identifier for the span
	ParentID      string              `json:"parentID"`      // Unique identifier for the parent span
	OperationName string              `json:"operationName"` // Name of the operation
	SpanKind      SpanKindType        `json:"spanKind"`      // Kind of the span
	References    []map[string]string `json:"references"`    // References to other spans
	StartTime     time.Time           `json:"startTime"`     // Start time of the span
	Duration      int64               `json:"duration"`      // Duration of the span, in microseconds
	Tags          []TagEntry          `json:"tags"`          // Tags associated with the span
	Process       ProcessValueEntry   `json:"process"`       // Process information
	Logs          []LogEntry          `json:"logs"`          // Log entries associated with the span
	ChildrenIDs   []string            `json:"childrenIDs"`   // Children spans' IDs
}

// ProcessValueEntry represents the process information in a trace.
type ProcessValueEntry struct {
	ServiceName string     `json:"serviceName"` // Name of the service
	Tags        []TagEntry `json:"tags"`        // Tags associated with the process
}

// LogEntry represents a log entry in a trace.
type LogEntry struct {
	Timestamp time.Time     `json:"timestamp"` // Timestamp of the log entry
	Fields    []interface{} `json:"fields"`    // Fields associated with the log entry
}

type TagEntry struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// CallInfo represents the information of a call between two services.
type CallInfo struct {
	// SourceService is the name of the source service.
	SourceService string `json:"sourceService"`
	// TargetService is the name of the target service.
	TargetService string `json:"targetService"`
	// SourceMethodTraceName is the name of the source method.
	SourceMethodTraceName string `json:"sourceMethodTraceName"`
	// TargetMethodTraceName is the name of the target method.
	TargetMethodTraceName string `json:"targetMethodTraceName"`
}

// SpanKindType represents the type of a span.
// See [OpenTelemetry specification](https://opentelemetry.io/docs/specs/otel/trace/api/#spankind) for more details.
type SpanKindType string

// SpanKindType values.
const (
    // CLIENT indicates that the span describes a request to an external service.
    CLIENT SpanKindType = "SPAN_KIND_CLIENT"
    // SERVER indicates that the span describes a request to the server.
    SERVER SpanKindType = "SPAN_KIND_SERVER"
    // PRODUCER indicates that the span describes a producer sending a message to a broker.
    PRODUCER SpanKindType = "SPAN_KIND_PRODUCER"
    // CONSUMER indicates that the span describes a consumer receiving a message from a broker.
    CONSUMER SpanKindType = "SPAN_KIND_CONSUMER"
    // INTERNAL indicates that the span describes an internal operation within an application.
    INTERNAL SpanKindType = "SPAN_KIND_INTERNAL"
    // UNSPECIFIED indicates that the span kind is unspecified.
    UNSPECIFIED SpanKindType = "SPAN_KIND_UNSPECIFIED"
)

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

// convertTraceTagValueToSpanKind converts a trace tag value to a SpanKindType.
// If the tag value is not recognized, it returns UNSPECIFIED.
func convertTraceTagValueToSpanKind(tagValue string) SpanKindType {
	switch tagValue {
	case "client":
		return CLIENT
	case "server":
		return SERVER
	case "producer":
		return PRODUCER
	case "consumer":
		return CONSUMER
	case "internal":
		return INTERNAL
	default:
		return UNSPECIFIED
	}
}

type JaegerTrace struct {
	TraceID   string                       `json:"traceID"`
	Spans     []JaegerTraceSpan            `json:"spans"`
	Processes map[string]ProcessValueEntry `json:"processes"`
}

// JaegerTraceSpan represents a span in a Jaeger trace.
type JaegerTraceSpan struct {
	TraceID       string                   `json:"traceID"`       // Unique identifier for the trace
	SpanID        string                   `json:"spanID"`        // Unique identifier for the span
	ParentID      string                   `json:"parentID"`      // Unique identifier for the parent span
	OperationName string                   `json:"operationName"` // Name of the operation
	References    []map[string]string      `json:"references"`    // References to other spans
	StartTime     int64                    `json:"startTime"`     // Start time of the span
	Duration      int64                    `json:"duration"`      // Duration of the span
	Tags          []TagEntry               `json:"tags"`          // Tags associated with the span
	Logs          []map[string]interface{} `json:"logs"`          // Log entries associated with the span
	ProcessID     string                   `json:"processID"`     // Process ID
	Warnings      interface{}              `json:"-"`             // Warnings associated with the span TODO: check here @xunzhou24
}

// ToSimplifiedTrace converts a JaegerTrace to a SimplifiedTrace.
func (j *JaegerTrace) ToSimplifiedTrace() *SimplifiedTrace {
	spanMap := make(map[string]*SimplifiedTraceSpan)
	startTime := time.Now()
	for _, span := range j.Spans {
		spanMap[span.SpanID] = span.ToSimplifiedTraceSpan(j.Processes)
		if spanMap[span.SpanID].StartTime.Before(startTime) {
			startTime = spanMap[span.SpanID].StartTime
		}
	}
	return &SimplifiedTrace{
		TraceID:   j.TraceID,
		SpanMap:   spanMap,
		StartTime: startTime,
	}
}

// ToSimplifiedTraceSpan converts a JaegerTraceSpan to a SimplifiedTraceSpan.
// For most fields, it's a direct copy. For fields like the SpanKind, they are parsed from the tags.
func (j *JaegerTraceSpan) ToSimplifiedTraceSpan(processMap map[string]ProcessValueEntry) *SimplifiedTraceSpan {
	span := &SimplifiedTraceSpan{
		TraceID:       j.TraceID,
		SpanID:        j.SpanID,
		ParentID:      j.ParentID,
		OperationName: j.OperationName,
		StartTime:     time.Unix(0, j.StartTime*int64(time.Microsecond)),
		Duration:      j.Duration,
		Process:       processMap[j.ProcessID],
	}

	// copy tags, references, and logs
	span.Tags = append(span.Tags, j.Tags...)
	span.References = append(span.References, j.References...)
	for _, log := range j.Logs {
		span.Logs = append(span.Logs, LogEntry{
			Timestamp: time.Unix(0, int64(log["timestamp"].(float64))*int64(time.Microsecond)),
			Fields:    log["fields"].([]interface{}),
		})
	}

	// parse span kind
	spanKind := UNSPECIFIED
	for _, tag := range j.Tags {
		if tag.Key == "span.kind" {
			spanKind = convertTraceTagValueToSpanKind(tag.Value.(string))
			break
		}
	}
	span.SpanKind = spanKind
	return span
}
