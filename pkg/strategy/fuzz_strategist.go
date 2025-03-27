package strategy

import (
	"resttracefuzzer/pkg/resource"

	"github.com/getkin/kin-openapi/openapi3"
)

// FuzzStrategist is a strategist for fuzzing process.
type FuzzStrategist struct {
	// SchemaToValueStrategy is the strategy for generating values from schemas.
	SchemaToValueStrategy *SchemaToValueStrategy

	// ResourceMutateStrategy is the strategy for mutating resources.
	ResourceMutateStrategy *ResourceMutateStrategy
}

// NewFuzzStrategist creates a new FuzzStrategist.
func NewFuzzStrategist(
	resourceManager *resource.ResourceManager,
) *FuzzStrategist {
	schemaToValueStrategy := NewSchemaToValueStrategy(resourceManager)
	resourceMutateStrategy := NewResourceMutateStrategy()
	return &FuzzStrategist{
		SchemaToValueStrategy: schemaToValueStrategy,
		ResourceMutateStrategy: resourceMutateStrategy,
	}
}

// GenerateValueForSchema generates a value for a given schema.
// name is the name, type or key etc. of the value, and schema is the schema of the value.
func (s *FuzzStrategist) GenerateValueForSchema(name string, schema *openapi3.SchemaRef) (resource.Resource, error) {
	return s.SchemaToValueStrategy.GenerateValueForSchema(name, schema)
}

// MutateResource mutates a resource.
func (s *FuzzStrategist) MutateResource(resource resource.Resource) (resource.Resource, error) {
	return s.ResourceMutateStrategy.MutateResource(resource)
}
