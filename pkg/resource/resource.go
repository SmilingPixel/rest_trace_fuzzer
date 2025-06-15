package resource

import (
	"fmt"
	"hash/fnv"
	"math"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

// Resource represents a resource in the system.
// It is an interface that can be implemented by different types of resources.
// You can access value of a resource by access field `Value` directly, or use the `GetRawValue` method.
// The difference between them are:
//  - `Value` stores a parsed value, e.g., a map[string]interface{} would be stored as a `map[string]Resource` type.
//  - `GetRawValue` returns the raw value, e.g., you can get the original map[string]interface{} value by calling `GetRawValue` of a ResourceObject.
// As to the usage:
//  - If you want to reduce the overhead (GetRawValue would try to convert the value to the original type), you can use `Value` directly.
//  - If you do not want to include more dependencies, you can use `GetRawValue` to access the original value.
// The same with `SetByRawValue` and `Value`.
type Resource interface {

	// String returns the string representation of the resource.
	String() string

	// ToJSONObject returns the JSON object representation of the resource.
	ToJSONObject() any

	// Typ returns the type of the resource.
	Typ() static.SimpleAPIPropertyType

	// Hashcode returns the hashcode of the resource.
	// It is used to compare resources (not precisely).
	Hashcode() uint64

	// GetRawValue returns the raw value of the resource.
	GetRawValue() any

	// SetByRawValue sets the value of the resource.
	SetByRawValue(value any)

	// Copy creates a deep copy of the resource.
	Copy() Resource
}

// type ResourceEmpty represents an empty resource.
type ResourceEmpty struct {
}

func NewResourceEmpty() *ResourceEmpty {
	return &ResourceEmpty{}
}

func (r *ResourceEmpty) String() string {
	return "{}"
}

func (r *ResourceEmpty) ToJSONObject() any {
	return nil
}

// Type of ResourceEmpty is unknown.
func (r *ResourceEmpty) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeUnknown
}

func (r *ResourceEmpty) Hashcode() uint64 {
	return 0
}

func (r *ResourceEmpty) GetRawValue() any {
	return nil
}

func (r *ResourceEmpty) SetByRawValue(value any) {
	// We assume the value is map[string]interface{} type.
	// Do nothing, as the resource is empty.
}

func (r *ResourceEmpty) Copy() Resource {
	return &ResourceEmpty{}
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

func (r *ResourceInteger) ToJSONObject() any {
	return r.Value
}

func (r *ResourceInteger) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeInteger
}

func (r *ResourceInteger) Hashcode() uint64 {
	// As int64 and uint64 have the same scope, we can directly convert int64 to uint64.
	return uint64(r.Value)
}

func (r *ResourceInteger) GetRawValue() any {
	return r.Value
}

func (r *ResourceInteger) SetByRawValue(value any) {
	// We assume the value is int64 type.
	r.Value = value.(int64)
}

func (r *ResourceInteger) Copy() Resource {
	return &ResourceInteger{
		Value: r.Value,
	}
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

func (r *ResourceFloat) ToJSONObject() any {
	return r.Value
}

func (r *ResourceFloat) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeFloat
}

func (r *ResourceFloat) Hashcode() uint64 {
	return math.Float64bits(r.Value)
}

func (r *ResourceFloat) GetRawValue() any {
	return r.Value
}

func (r *ResourceFloat) SetByRawValue(value any) {
	// We assume the value is float64 type.
	r.Value = value.(float64)
}

func (r *ResourceFloat) Copy() Resource {
	return &ResourceFloat{
		Value: r.Value,
	}
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

func (r *ResourceString) ToJSONObject() any {
	return r.Value
}

func (r *ResourceString) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeString
}

func (r *ResourceString) Hashcode() uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(r.Value))
	return hasher.Sum64()
}

func (r *ResourceString) GetRawValue() any {
	return r.Value
}

func (r *ResourceString) SetByRawValue(value any) {
	r.Value = value.(string)
}

func (r *ResourceString) Copy() Resource {
	return &ResourceString{
		Value: r.Value,
	}
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

func (r *ResourceBoolean) ToJSONObject() any {
	return r.Value
}

func (r *ResourceBoolean) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeBoolean
}

