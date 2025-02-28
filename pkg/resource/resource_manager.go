package resource

import (
	"io"
	"math/rand/v2"
	"os"
	"resttracefuzzer/pkg/static"

	"github.com/bytedance/sonic"
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
		return m.GetSingleResourceByType(static.SimpleAPIPropertyTypeFloat)
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

// LoadFromExternalDict loads resources from an external dictionary.
// The dictionary should be a json file with the following format:
//
//	[
//	    {
//	        "name": "resource1",
//	        "type": "string",
//	        "value": "value1"
//	    },
//	    {
//	        "name": "resource2",
//	        "type": "number",
//	        "value": 1.0
//	    }
//	]
//
// The type should be one of "string", "number", "integer", "boolean".
// It returns an error if any.
func (m *ResourceManager) LoadFromExternalDictFile(filePath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		log.Err(err).Msgf("[ResourceManager.LoadFromExternalDictFile] Failed to open file: %s", filePath)
		return err
	}
	defer file.Close()

	// Read file content
	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Err(err).Msgf("[ResourceManager.LoadFromExternalDictFile] Failed to read file: %s", filePath)
		return err
	}

	log.Info().Msgf("[ResourceManager.LoadFromExternalDictFile] Loading resources from file: %s", filePath)

	// Parse JSON content
	var dictValues []struct {
		Name  string      `json:"name"`
		Value interface{} `json:"value"`
	}
	err = sonic.Unmarshal(bytes, &dictValues)
	if err != nil {
		log.Error().Err(err).Msg("[ResourceManager.LoadFromExternalDictFile] Failed to unmarshal JSON")
		return err
	}

	// Populate ResourceManager maps
	succCnt := 0
	for _, dictValue := range dictValues {
		// parse value and create a new resource
		resource, err := NewResourceFromValue(dictValue.Name, dictValue.Value)
		if err != nil {
			log.Warn().Msgf("[ResourceManager.LoadFromExternalDictFile] Failed to create resource: %s, err: %v", dictValue.Name, err)
			continue
		}
		m.ResourceTypeMap[resource.Typ()] = append(m.ResourceTypeMap[resource.Typ()], resource)
		m.ResourceNameMap[dictValue.Name] = append(m.ResourceNameMap[dictValue.Name], resource)
		succCnt++
	}
	log.Info().Msgf("[ResourceManager.LoadFromExternalDictFile] Loaded %d resources", succCnt)
	return nil
}
