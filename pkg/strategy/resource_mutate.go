package strategy

import "resttracefuzzer/pkg/resource"

// ResourceMutateStrategy is a strategy for mutating resources.
// We plan to apply 2 types of mutation:
//  - Mutation of resource value
//  - Mutation of resource structure
type ResourceMutateStrategy struct {
}

// NewResourceMutateStrategy creates a new ResourceMutateStrategy.
func NewResourceMutateStrategy() *ResourceMutateStrategy {
	return &ResourceMutateStrategy{}
}

// MutateResource mutates a resource.
// It is the entry of the mutation process.
// We will apply different mutation strategies based applied strategies.
// The method will return the mutated resource, and error if any.
func (s *ResourceMutateStrategy) MutateResource(resource resource.Resource) (resource.Resource, error) {
	// TODO: implement the mutation process. @xunzhou24
	return resource, nil
}


// mutateObjectResource mutates an object resource.
func (s *ResourceMutateStrategy) mutateObjectResource(resource resource.Resource) (resource.Resource, error) {
	// TODO: implement the mutation process. @xunzhou24
	return resource, nil
}

// mutateArrayResource mutates an array resource.
func (s *ResourceMutateStrategy) mutateArrayResource(resource resource.Resource) (resource.Resource, error) {
	// TODO: implement the mutation process. @xunzhou24
	return resource, nil
}

// mutatePrimitiveResource mutates a primitive resource.
func (s *ResourceMutateStrategy) mutatePrimitiveResource(resource resource.Resource) (resource.Resource, error) {
	// TODO: implement the mutation process. @xunzhou24
	return resource, nil
}
