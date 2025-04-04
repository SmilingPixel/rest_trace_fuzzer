package trace

import (
	"regexp"
	"resttracefuzzer/pkg/utils"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
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
	SemanticConvention SemanticConventionType `json:"semanticConvention"` // Semantic convention of the span
	References    []map[string]string `json:"references"`    // References to other spans
	StartTime     time.Time           `json:"startTime"`     // Start time of the span
	Duration      int64               `json:"duration"`      // Duration of the span, in microseconds
	TagMap        map[string]TagEntry `json:"tagMap"`       // Tags associated with the span, map from tag key to tag entry
	Process       *ProcessValueEntry  `json:"process"`       // Process information
	Logs          []LogEntry          `json:"logs"`          // Log entries associated with the span
	ChildrenIDs   []string            `json:"childrenIDs"`  // Children spans' IDs
}

// ProcessValueEntry represents the process information in a trace.
type ProcessValueEntry struct {
	ServiceName string     `json:"serviceName"` // Name of the service
	Tags        []TagEntry `json:"tags"`         // Tags associated with the process
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
	// Method is the called method.
	Method string `json:"method"`
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

// SemanticConventionType represents the type of a semantic convention.
// See [OpenTelemetry specification](https://opentelemetry.io/docs/specs/semconv/) for more details.
type SemanticConventionType string

const (
	// SemanticConventionTypeHTTP represents the HTTP semantic convention.
	SemanticConventionTypeHTTP SemanticConventionType = "SEMANTIC_CONVENTION_HTTP"

	// SemanticConventionTypeRPC represents the RPC semantic convention.
	SemanticConventionTypeRPC SemanticConventionType = "SEMANTIC_CONVENTION_RPC"

	// SemanticConventionTypeMessaging represents the messaging system semantic convention.
	SemanticConventionTypeMessaging SemanticConventionType = "SEMANTIC_CONVENTION_MESSAGING"

	// TODO: support more semantic conventions @xunzhou24

	// SemanticConventionTypeUnknown represents an unknown semantic convention.
	SemanticConventionTypeUnknown SemanticConventionType = "SEMANTIC_CONVENTION_UNKNOWN"
)

func (s SemanticConventionType) String() string {
	return string(s)
}

func (s SemanticConventionType) MarshalJSON() ([]byte, error) {
	return sonic.Marshal(string(s))
}

func (s *SemanticConventionType) UnmarshalJSON(data []byte) error {
	*s = SemanticConventionType(string(data))
	return nil
}

// RetrieveCalledMethod retrieves the called method that the span represents.
// For example, if service A called method 'f' of service B, this method returns 'f'.
// The second returned value indicates whether the method exists (or can be found).
func (s *SimplifiedTraceSpan) RetrieveCalledMethod() (string, bool) {
	method, ok := "", false

	// For now, we only consider invocation between services.
	// So internal spans are ignored.
	if s.SpanKind == INTERNAL {
		return method, ok
	}

	// For different semantic conventions, we have different ways to retrieve the called method.
	switch s.SemanticConvention {

	// For HTTP, the span name is '{method} {target}' or '{method}'， and we can directly get the target.
	// The target is http.route or url.template.
	// See [OpenTelemetry specification](https://opentelemetry.io/docs/specs/semconv/http/http-spans/) for more details.
	// Note that HTTP method (GET, POST, etc.) is different from the method name we want (e.g., 'f').
	//
	// For some frameworks, gRPC call would probably be recorded as HTTP spans when gRPC is over HTTP/2.
	// In this case, we notice that the operation name is in the format of 'POST {RPC Operation Name}'. For example, 'POST /grpc.test.EchoService/Echo'.
	// We split the operation name by '/' and get the last part.
	case SemanticConventionTypeHTTP:
		operationNameParts := strings.Split(s.OperationName, " ")
		var targetName string
		if len(operationNameParts) >= 2 {
			targetName = operationNameParts[1]
		}
		// gRPC over HTTP/2
		if grpcMethod, exist := s.TagMap["grpc.method"]; exist {
			targetName = utils.ExtractLastSegment(grpcMethod.Value.(string), []string{"/"})
		}
		return targetName, (targetName != "")

	// For RPC, the span name is '$package.$service/$method' (e.g., 'grpc.test.EchoService/Echo').
	// We split the operation name by '/' and get the last part.
	// See [OpenTelemetry specification](https://opentelemetry.io/docs/specs/semconv/rpc/rpc-spans/) for more details.
	case SemanticConventionTypeRPC:
		operationNameParts := strings.Split(s.OperationName, "/")
		if len(operationNameParts) < 2 {
			return "", false
		}
		return operationNameParts[len(operationNameParts)-1], true

	// TODO: support more semantic conventions @xunzhou24
	default:
		log.Warn().Msgf("[SimplifiedTraceSpan.RetrieveCalledMethod] Unsupported semantic convention: %s", s.SemanticConvention)
		return "", false
	}
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
	Processes map[string]*ProcessValueEntry `json:"processes"`
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
func (j *JaegerTraceSpan) ToSimplifiedTraceSpan(processMap map[string]*ProcessValueEntry) *SimplifiedTraceSpan {
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
	span.TagMap = make(map[string]TagEntry)
	for _, tag := range j.Tags {
		span.TagMap[tag.Key] = tag
	}
	span.References = append(span.References, j.References...)
	for _, log := range j.Logs {
		span.Logs = append(span.Logs, LogEntry{
			Timestamp: time.Unix(0, int64(log["timestamp"].(float64))*int64(time.Microsecond)),
			Fields:    log["fields"].([]interface{}),
		})
	}

	// parse span kind
	spanKind := UNSPECIFIED
	if spandKindValue, exist := span.TagMap["span.kind"]; exist {
		spanKind = convertTraceTagValueToSpanKind(spandKindValue.Value.(string))
	}
	span.SpanKind = spanKind

	// parse semantic convention
	span.SemanticConvention = j.InferSemanticConvention()

	return span
}

// InferSemanticConvention infers the semantic convention of a JaegerTraceSpan.
// Unsupported semantic conventions are returned as SemanticConventionTypeUnknown.
// Note: the result may not be accurate, as it's based on the tags and name format.
// We use required tags and name format that are specific to the semantic conventions to determine the semantic convention.
// See [OpenTelemetry specification](https://opentelemetry.io/docs/specs/semconv/) for more details.
// TODO: support more semantic conventions @xunzhou24
func (j *JaegerTraceSpan) InferSemanticConvention() SemanticConventionType {
	// For HTTP, the span name is '{method} {target}' or '{method}'
	httpOperationNameRegex := `^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS|TRACE)(?:\s+(\S+))?$`
	if matched, _ := regexp.MatchString(httpOperationNameRegex, j.OperationName); matched {
		return SemanticConventionTypeHTTP
	}

	// For RPC and Messaging system, tags are used to determine the semantic convention.
	for _, tag := range j.Tags {
		if tag.Key == "rpc.system" {
			return SemanticConventionTypeRPC
		}
		if tag.Key == "messaging.system" {
			return SemanticConventionTypeMessaging
		}
	}

	return SemanticConventionTypeUnknown
}