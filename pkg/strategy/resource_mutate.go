package strategy

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"resttracefuzzer/pkg/resource"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"
)

const (
	// StringMutateProbability is the probability of mutating for each byte in a string.
	StringMutateProbability = 0.1

	// MaxStringMutations is the maximum number of bytes to mutate in a string.
	MaxStringMutations = 5


	// MutationPlanRandom is the key for random mutation plan.
	MutationPlanRandom = "RANDOM"

	// MutationPlanStructure is the key for structure mutation plan.
	MutationPlanStructure = "STRUCTURE"

	// NoMutationPlan is the key for no mutation plan, i.e., do not mutate.
	NoMutationPlan = "NONE"
)

// ResourceMutateStrategy is a strategy for mutating resources.
// We plan to apply 2 types of mutation:
//   - Mutation of resource value
//   - Mutation of resource structure TODO @xunzhou24
type ResourceMutateStrategy struct {

	// MutationPlanWeightMap is the weight map for different mutation plans.
	// It determines whether to mutate, which type of mutation to apply.
	// It must have 3 keys (RANDOM, STRUCTURE, NONE) with non-negative integer weights.
	MutationPlanWeightMap WeightMapStrategy
}

// NewResourceMutateStrategy creates a new ResourceMutateStrategy.
// By default we use constant weight value, and the weight of random mutation, structure mutation, and no mutation are 1, 0, 3, respectively.
// If you do not want to apply structure mutation, you can set its weight to 0.
// TODO: initialize the weight map from configuration. @xunzhou24
func NewResourceMutateStrategy() *ResourceMutateStrategy {
	return &ResourceMutateStrategy{
		MutationPlanWeightMap: NewConstantWeightMapStrategy(
			map[string]int{
				MutationPlanRandom:    1,
				MutationPlanStructure: 0,
				NoMutationPlan:        3,
			},
		),
	}
}

// MutateResource mutates a resource.
// It is the entry of the mutation process.
// We will apply different mutation strategies based applied strategies.
// The method will return the mutated resource, and error if any.
// Note that the parameter resource will be mutated in place (the returned resource is the same as the parameter).
func (s *ResourceMutateStrategy) MutateResource(resrc resource.Resource) (resource.Resource, error) {
	switch resrc.Typ() {
	case static.SimpleAPIPropertyTypeObject:
		return s.mutateObjectResource(resrc)
	case static.SimpleAPIPropertyTypeArray:
		return s.mutateArrayResource(resrc)
	case static.SimpleAPIPropertyTypeInteger, static.SimpleAPIPropertyTypeFloat, static.SimpleAPIPropertyTypeBoolean, static.SimpleAPIPropertyTypeString:
		return s.mutatePrimitiveResource(resrc)
	default:
		// We do not support other types.
		return nil, fmt.Errorf("unsupported type: %v", resrc.Typ())
	}
}

// mutateObjectResource mutates an object resource.
func (s *ResourceMutateStrategy) mutateObjectResource(resrc resource.Resource) (resource.Resource, error) {
	if resrc == nil || resrc.Typ() != static.SimpleAPIPropertyTypeObject {
		return nil, fmt.Errorf("invalid object resource")
	}

	mutatedResrc, applied, err := s.precheckAndTryApplyMutationPlan(resrc)
	if err != nil {
		return nil, err
	}
	if applied {
		return mutatedResrc, nil
	}

	object := resrc.(*resource.ResourceObject).Value
	if len(object) == 0 {
		return resrc, nil
	}

	for key, value := range object {
		mutatedValue, err := s.MutateResource(value)
		if err != nil {
			return nil, err
		}
		object[key] = mutatedValue
	}
	return resrc, nil
}

// mutateArrayResource mutates an array resource.
func (s *ResourceMutateStrategy) mutateArrayResource(resrc resource.Resource) (resource.Resource, error) {
	// We do not try to apply mutation plan for array, i.e., array is not seen as a whole resource.
	// Instead, we apply mutation plan to each element in the array.

	if resrc == nil || resrc.Typ() != static.SimpleAPIPropertyTypeArray {
		return nil, fmt.Errorf("invalid array resource")
	}
	
	array := resrc.(*resource.ResourceArray).Value
	if len(array) == 0 {
		return resrc, nil
	}

	for i, value := range array {
		mutatedValue, err := s.MutateResource(value)
		if err != nil {
			return nil, err
		}
		array[i] = mutatedValue
	}
	return resrc, nil
}