func (r *ResourceBoolean) Hashcode() uint64 {
	if r.Value {
		return 1
	} else {
		return 0
	}
}

func (r *ResourceBoolean) GetRawValue() any {
	return r.Value
}

func (r *ResourceBoolean) SetByRawValue(value any) {
	// We assume the value is bool type.
	r.Value = value.(bool)
}

func (r *ResourceBoolean) Copy() Resource {
	return &ResourceBoolean{
		Value: r.Value,
	}
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
	s, err := sonic.MarshalString(r.ToJSONObject())
	if err != nil {
		log.Err(err).Msg("[ResourceObject.String] Failed to marshal object resource")
		return "{}"
	}
	return s
}

func (r *ResourceObject) ToJSONObject() any {
	result := make(map[string]any)
	for key, value := range r.Value {
		result[key] = value.ToJSONObject()
	}
	return result
}

func (r *ResourceObject) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeObject
}

func (r *ResourceObject) Hashcode() uint64 {
	hasher := fnv.New64a()
	var res = uint64(len(r.Value))
	for key, v := range r.Value {
		hasher.Write([]byte(key))
		keyHash := hasher.Sum64()
		res = (res*17 + keyHash + v.Hashcode())
	}
	return res
}

func (r *ResourceObject) GetRawValue() any {
	result := make(map[string]interface{})
	for key, value := range r.Value {
		result[key] = value.GetRawValue()
	}
	return result
}

func (r *ResourceObject) SetByRawValue(value any) {
	// We assume the value is map[string]interface{} type.
	objectValue := value.(map[string]interface{})
	for key, val := range objectValue {
		subResource, err := NewResourceFromValue(val)
		if err != nil {
			log.Err(err).Msg("[ResourceObject.SetValue] Failed to set value for object resource")
			return
		}
		r.Value[key] = subResource
	}
}

func (r *ResourceObject) Copy() Resource {
	result := &ResourceObject{
		Value: make(map[string]Resource),
	}
	for key, value := range r.Value {
		result.Value[key] = value.Copy()
	}
	return result
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
	s, err := sonic.MarshalString(r.ToJSONObject())
	if err != nil {
		log.Err(err).Msg("[ResourceArray.String] Failed to marshal array resource")
		return "[]"
	}
	return s
}

func (r *ResourceArray) ToJSONObject() any {
	result := make([]any, 0, len(r.Value))
	for _, value := range r.Value {
		result = append(result, value.ToJSONObject())
	}
	return result
}

func (r *ResourceArray) Typ() static.SimpleAPIPropertyType {
	return static.SimpleAPIPropertyTypeArray
}

func (r *ResourceArray) Hashcode() uint64 {
	var res = uint64(len(r.Value))
	for _, v := range r.Value {
		res = (res*17 + v.Hashcode())
	}
	return res
}

func (r *ResourceArray) GetRawValue() any {
	result := make([]interface{}, 0, len(r.Value))
	for _, value := range r.Value {
		result = append(result, value.GetRawValue())
	}
	return result
}

func (r *ResourceArray) SetByRawValue(value any) {
	// We assume the value is []interface{} type.
	arrayValue := value.([]interface{})
	for _, val := range arrayValue {
		subResource, err := NewResourceFromValue(val)
		if err != nil {
			log.Err(err).Msg("[ResourceArray.SetValue] Failed to set value for array resource")
			return
		}
		r.Value = append(r.Value, subResource)
	}
}

func (r *ResourceArray) Copy() Resource {
	result := &ResourceArray{
		Value: make([]Resource, 0, len(r.Value)),
	}
	for _, value := range r.Value {
		result.Value = append(result.Value, value.Copy())
	}
	return result
}

// NewResourceFromValue creates a new resource.
// For non-primitive types, it recursively creates sub-resources.
func NewResourceFromValue(value any) (Resource, error) {
	if value == nil {
		return NewResourceEmpty(), nil
	}
	
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
			subResource, err := NewResourceFromValue(val)
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
		for _, val := range arrayValue {
			subResource, err := NewResourceFromValue(val)
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
