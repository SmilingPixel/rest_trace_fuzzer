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
// name is the name, type or key etc. of the value, and schema is the schema of the value.
func (s *FuzzStrategist) GenerateValueForSchema(name string, schema *openapi3.SchemaRef) (resource.Resource, error) {
	return s.SchemaToValueStrategy.GenerateValueForSchema(name, schema)
}
