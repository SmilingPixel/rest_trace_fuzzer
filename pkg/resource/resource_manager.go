package resource

import (
	"math/rand/v2"
	"resttracefuzzer/pkg/static"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

// Resource represents a resource in the resource pool, and ResourceManager manages the resource pool.
// The resource pool is a set of resources, and several maps are used to index the resources, all of them having consistent data.
type ResourceManager struct {

	// Value set of all map below should be the same.
	// ResourceTypeMap is a map from the property type to the list of resources.
	ResourceTypeMap map[static.SimpleAPIPropertyType][]Resource
	// ResourceNameMap is a map from the resource name to the resource.
	ResourceNameMap map[string][]Resource
}

// NewResourceManager creates a new ResourceManager.
func NewResourceManager() *ResourceManager {
	resourceTypeMap := make(map[static.SimpleAPIPropertyType][]Resource)
	resourceNameMap := make(map[string][]Resource)
	return &ResourceManager{
		ResourceTypeMap: resourceTypeMap,
		ResourceNameMap: resourceNameMap,
	}
}

// GetSingleResourceByType gets a resource from pool by the property type.
func (m *ResourceManager) GetSingleResourceByType(propertyType static.SimpleAPIPropertyType) Resource {
	resources := m.ResourceTypeMap[propertyType]
	if len(resources) == 0 {
		log.Warn().Msgf("[ResourceManager.GetRandomResourceByType] No resource of type %s", propertyType)
		return nil
	}
	return resources[rand.IntN(len(resources))]
}

// GetSingleResourceByType gets a resource from pool by the schema type(s).
func (m *ResourceManager) GetSingleResourceBySchemaType(schema *openapi3.Types) Resource {
	switch {
	case schema.Includes("string"):
		return m.GetSingleResourceByType(static.SimpleAPIPropertyTypeString)
	case schema.Includes("number"):
		return m.GetSingleResourceByType(static.SimpleAPIPropertyTypeNumber)
	case schema.Includes("integer"):
		return m.GetSingleResourceByType(static.SimpleAPIPropertyTypeInteger)
	case schema.Includes("boolean"):
		return m.GetSingleResourceByType(static.SimpleAPIPropertyTypeBoolean)
	default:
		log.Warn().Msgf("[ResourceManager.GetSingleResourceBySchemaType] No resource of schema type %v", schema)
		return nil
	}
}

// GetSingleResourceByName gets a resource from pool by the resource name.
func (m *ResourceManager) GetSingleResourceByName(resourceName string) Resource {
	resources := m.ResourceNameMap[resourceName]
	if len(resources) == 0 {
		log.Warn().Msgf("[ResourceManager.GetRandomResourceByName] No resource of name %s", resourceName)
		return nil
	}
	return resources[rand.IntN(len(resources))]
}
