package static

import (
	"math/rand/v2"
	"resttracefuzzer/pkg/utils"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

// SimpleAPIMethodType represents the type of an API method.
type SimpleAPIMethodType string

func (t SimpleAPIMethodType) String() string {
	return string(t)
}

func (t SimpleAPIMethodType) MarshalJSON() ([]byte, error) {
	return sonic.Marshal(t.String())
}

func (t *SimpleAPIMethodType) UnmarshalJSON(data []byte) error {
	*t = SimpleAPIMethodType(data)
	return nil
}

// SimpleAPIPropertyType represents the type of an API property.
type SimpleAPIPropertyType string

func (t SimpleAPIPropertyType) String() string {
	return string(t)
}

func (t SimpleAPIPropertyType) MarshalJSON() ([]byte, error) {
	return sonic.Marshal(t.String())
}

func (t *SimpleAPIPropertyType) UnmarshalJSON(data []byte) error {
	*t = SimpleAPIPropertyType(data)
	return nil
}

// OpenAPITypes2SimpleAPIPropertyType converts an OpenAPI schema type to a SimpleAPIPropertyType.
func OpenAPITypes2SimpleAPIPropertyType(types *openapi3.Types) SimpleAPIPropertyType {
	switch {
	case types.Includes(openapi3.TypeString):
		return SimpleAPIPropertyTypeString
	case types.Includes(openapi3.TypeInteger):
		return SimpleAPIPropertyTypeInteger
	case types.Includes(openapi3.TypeNumber):
		return SimpleAPIPropertyTypeFloat
	case types.Includes(openapi3.TypeBoolean):
		return SimpleAPIPropertyTypeBoolean
	case types.Includes(openapi3.TypeObject):
		return SimpleAPIPropertyTypeObject
	case types.Includes(openapi3.TypeArray):
		return SimpleAPIPropertyTypeArray
	default:
		log.Warn().Msgf("[OpenAPITypes2SimpleAPIPropertyType] Unknown types: %v", types)
		return SimpleAPIPropertyTypeUnknown
	}
}

// Name2SimpleAPIPropertyType converts a string to a SimpleAPIPropertyType.
func Name2SimpleAPIPropertyType(name string) SimpleAPIPropertyType {
	switch name {
	case "float":
		return SimpleAPIPropertyTypeFloat
	case "integer":
		return SimpleAPIPropertyTypeInteger
	case "string":
		return SimpleAPIPropertyTypeString
	case "boolean":
		return SimpleAPIPropertyTypeBoolean
	case "object":
		return SimpleAPIPropertyTypeObject
	case "array":
		return SimpleAPIPropertyTypeArray
	default:
		return SimpleAPIPropertyTypeUnknown
	}
}

// DeterminePropertyType determines the type of a property.
// It uses reflection to determine the type of the value.
func DeterminePropertyType(value any) SimpleAPIPropertyType {
	switch value.(type) {
	case string:
		return SimpleAPIPropertyTypeString
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return SimpleAPIPropertyTypeInteger
	case float32, float64:
		return SimpleAPIPropertyTypeFloat
	case bool:
		return SimpleAPIPropertyTypeBoolean
	case map[string]interface{}:
		return SimpleAPIPropertyTypeObject
	case []interface{}:
		return SimpleAPIPropertyTypeArray
	default:
		log.Warn().Msgf("[DeterminePropertyType] Unknown type: %T", value)
		return SimpleAPIPropertyTypeUnknown
	}
}

// DefaultValueForPrimitiveSimpleAPIPropertyType returns the default value for a primitive SimpleAPIPropertyType.
func DefaultValueForPrimitiveSimpleAPIPropertyType(typ SimpleAPIPropertyType) any {
	switch typ {
	case SimpleAPIPropertyTypeString:
		return "114-514"
	case SimpleAPIPropertyTypeInteger:
		return 114514
	case SimpleAPIPropertyTypeFloat:
		return 114.514
	case SimpleAPIPropertyTypeBoolean:
		return true
	default:
		log.Warn().Msgf("[DefaultValueForPrimitiveSimpleAPIPropertyType] Unknown type or non-primitive type: %v", typ)
		return nil
	}
}

// RandomValueForPrimitiveSimpleAPIPropertyType generates a random value for a SimpleAPIPropertyType.
func RandomValueForPrimitiveSimpleAPIPropertyType(typ SimpleAPIPropertyType) any {
	switch typ {
	case SimpleAPIPropertyTypeString:
		randLength := rand.IntN(114) + 1
		return utils.RandStringBytes(randLength)
	case SimpleAPIPropertyTypeInteger:
		return rand.IntN(114514)
	case SimpleAPIPropertyTypeFloat:
		return rand.Float64() + float64(rand.IntN(114514))
	case SimpleAPIPropertyTypeBoolean:
		return rand.IntN(2) == 1
	default:
		log.Warn().Msgf("[RandomValueForPrimitiveSimpleAPIPropertyType] Unknown type or non-primitive type: %v", typ)
		return nil
	}
}

// IsPrimitiveSimpleAPIPropertyType returns whether a SimpleAPIPropertyType is a primitive type.
func IsPrimitiveSimpleAPIPropertyType(typ SimpleAPIPropertyType) bool {
	switch typ {
	case SimpleAPIPropertyTypeString, SimpleAPIPropertyTypeInteger, SimpleAPIPropertyTypeFloat, SimpleAPIPropertyTypeBoolean:
		return true
	default:
		return false
	}
}

const (
	// SimpleAPIMethodTypeHTTP represents an HTTP API method.
	SimpleAPIMethodTypeHTTP SimpleAPIMethodType = "HTTP"
	// SimpleAPIMethodTypeGRPC represents a gRPC API method.
	SimpleAPIMethodTypeGRPC SimpleAPIMethodType = "gRPC"
	// SimpleAPIMethodTypeUnknown represent an unknown type of API method
	SimpleAPIMethodTypeUnknown SimpleAPIMethodType = "unknown"
	// TODO: Add more types if needed, such as MessageQueue, etc. @xunzhou24

	// SimpleAPIPropertyTypeFloat
	SimpleAPIPropertyTypeFloat SimpleAPIPropertyType = "float"

	SimpleAPIPropertyTypeInteger SimpleAPIPropertyType = "integer"

	// SimpleAPIPropertyTypeString
	SimpleAPIPropertyTypeString SimpleAPIPropertyType = "string"

	// SimpleAPIPropertyTypeBoolean
	SimpleAPIPropertyTypeBoolean SimpleAPIPropertyType = "boolean"

	// SimpleAPIPropertyTypeObject
	SimpleAPIPropertyTypeObject SimpleAPIPropertyType = "object"

	// SimpleAPIPropertyTypeArray
	SimpleAPIPropertyTypeArray SimpleAPIPropertyType = "array"

	// Empty, None, etc.
	SimpleAPIPropertyTypeEmpty SimpleAPIPropertyType = "empty"

	// Unkown
	SimpleAPIPropertyTypeUnknown SimpleAPIPropertyType = "unknown"
)

// SimpleAPIMethod represents an API method on a specific endpoint.
//   - If the API is an HTTP API, the method is the HTTP method, such as GET, POST, PUT, DELETE, and Endpoint is the URL path.
//   - If the API is a gRPC API, the endpoint is the gRPC method name, and the method field is undefined
//
// Endpoint is the URL path or the gRPC method name.
//
// You should use the struct by value, not by pointer.
type SimpleAPIMethod struct {
	Endpoint string              `json:"endpoint"`
	Method   string              `json:"method"`
	Typ      SimpleAPIMethodType `json:"type"`
}

// CompareSimpleAPIMethod compares two SimpleAPIMethods.
// It treats all fields as strings and compares them lexicographically.
// It returns -1 if a < b, 0 if a == b, and 1 if a > b.
func CompareSimpleAPIMethod(a, b SimpleAPIMethod) int {
	if a.Endpoint != b.Endpoint {
		return strings.Compare(a.Endpoint, b.Endpoint)
	}
	if a.Method != b.Method {
		return strings.Compare(a.Method, b.Method)
	}
	return strings.Compare(a.Typ.String(), b.Typ.String())
}

// InternalServiceEndpoint represents an endpoint of an internal service.
// It can be used as node in data flow graph.
// It implements [resttracefuzzer/pkg/utils/AbstractNode] interface, to support graph related algorithms.
type InternalServiceEndpoint struct {
	ServiceName     string          `json:"serviceName"`
	SimpleAPIMethod SimpleAPIMethod `json:"simpleAPIMethod"`
}

// ID returns a unique identifier for an InternalServiceEndpoint.
// We concatenate the service name and the endpoint to create a unique ID.
func (e InternalServiceEndpoint) ID() string {
	return e.ServiceName + "##" + e.SimpleAPIMethod.Endpoint
}

// CompareInternalServiceEndpoint compares two InternalServiceEndpoints.
// It compares the service name and the SimpleAPIMethod.
// It returns -1 if a < b, 0 if a == b, and 1 if a > b.
func CompareInternalServiceEndpoint(a, b InternalServiceEndpoint) int {
	if a.ServiceName != b.ServiceName {
		return strings.Compare(a.ServiceName, b.ServiceName)
	}
	return CompareSimpleAPIMethod(a.SimpleAPIMethod, b.SimpleAPIMethod)
}

// SimpleAPIProperty represents a property of an API.
// For example, the property can be a parameter, a request body, a response, or any variables defined in them.
// You should use the struct by value, not by pointer.
// TODO: Add more @xunzhou24
// TODO: fill type field when created @xunzhou24
type SimpleAPIProperty struct {
	// Name is the name of the property.
	Name string `json:"name"`

	// Typ is the type of the property.
	Typ SimpleAPIPropertyType `json:"type"`
}
