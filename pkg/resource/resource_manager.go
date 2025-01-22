package resource

import (
	"math/rand/v2"
	"resttracefuzzer/pkg/static"

	"github.com/rs/zerolog/log"
)



type ResourceManager struct {

    // ResourcePool is a map from the property type to the list of resources.
    ResourcePool map[static.SimpleAPIPropertyType][]*Resource

}

// NewResourceManager creates a new ResourceManager.
func NewResourceManager() *ResourceManager {
    resourcePool := make(map[static.SimpleAPIPropertyType][]*Resource)
    return &ResourceManager{
        ResourcePool: resourcePool,
    }
}

// GetRandomResourceByType gets a random resource by the property type.
func (m *ResourceManager) GetRandomResourceByType(propertyType static.SimpleAPIPropertyType) *Resource {
    resources := m.ResourcePool[propertyType]
    if len(resources) == 0 {
        log.Warn().Msgf("[ResourceManager.GetRandomResourceByType]No resource of type %s", propertyType)
        return nil
    }
    return resources[rand.IntN(len(resources))]
}
