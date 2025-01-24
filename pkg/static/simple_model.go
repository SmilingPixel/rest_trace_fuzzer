package static

import "github.com/bytedance/sonic"

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

const (
	// SimpleAPIMethodTypeHTTP represents an HTTP API method.
	SimpleAPIMethodTypeHTTP SimpleAPIMethodType = "HTTP"
	// SimpleAPIMethodTypeGRPC represents a gRPC API method.
	SimpleAPIMethodTypeGRPC SimpleAPIMethodType = "gRPC"
	// TODO: Add more types if needed, such as MessageQueue, etc. @xunzhou24

	// SimpleAPIPropertyTypeNumber
	// TODO: do we need to distinguish between integer and float? @xunzhou24
	SimpleAPIPropertyTypeNumber SimpleAPIPropertyType = "number"

	// SimpleAPIPropertyTypeString
	SimpleAPIPropertyTypeString SimpleAPIPropertyType = "string"

	// SimpleAPIPropertyTypeBoolean
	SimpleAPIPropertyTypeBoolean SimpleAPIPropertyType = "boolean"

	// SimpleAPIPropertyTypeObject
	SimpleAPIPropertyTypeObject SimpleAPIPropertyType = "object"

	// SimpleAPIPropertyTypeArray
	SimpleAPIPropertyTypeArray SimpleAPIPropertyType = "array"
)

// SimpleAPIMethod represents an API method on a specific endpoint.
//
//	- If the API is an HTTP API, the method is the HTTP method, such as GET, POST, PUT, DELETE, and Endpoint is the URL path.
//	- If the API is a gRPC API, the method is the gRPC method name.
//
// You should use the struct by value, not by pointer.
type SimpleAPIMethod struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	Type     SimpleAPIMethodType `json:"type"`
}

// SimpleAPIProperty represents a property of an API.
// For example, the property can be a parameter, a request body, a response, or any variables defined in them.
// You should use the struct by value, not by pointer.
// TODO: Add more @xunzhou24
// TODO: fill type field when creating @xunzhou24
type SimpleAPIProperty struct {
	// Name is the name of the property.
	Name        string

	// Type is the type of the property.
	Type 	    SimpleAPIPropertyType
}
