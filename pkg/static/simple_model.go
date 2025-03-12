package static

import (
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

const (
	// SimpleAPIMethodTypeHTTP represents an HTTP API method.
	SimpleAPIMethodTypeHTTP SimpleAPIMethodType = "HTTP"
	// SimpleAPIMethodTypeGRPC represents a gRPC API method.
	SimpleAPIMethodTypeGRPC SimpleAPIMethodType = "gRPC"
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

	// Unkown
	SimpleAPIPropertyTypeUnknown SimpleAPIPropertyType = "unknown"
)

// SimpleAPIMethod represents an API method on a specific endpoint.
//   - If the API is an HTTP API, the method is the HTTP method, such as GET, POST, PUT, DELETE, and Endpoint is the URL path.
//   - If the API is a gRPC API, the method is the gRPC method name.
//
// Endpoint is the URL path or the gRPC method name.
//
// You should use the struct by value, not by pointer.
type SimpleAPIMethod struct {
	Endpoint string              `json:"endpoint"`
	Method   string              `json:"method"`
	Typ      SimpleAPIMethodType `json:"type"`
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
