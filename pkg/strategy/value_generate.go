package strategy

import (
	"fmt"
	"math/rand/v2"
	"resttracefuzzer/pkg/resource"
	"resttracefuzzer/pkg/utils"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

const (
	// VALUE_SOURCE_RANDOM is the key for random value source.
	VALUE_SOURCE_RANDOM = "RANDOM"

	// VALUE_SOURCE_RESOURCE_POOL is the key for value from resource pool.
	VALUE_SOURCE_RESOURCE_POOL = "RESOURCE_POOL"

	// VALUE_SOURCE_MUTATION is the key for mutation of values.
	VALUE_SOURCE_MUTATION = "MUTATION"
)

// SchemaToValueStrategy is a strategy for generating values from schemas.
// It uses 3 kinds of strategies:
//  1. Random value, only applicable to primitive types.
//  2. Value from resource pool, including values from dictionary and test case response.
//  3. Mutation of values from 1 and 2.
//
// You can control the strategy by setting the configuration. At present you can set:
//  1. The ratio of random value, value from resource pool, and mutation.
type SchemaToValueStrategy struct {

	// ResourceManager is the resource manager for fetching resources.
	ResourceManager *resource.ResourceManager

	// ValueSourceWeightMap is the weight map for different value sources.
	// It can use different strategies to determine the weight of each value source.
	// It must have 3 keys (RANDOM, RESOURCE_POOL, MUTATION) with non-negative integer weights.
	ValueSourceWeightMap WeightMapStrategy
}

// NewSchemaToValueStrategy creates a new SchemaToValueStrategy.
// By default we use constant weight value, and the weight of random value, value from resource pool, and mutation are all 1.
// If you do not want resource pool or mutation to interfere, you can set their weight to 0.
// TODO: initialize the weight map from configuration. @xunzhou24
func NewSchemaToValueStrategy(resourceManager *resource.ResourceManager) *SchemaToValueStrategy {
	valueSourceWeightMap := NewConstantWeightMapStrategy(
		map[string]int{
			VALUE_SOURCE_RANDOM:       1,
			VALUE_SOURCE_RESOURCE_POOL: 1,
			VALUE_SOURCE_MUTATION:     1,
		},
	)
	return &SchemaToValueStrategy{
		ResourceManager:      resourceManager,
		ValueSourceWeightMap: valueSourceWeightMap,
	}
}

// GenerateValueForSchema generates a resource value for a given schema.
func (s *SchemaToValueStrategy) GenerateValueForSchema(schema *openapi3.SchemaRef) (resource.Resource, error) {
	// Try to apply value source.
	value, generated, err := s.preCheckAndTryApplyValueSource(schema)
	if err != nil {
		return nil, err
	}
	if generated {
		return value, nil
	}	
	
	if schema == nil || schema.Value == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	switch {
	case schema.Value.Type.Includes("object"):
		return s.generateObjectValueForSchema(schema)
	case schema.Value.Type.Includes("array"):
		return s.generateArrayValueForSchema(schema)
	default:
		return s.generatePrimitiveValueForSchema(schema)
	}
}

// generateObjectValueForSchema generates a json object resource value from a schema.
// It returns a json object resource, and error if any.
// The returned object is of type ResourceObject.
func (s *SchemaToValueStrategy) generateObjectValueForSchema(schema *openapi3.SchemaRef) (resource.Resource, error) {
	// Try to apply value source.
	value, generated, err := s.preCheckAndTryApplyValueSource(schema)
	if err != nil {
		return nil, err
	}
	if generated {
		return value, nil
	}	
	
	if schema == nil || schema.Value == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	result := resource.NewResourceObject(make(map[string]resource.Resource))

	for propName, propSchema := range schema.Value.Properties {
		propValue, err := s.GenerateValueForSchema(propSchema)
		if err != nil {
			return nil, err
		}
		result.Value[propName] = propValue
	}
	return result, nil
}

// generateArrayValueForSchema generates a json array resource value from a schema.
// It returns a json array resource, and error if any.
// The returned array is of type *ResourceArray.
func (s *SchemaToValueStrategy) generateArrayValueForSchema(schema *openapi3.SchemaRef) (resource.Resource, error) {
	// Try to apply value source.
	value, generated, err := s.preCheckAndTryApplyValueSource(schema)
	if err != nil {
		return nil, err
	}
	if generated {
		return value, nil
	}

	if schema == nil || schema.Value == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	result := resource.NewResourceArray(make([]resource.Resource, 0))

	// TODO: control the array size @xunzhou24
	// For now, we generate an array with one element.
	elementValue, err := s.GenerateValueForSchema(schema.Value.Items)
	if err != nil {
		return nil, err
	}
	result.Value = append(result.Value, elementValue)

	return result, nil
}

// generatePrimitiveValueForSchema generates a primitive resource value from a schema.
// It returns a primitive value resource, and error if any.
// The returned value is of type *ResourceNumber, *ResourceString, etc.
func (s *SchemaToValueStrategy) generatePrimitiveValueForSchema(schema *openapi3.SchemaRef) (resource.Resource, error) {
	// Try to apply value source.
	value, generated, err := s.preCheckAndTryApplyValueSource(schema)
	if err != nil {
		return nil, err
	}
	if generated {
		return value, nil
	}

	if schema == nil || schema.Value == nil {
		return nil, fmt.Errorf("schema is nil")
	}
	defaultValue := utils.GenerateDefaultValueForPrimitiveSchemaType(schema.Value.Type)
	result, err := resource.NewResourceFromValue("", defaultValue)
	if err != nil {
		return nil, err
	}
	return result, nil
}


// preCheckAndTryApplyValueSource checks the schema and applies the value source.
// It returns:
//  1. The generated value, if successful.
//  2. A boolean indicating whether the value is generated, if successful.
//  3. An error, if any.
// The method is inserted into each of the generate methods.
func (s *SchemaToValueStrategy) preCheckAndTryApplyValueSource(schema *openapi3.SchemaRef) (resource.Resource, bool, error) {
	if schema == nil || schema.Value == nil {
		return nil, false, fmt.Errorf("schema is nil")
	}

	// Decide the value source based on weights.
	valueSource := s.decideValueSource()
	switch valueSource {
	case VALUE_SOURCE_RANDOM:
		// random can only apply to primitive types
		if !utils.IncludePrimitiveType(schema.Value.Type) {
			return nil, false, nil
		}
		randomValue := utils.GenerateRandomValueForPrimitiveSchemaType(schema.Value.Type)
		result, err := resource.NewResourceFromValue("", randomValue)
		if err != nil {
			return nil, false, err
		}
		return result, true, nil
	case VALUE_SOURCE_RESOURCE_POOL:
		resource := s.ResourceManager.GetSingleResourceBySchemaType(schema.Value.Type)
		// resource of a specific type is not found
		if resource == nil {
			return nil, false, nil
		}
		return resource, true, nil
	case VALUE_SOURCE_MUTATION: // TODO: implement mutation @xunzhou24
		return nil, false, nil
	default:
		return nil, false, fmt.Errorf("unknown value source: %s", valueSource)
	}
}

// decideValueSource returns the selected value source based on weights.
func (s *SchemaToValueStrategy) decideValueSource() string {
	totalWeight := 0
	for _, weight := range s.ValueSourceWeightMap.GetMapWithParam(WEIGHT_MAP_STRATEGY_PARAM_PLACEHOLDER) {
		totalWeight += weight
	}

	randomNumber := rand.IntN(totalWeight)
	cumulativeWeight := 0
	for source, weight := range s.ValueSourceWeightMap.GetMapWithParam(WEIGHT_MAP_STRATEGY_PARAM_PLACEHOLDER) {
		cumulativeWeight += weight
		if randomNumber < cumulativeWeight {
			return source
		}
	}

	// As a fallback, return a default source. This line should normally never be reached.
	log.Warn().Msgf("[SchemaToValueStrategy.DecideValueSource] Fallback to default value source (RANDOM)")
	return VALUE_SOURCE_RANDOM
}
