package resource

import (
	"fmt"
	"resttracefuzzer/pkg/static"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

// Resource represents a resource in the system.
// It is an interface that can be implemented by different types of resources.
type Resource interface {
	// String returns the string representation of the resource.
	String() string
	// Typ returns the type of the resource.
	Typ() static.SimpleAPIPropertyType
}

// ResourceNumber represents a number resource.
type ResourceNumber struct {
	Value int64
}

func (r *ResourceNumber) String() string {
	return strconv.FormatInt(r.Value, 10)
}

func (r *ResourceNumber) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeNumber
}

// ResourceString represents a string resource.
type ResourceString struct {
	Value string
}

func (r *ResourceString) String() string {
	return r.Value
}

func (r *ResourceString) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeString
}

// ResourceBoolean represents a boolean resource.
type ResourceBoolean struct {
	Value bool
}

func (r *ResourceBoolean) String() string {
	s, err := sonic.MarshalString(r.Value)
	if err != nil {
		log.Err(err).Msg("[ResourceBoolean.String] Failed to marshal boolean resource")
		return "false"
	}
	return s
}

func (r *ResourceBoolean) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeBoolean
}

// ResourceObject represents an object resource.
type ResourceObject struct {
	Value map[string]Resource
}

func (r *ResourceObject) String() string {
	s, err := sonic.MarshalString(r.Value)
	if err != nil {
		log.Err(err).Msg("[ResourceObject.String] Failed to marshal object resource")
		return "{}"
	}
	return s
}

func (r *ResourceObject) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeObject
}

// ResourceArray represents an array resource.
type ResourceArray struct {
	Value []Resource
}

func (r *ResourceArray) String() string {
	s, err := sonic.MarshalString(r.Value)
	if err != nil {
		log.Err(err).Msg("[ResourceArray.String] Failed to marshal array resource")
		return "[]"
	}
	return s
}

func (r *ResourceArray) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeArray
}


// NewResource creates a new resource.
// For non-primitive types, it recursively creates sub-resources.
func NewResource(name string, value interface{}) (Resource, error) {
	propertyType := static.DeterminePropertyType(value)
	switch propertyType {
	case static.SimpleAPIPropertyTypeString:
		return &ResourceString{
			Value: value.(string),
		}, nil
	case static.SimpleAPIPropertyTypeNumber:
		return &ResourceNumber{
			Value: value.(int64),
		}, nil
	case static.SimpleAPIPropertyTypeBoolean:
		return &ResourceBoolean{
			Value: value.(bool),
		}, nil
	case static.SimpleAPIPropertyTypeObject:
		objectValue := value.(map[string]interface{})
		resource := &ResourceObject{
			Value: make(map[string]Resource),
		}
		for key, val := range objectValue {
			subResource, err := NewResource(key, val)
			if err != nil {
				return nil, err
			}
			resource.Value[key] = subResource
		}
		return resource, nil
	case static.SimpleAPIPropertyTypeArray:
		arrayValue := value.([]interface{})
		resource := &ResourceArray{
			Value: make([]Resource, 0, len(arrayValue)),
		}
		for i, val := range arrayValue {
			subResource, err := NewResource(fmt.Sprintf("%s[%d]", name, i), val)
			if err != nil {
				return nil, err
			}
			resource.Value = append(resource.Value, subResource)
		}
		return resource, nil
	default:
		return nil, fmt.Errorf("unsupported property type %s", propertyType)
	}
}

