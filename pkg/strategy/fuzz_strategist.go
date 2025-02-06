package strategy

import (
	"resttracefuzzer/pkg/resource"

	"github.com/getkin/kin-openapi/openapi3"
)

// FuzzStrategist is a strategist for fuzzing process.
type FuzzStrategist struct {
	// SchemaToValueStrategy is the strategy for generating values from schemas.
	SchemaToValueStrategy *SchemaToValueStrategy
}

// NewFuzzStrategist creates a new FuzzStrategist.
func NewFuzzStrategist(
	resourceManager *resource.ResourceManager,
) *FuzzStrategist {
	schemaToValueStrategy := NewSchemaToValueStrategy(resourceManager)
	return &FuzzStrategist{
		SchemaToValueStrategy: schemaToValueStrategy,
	}
}

// GenerateValueForSchema generates a value for a given schema.
func (s *FuzzStrategist) GenerateValueForSchema(schema *openapi3.SchemaRef) (interface{}, error) {
	return s.SchemaToValueStrategy.GenerateValueForSchema(schema)
}
