package resource

import (
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

// Resource represents a resource in the system.
// It is an interface that can be implemented by different types of resources.
type Resource interface {

	// String returns the string representation of the resource.
	String() string
}

// ResourceNumber represents a number resource.
type ResourceNumber struct {
	Value int64
}

func (r *ResourceNumber) String() string {
	return strconv.FormatInt(r.Value, 10)
}

// ResourceString represents a string resource.
type ResourceString struct {
	Value string
}

func (r *ResourceString) String() string {
	return r.Value
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

// ResourceObject represents an object resource.
type ResourceObject struct {
	Value map[string]interface{}
}

func (r *ResourceObject) String() string {
	s, err := sonic.MarshalString(r.Value)
	if err != nil {
		log.Err(err).Msg("[ResourceObject.String] Failed to marshal object resource")
		return "{}"
	}
	return s
}

// ResourceArray represents an array resource.
type ResourceArray struct {
	Value []interface{}
}

func (r *ResourceArray) String() string {
	s, err := sonic.MarshalString(r.Value)
	if err != nil {
		log.Err(err).Msg("[ResourceArray.String] Failed to marshal array resource")
		return "[]"
	}
	return s
}