// mutatePrimitiveResource mutates a primitive resource.
func (s *ResourceMutateStrategy) mutatePrimitiveResource(resrc resource.Resource) (resource.Resource, error) {
	if resrc == nil || !static.IsPrimitiveSimpleAPIPropertyType(resrc.Typ()) {
		return nil, fmt.Errorf("invalid primitive resource")
	}

	mutatedResrc, applied, err := s.precheckAndTryApplyMutationPlan(resrc)
	if err != nil {
		return nil, err
	}
	if applied {
		return mutatedResrc, nil
	}

	return resrc, nil
}

// mutatePrimitiveResourceByRandom mutates a primitive resource.
//   - For integer, float and bool, a new random value will be generated.
//   - For string, random bytes of the text will be changed.
func (s *ResourceMutateStrategy) mutatePrimitiveResourceByRandom(resrc resource.Resource) (resource.Resource, error) {
	switch resrc.Typ() {
	case static.SimpleAPIPropertyTypeInteger:
		newValue := utils.RandomValueForPrimitiveTypeKind(reflect.Int64)
		resrc.SetByRawValue(newValue)
	case static.SimpleAPIPropertyTypeFloat:
		newValue := utils.RandomValueForPrimitiveTypeKind(reflect.Float64)
		resrc.SetByRawValue(newValue)
	case static.SimpleAPIPropertyTypeBoolean:
		newValue := utils.RandomValueForPrimitiveTypeKind(reflect.Bool)
		resrc.SetByRawValue(newValue)
	case static.SimpleAPIPropertyTypeString:
		newValue := utils.MutateRandBytesForString(resrc.GetRawValue().(string), StringMutateProbability, MaxStringMutations)
		resrc.SetByRawValue(newValue)
	default:
		// We do not support other types.
		return nil, fmt.Errorf("unsupported type: %v", resrc.Typ())
	}
	return resrc, nil
}

// mutateObjectResourceStructure mutates the structure of an object resource.
// It will change the structure of the object resource, e.g., add or remove fields.
func (s *ResourceMutateStrategy) mutateObjectResourceStructure(resrc resource.Resource) (resource.Resource, error) {
	// TODO: implement the method. @xunzhou24
	return resrc, nil
}


// precheckAndTryApplyMutationPlan prechecks the resource and tries to apply the mutation plan.
// It returns:
//  - The mutated resource if the mutation plan is applied.
//  - A boolean value indicating whether the mutation plan is applied (including no mutation).
//  - An error if any.
func (s *ResourceMutateStrategy) precheckAndTryApplyMutationPlan(resrc resource.Resource) (resource.Resource, bool, error) {
	if resrc == nil {
		return nil, false, fmt.Errorf("resource is nil")
	}

	// decide the mutation plan
	// Primitive resources can only be mutated by random mutation, object resources can only be mutated by structure mutation,
	// and array resources cannot be mutated (but we should continue to check elements of the array).
	mutationPlan := s.decideMutationPlan()
	if static.IsPrimitiveSimpleAPIPropertyType(resrc.Typ()) && mutationPlan != MutationPlanRandom {
		return resrc, false, nil
	}
	if resrc.Typ() == static.SimpleAPIPropertyTypeObject && mutationPlan != MutationPlanStructure {
		return resrc, false, nil
	}
	if resrc.Typ() == static.SimpleAPIPropertyTypeArray && mutationPlan != NoMutationPlan {
		return resrc, false, nil
	}
	
	switch mutationPlan {
	case MutationPlanRandom:
		mutatedResrc, err := s.mutatePrimitiveResourceByRandom(resrc)
		if err != nil {
			return nil, false, err
		}
		return mutatedResrc, true, nil
	case MutationPlanStructure:
		mutatedResrc, err := s.mutateObjectResourceStructure(resrc)
		if err != nil {
			return nil, false, err
		}
		return mutatedResrc, true, nil
	case NoMutationPlan:
		return resrc, true, nil
	default:
		return nil, false, fmt.Errorf("unsupported mutation plan: %v", mutationPlan)
	}
}



// decideMutationPlan decides the mutation plan based on the weight map.
func (s *ResourceMutateStrategy) decideMutationPlan() string {
	totalWeight := 0
	for _, weight := range s.MutationPlanWeightMap.GetMapWithParam(WEIGHT_MAP_STRATEGY_PARAM_PLACEHOLDER) {
		totalWeight += weight
	}

	randomNumber := rand.IntN(totalWeight)
	cumulativeWeight := 0
	for source, weight := range s.MutationPlanWeightMap.GetMapWithParam(WEIGHT_MAP_STRATEGY_PARAM_PLACEHOLDER) {
		cumulativeWeight += weight
		if randomNumber < cumulativeWeight {
			return source
		}
	}

	// As a fallback, return no mutation plan. This line should normally never be reached.
	return NoMutationPlan
}
