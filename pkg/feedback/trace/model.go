package trace

import (
	"regexp"
	"resttracefuzzer/pkg/utils"
	"strconv"
	"strings"
	"time"

	"maps"

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
	StartTime     time.Time           `json:"startTime"`     // Start time of the span
	Duration      int64               `json:"duration"`      // Duration of the span, in microseconds
	AttributeMap        map[string]AttributeEntry `json:"attributeMap"`       // Attributes associated with the span, map from tag key to attribute entry
	ServiceName   string              `json:"serviceName"`   // Name of the service
}

type AttributeEntry struct {
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

// NewCallInfo creates a new CallInfo instance.
// The service would be formatted, as there may be differences in service names between those in trace and doc.
func NewCallInfo(sourceService, targetService, method string) *CallInfo {
	return &CallInfo{
		SourceService: utils.FormatServiceName(sourceService),
		TargetService: utils.FormatServiceName(targetService),
		Method:        method,
	}
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

	// SemanticConventionTypeDatabase represents the database semantic convention.
	SemanticConventionTypeDatabase SemanticConventionType = "SEMANTIC_CONVENTION_DATABASE"

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
		if grpcMethod, exist := s.AttributeMap["grpc.method"]; exist {
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

// convertJaegerTraceTagValueToSpanKind converts a Jaeger trace tag value to a SpanKindType.
// If the tag value is not recognized, it returns UNSPECIFIED.
func convertJaegerTraceTagValueToSpanKind(tagValue string) SpanKindType {
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
	Processes map[string]*JaegerProcessValueEntry `json:"processes"`
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
	Tags          []JaegerTagEntry               `json:"tags"`          // Tags associated with the span
	Logs          []map[string]interface{} `json:"logs"`          // Log entries associated with the span
	ProcessID     string                   `json:"processID"`     // Process ID
	Warnings      interface{}              `json:"-"`             // Warnings associated with the span TODO: check here @xunzhou24
}

// JaegerTagEntry represents a tag entry in a Jaeger trace.
type JaegerTagEntry struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// ToAttributeEntry converts a JaegerTagEntry to an AttributeEntry.
func (j *JaegerTagEntry) ToAttributeEntry() AttributeEntry {
	return AttributeEntry{
		Key:   j.Key,
		Type:  j.Type,
		Value: j.Value,
	}
}

// JaegerProcessValueEntry represents the process information in a Jaeger trace.
type JaegerProcessValueEntry struct {
	ServiceName string          `json:"serviceName"` // Name of the service
	Tags        []JaegerTagEntry `json:"tags"`         // Tags associated with the process
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
func (j *JaegerTraceSpan) ToSimplifiedTraceSpan(processMap map[string]*JaegerProcessValueEntry) *SimplifiedTraceSpan {
	span := &SimplifiedTraceSpan{
		TraceID:       j.TraceID,
		SpanID:        j.SpanID,
		ParentID:      j.ParentID,
		OperationName: j.OperationName,
		StartTime:     time.Unix(0, j.StartTime*int64(time.Microsecond)),
		Duration:      j.Duration,
		ServiceName:  	processMap[j.ProcessID].ServiceName,
	}

	// convert tags to attributes
	span.AttributeMap = make(map[string]AttributeEntry)
	for _, tag := range j.Tags {
		span.AttributeMap[tag.Key] = tag.ToAttributeEntry()
	}

	// parse parent span ID from references
	for _, ref := range j.References {
		if ref["refType"] == "CHILD_OF" {
			parentSpanID := ref["spanID"]
			if parentSpanID != "" {
				span.ParentID = parentSpanID
			}
			break
		}
	}

	// parse span kind
	spanKind := UNSPECIFIED
	if spanKindValue, exist := span.AttributeMap["span.kind"]; exist {
		spanKind = convertJaegerTraceTagValueToSpanKind(spanKindValue.Value.(string))
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

	// For other system, tags are used to determine the semantic convention.
	for _, tag := range j.Tags {
		if tag.Key == "rpc.system" {
			return SemanticConventionTypeRPC
		}
		if tag.Key == "messaging.system" {
			return SemanticConventionTypeMessaging
		}
		if strings.HasPrefix(tag.Key, "db.") {
			return SemanticConventionTypeDatabase
		}
	}

	return SemanticConventionTypeUnknown
}


// TempoTrace represents a trace in Tempo format.
type TempoTrace struct {
	Batches     []TempoSpanBatchElement  `json:"batches"`
}


// TempoSpanBatchElement represents a batch of spans in a Tempo trace.
// It contains information about resources, scope spans, and their associated attributes.
// Below is an example:
//  "resource": {
//   "attributes": [
//    {
//     "key": "container.id",
//     "value": {
//      "stringValue": "1f20862f01e732486b72e16ee447f4cdf4ffc5424555561e3f4e070864c20d09"
//     }
//    },
//    {
//     "key": "host.arch",
//     "value": {
//      "stringValue": "amd64"
//     }
//    },
//    ...
//   ]
//  },
//  "scopeSpans": [
//   {
//    "scope": {
//     "name": "io.opentelemetry.netty-4.1",
//     "version": "1.28.0-alpha"
//    },
//    "spans": [
//     {
//      "traceId": "vuFZzuKAbn8wb5p3lSVqgg==",
//      "spanId": "fWjJ6TJR/u0=",
//      "parentSpanId": "cFIvnt3VSeo=",
//      "name": "POST",
//      "kind": "SPAN_KIND_CLIENT",
//      "startTimeUnixNano": "1744856899143930734",
//      "endTimeUnixNano": "1744856899904048439",
//      "attributes": [
//       {
//        "key": "net.peer.port",
//        "value": {
//         "intValue": "12346"
//        }
//       },
//       {
//        "key": "user_agent.original",
//        "value": {
//         "stringValue": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0"
//        }
//       },
//       ...
//      ],
//      "status": {}
//     }
//    ]
//   }
//  ]
// 
// Resource:
// - A map where the key is a string representing the resource name, and the value is a struct
//   containing a list of attributes. Each attribute has a key and a value. The value can be
//   one of the following types: string, integer, boolean, or double.
// 
// ScopeSpans:
// - A list of scope spans in the Tempo trace. Each scope span contains:
//   - Scope: Metadata about the scope, including its name and version.
//   - Spans: A list of TempoTraceSpan objects representing individual spans within the scope.
type TempoSpanBatchElement struct {
	Resource struct {
		Attributes []TempoAttributeEntry `json:"attributes"`
	} `json:"resource"`

	// ScopeSpans is a list of scope spans in the Tempo trace.
	// Each scope span contains a scope and a list of spans.
	ScopeSpans []struct {
		Scope struct {
			Name    string `json:"name"`    // Name of the scope
			Version string `json:"version"` // Version of the scope
		} `json:"scope"` // Scope of the span
		Spans []TempoTraceSpan `json:"spans"` // List of spans in the scope
	} `json:"scopeSpans"` // List of scope spans in the Tempo trace
}


// TempoTraceSpan represents a span in a Tempo trace.
type TempoTraceSpan struct {
	TraceID            string `json:"traceId"`            // Unique identifier for the trace
	SpanID             string `json:"spanId"`            // Unique identifier for the span
	ParentSpanId 	string `json:"parentSpanId"`      // Unique identifier for the parent span (this field may not exist in some cases)
	Name               string `json:"name"`              // Name of the span
	Kind               string `json:"kind"`              // Kind of the span (e.g., SPAN_KIND_SERVER)
	StartTimeUnixNano  string `json:"startTimeUnixNano"`  // Start time in Unix nanoseconds
	EndTimeUnixNano    string `json:"endTimeUnixNano"`    // End time in Unix nanoseconds
	Attributes         []TempoAttributeEntry `json:"attributes"` // List of attributes
	Status interface{} `json:"status"` // Status of the span
}

type TempoAttributeEntry struct {
	Key   string `json:"key"` // Key of the attribute
	// All values are of string type, and should be converted to the appropriate type.
	Value struct {
		StringValue string  `json:"stringValue,omitempty"` // String value of the attribute
		IntValue    string   `json:"intValue,omitempty"`    // Integer value of the attribute
		BoolValue   string    `json:"boolValue,omitempty"`   // Boolean value of the attribute
		DoubleValue string `json:"doubleValue,omitempty"` // Double value of the attribute
	} `json:"value"` // Value of the attribute
}

// ToSimplifiedTrace converts a TempoTrace to a SimplifiedTrace.
func (t *TempoTrace) ToSimplifiedTrace() *SimplifiedTrace {
	spanMap := make(map[string]*SimplifiedTraceSpan)
	startTime := time.Now()
	var traceId string
	for _, batch := range t.Batches {
		for _, scopeSpan := range batch.ScopeSpans {
			for _, span := range scopeSpan.Spans {
				hexSpanID, err := utils.Base64ToHex(span.SpanID)
				if err != nil {
					log.Err(err).Msgf("[TempoTrace.ToSimplifiedTrace] Failed to decode span ID: %v", err)
					continue
				}
				spanMap[hexSpanID] = span.ToSimplifiedTraceSpan(batch.Resource.Attributes)
				if spanMap[hexSpanID].StartTime.Before(startTime) {
					startTime = spanMap[hexSpanID].StartTime
				}
				traceId = span.TraceID
			}
		}
	}

	// convert trace ID from base64 to hex
	decodedTraceID, err := utils.Base64ToHex(traceId)
	if err != nil {
		log.Err(err).Msgf("[TempoTrace.ToSimplifiedTrace] Failed to decode trace ID: %s", traceId)
		return nil
	}
	return &SimplifiedTrace{
		TraceID:   decodedTraceID,
		SpanMap:   spanMap,
		StartTime: startTime,
	}
}

// ToSimplifiedTraceSpan converts a TempoTraceSpan to a SimplifiedTraceSpan.
func (t *TempoTraceSpan) ToSimplifiedTraceSpan(resourceAttributes []TempoAttributeEntry) *SimplifiedTraceSpan {
	span := &SimplifiedTraceSpan{}

	// Tempo returns trace ID and span ID in base64 format.
	// We need to decode them to hex format.
	decodedTraceID, err := utils.Base64ToHex(t.TraceID)
	if err != nil {
		log.Err(err).Msgf("[TempoTraceSpan.ToSimplifiedTraceSpan] Failed to decode trace ID: %v", err)
		return nil
	}
	decodedSpanID, err := utils.Base64ToHex(t.SpanID)
	if err != nil {
		log.Err(err).Msgf("[TempoTraceSpan.ToSimplifiedTraceSpan] Failed to decode span ID: %v", err)
		return nil
	}
	span.TraceID = decodedTraceID
	span.SpanID = decodedSpanID
	span.OperationName = t.Name

	// Find and set the parent span ID.
	// If not found, an empty string is left.
	if t.ParentSpanId != "" {
		decodedParentSpanID, err := utils.Base64ToHex(t.ParentSpanId)
		if err != nil {
			log.Err(err).Msgf("[TempoTraceSpan.ToSimplifiedTraceSpan] Failed to decode parent span ID: %v", err)
			return nil
		}
		span.ParentID = decodedParentSpanID
	}

	var serviceName string
	for _, resourceAttribute := range resourceAttributes {
		if resourceAttribute.Key == "service.name" {
			serviceName = resourceAttribute.Value.StringValue
			break
		}
	}
	if serviceName == "" {
		log.Warn().Msgf("[TempoTraceSpan.ToSimplifiedTraceSpan] Service name not found in resource attributes, trace ID: %s, span ID: %s", t.TraceID, t.SpanID)
	}
	span.ServiceName = serviceName

	// convert start time from Unix nanoseconds to time.Time
	startTimeUnixNano, err := strconv.ParseInt(t.StartTimeUnixNano, 10, 64)
	if err != nil {
		log.Err(err).Msgf("[TempoTraceSpan.ToSimplifiedTraceSpan] Failed to parse start time: %v", err)
		return nil
	}
	span.StartTime = time.Unix(0, startTimeUnixNano*int64(time.Nanosecond))
	// convert duration from Unix nanoseconds to microseconds
	endTimeUnixNano, err := strconv.ParseInt(t.EndTimeUnixNano, 10, 64)
	if err != nil {
		log.Err(err).Msgf("[TempoTraceSpan.ToSimplifiedTraceSpan] Failed to parse end time: %v", err)
		return nil
	}
	span.Duration = (endTimeUnixNano - startTimeUnixNano) / int64(time.Microsecond)

	// convert attributes
	span.AttributeMap = make(map[string]AttributeEntry)
	attributesFromSpan := convertTempoAttributesToAttributeEntries(t.Attributes)
	attributesFromResource := convertTempoAttributesToAttributeEntries(resourceAttributes)
	maps.Copy(span.AttributeMap, attributesFromSpan)
	for key, value := range attributesFromResource {
		// if the key already exists in the span attributes, skip it
		if _, exist := span.AttributeMap[key]; exist {
			continue
		}
		span.AttributeMap[key] = value
	}

	// parse span kind
	spanKind := convertTempoTraceKindToSpanKind(t.Kind)
	span.SpanKind = spanKind

	// parse semantic convention
	semanticConvention := t.InferSemanticConvention()
	span.SemanticConvention = semanticConvention

	return span
}


// convertTempoAttributesToAttributeEntries converts a list of TempoAttributeEntry to a map of AttributeEntry.
func convertTempoAttributesToAttributeEntries(tempoAttributes []TempoAttributeEntry) map[string]AttributeEntry {
	attributeMap := make(map[string]AttributeEntry)
	for _, attribute := range tempoAttributes {
		if attribute.Key == "" {
			continue
		}
		if attribute.Value.StringValue != "" {
			attributeMap[attribute.Key] = AttributeEntry{
				Key:   attribute.Key,
				Type:  "string",
				Value: attribute.Value.StringValue,
			}
		} else if attribute.Value.IntValue != "" {
			intValue, err := strconv.ParseInt(attribute.Value.IntValue, 10, 64)
			if err != nil {
				log.Err(err).Msgf("[convertTempoAttributesToAttributeEntries] Failed to parse int value: %v", err)
				continue
			}
			attributeMap[attribute.Key] = AttributeEntry{
				Key:   attribute.Key,
				Type:  "int",
				Value: intValue,
			}
		} else if attribute.Value.BoolValue != "" {
			boolValue, err := strconv.ParseBool(attribute.Value.BoolValue)
			if err != nil {
				log.Err(err).Msgf("[convertTempoAttributesToAttributeEntries] Failed to parse bool value: %v", err)
				continue
			}
			attributeMap[attribute.Key] = AttributeEntry{
				Key:   attribute.Key,
				Type:  "bool",
				Value: boolValue,
			}
		} else if attribute.Value.DoubleValue != "" {
			doubleValue, err := strconv.ParseFloat(attribute.Value.DoubleValue, 64)
			if err != nil {
				log.Err(err).Msgf("[convertTempoAttributesToAttributeEntries] Failed to parse double value: %v", err)
				continue
			}
			attributeMap[attribute.Key] = AttributeEntry{
				Key:   attribute.Key,
				Type:  "double",
				Value: doubleValue,
			}
		} else {
			log.Error().Msgf("[convertTempoAttributesToAttributeEntries] Unsupported attribute value type: %v", attribute.Value)
			continue
		}
	}
	return attributeMap
}

// convertTempoTraceKindToSpanKind converts a Tempo attribute value to a SpanKindType.
// If the attribute value is not recognized, it returns UNSPECIFIED.
func convertTempoTraceKindToSpanKind(tempoSpanKind string) SpanKindType {
	switch tempoSpanKind {
	case "SPAN_KIND_CLIENT":
		return CLIENT
	case "SPAN_KIND_SERVER":
		return SERVER
	case "SPAN_KIND_PRODUCER":
		return PRODUCER
	case "SPAN_KIND_CONSUMER":
		return CONSUMER
	case "SPAN_KIND_INTERNAL":
		return INTERNAL
	default:
		return UNSPECIFIED
	}
}

// InferSemanticConvention infers the semantic convention of a TempoTraceSpan.
// Unsupported semantic conventions are returned as SemanticConventionTypeUnknown.
// Note: the result may not be accurate, as it's based on the attributes and name format.
// We use required attributes and name format that are specific to the semantic conventions to determine the semantic convention.
// See [OpenTelemetry specification](https://opentelemetry.io/docs/specs/semconv/) for more details.
// TODO: support more semantic conventions @xunzhou24
func (j *TempoTraceSpan) InferSemanticConvention() SemanticConventionType {
	// For HTTP, the span name is '{method} {target}' or '{method}'
	httpOperationNameRegex := `^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS|TRACE)(?:\s+(\S+))?$`
	if matched, _ := regexp.MatchString(httpOperationNameRegex, j.Name); matched {
		return SemanticConventionTypeHTTP
	}

	// For RPC and Messaging system, attributes are used to determine the semantic convention.
	for _, attribute := range j.Attributes {
		if attribute.Key == "rpc.system" {
			return SemanticConventionTypeRPC
		}
		if attribute.Key == "messaging.system" {
			return SemanticConventionTypeMessaging
		}
		if strings.HasPrefix(attribute.Key, "db.") {
			return SemanticConventionTypeDatabase
		}
	}

	return SemanticConventionTypeUnknown
}
