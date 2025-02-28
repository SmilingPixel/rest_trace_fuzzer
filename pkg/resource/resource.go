package resource

import (
	"fmt"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"
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

// ResourceInteger represents a integer resource.
type ResourceInteger struct {
	Value int64
}

func NewResourceInteger(value int64) *ResourceInteger {
	return &ResourceInteger{
		Value: value,
	}
}

func (r *ResourceInteger) String() string {
	return strconv.FormatInt(r.Value, 10)
}

func (r *ResourceInteger) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeFloat
}

// ResourceFloat represents a float resource.
type ResourceFloat struct {
	Value float64
}

func NewResourceFloat(value float64) *ResourceFloat {
	return &ResourceFloat{
		Value: value,
	}
}

func (r *ResourceFloat) String() string {
	return strconv.FormatFloat(r.Value, 'f', -1, 64)
}

func (r *ResourceFloat) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeFloat
}

// ResourceString represents a string resource.
type ResourceString struct {
	Value string
}

func NewResourceString(value string) *ResourceString {
	return &ResourceString{
		Value: value,
	}
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

func NewResourceBoolean(value bool) *ResourceBoolean {
	return &ResourceBoolean{
		Value: value,
	}
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

func NewResourceObject(value map[string]Resource) *ResourceObject {
	return &ResourceObject{
		Value: value,
	}
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

func NewResourceArray(value []Resource) *ResourceArray {
	return &ResourceArray{
		Value: value,
	}
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

// NewResourceFromValue creates a new resource.
// For non-primitive types, it recursively creates sub-resources.
func NewResourceFromValue(name string, value interface{}) (Resource, error) {
	propertyType := static.DeterminePropertyType(value)
	switch propertyType {
	case static.SimpleAPIPropertyTypeString:
		return &ResourceString{
			Value: value.(string),
		}, nil
	case static.SimpleAPIPropertyTypeFloat:
		return &ResourceFloat{
			Value: utils.ConvertFloatTo64BitType(value),
		}, nil
	case static.SimpleAPIPropertyTypeInteger:
		return &ResourceInteger{
			Value: utils.ConvertIntTo64BitType(value),
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
			subResource, err := NewResourceFromValue(key, val)
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
			subResource, err := NewResourceFromValue(fmt.Sprintf("%s[%d]", name, i), val)
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
