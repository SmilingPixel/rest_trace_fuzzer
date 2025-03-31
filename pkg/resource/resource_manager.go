package resource

import (
	"io"
	"math/rand/v2"
	"os"
	"resttracefuzzer/pkg/static"
	"resttracefuzzer/pkg/utils"

	"github.com/bytedance/sonic/decoder"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

// Resource represents a resource in the resource pool, and ResourceManager manages the resource pool.
// The resource pool is a set of resources, and several maps are used to index the resources, all of them having consistent data.
// To improve readability, only ResourceNameMap would be serialized.
type ResourceManager struct {

	// Value set of all map below should be the same.
	// ResourceTypeMap is a map from the property type to the list of resources.
	// Note: ResourceTypeMap only stores resources of primitive types.
	ResourceTypeMap map[static.SimpleAPIPropertyType][]Resource `json:"-"`

	// ResourceNameMap is a map from the resource name to the resource.
	ResourceNameMap map[string][]Resource `json:"resourceNameMap"`

	// ResourceName2HashSet is used to store the hashcode of resources, preventing duplicate resources.
	// It maps resource name to resource set, i.e., we do not allow duplicate resources with the same name.
	ResourceName2HashSet map[string]map[uint64]struct{} `json:"-"`
}

// NewResourceManager creates a new ResourceManager.
func NewResourceManager() *ResourceManager {
	resourceTypeMap := make(map[static.SimpleAPIPropertyType][]Resource)
	resourceNameMap := make(map[string][]Resource)
	resourceHashSet := make(map[string]map[uint64]struct{})
	return &ResourceManager{
		ResourceTypeMap: resourceTypeMap,
		ResourceNameMap: resourceNameMap,
		ResourceName2HashSet: resourceHashSet,
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

// GetSingleResourceBySchemaTypes gets a resource from pool by the schema type(s).
// Only supports primitive types: string, number, integer, boolean.
func (m *ResourceManager) GetSingleResourceBySchemaTypes(schemaTypes *openapi3.Types) Resource {
	switch {
	case schemaTypes.Includes(openapi3.TypeString):
		return m.GetSingleResourceByType(static.SimpleAPIPropertyTypeString)
	case schemaTypes.Includes(openapi3.TypeNumber):
		return m.GetSingleResourceByType(static.SimpleAPIPropertyTypeFloat)
	case schemaTypes.Includes(openapi3.TypeInteger):
		return m.GetSingleResourceByType(static.SimpleAPIPropertyTypeInteger)
	case schemaTypes.Includes(openapi3.TypeBoolean):
		return m.GetSingleResourceByType(static.SimpleAPIPropertyTypeBoolean)
	default:
		// log.Warn().Msgf("[ResourceManager.GetSingleResourceBySchemaTypes] No resource of schema type %v", schemaTypes)
		return nil
	}
}

// GetSingleResourceByName gets a resource from pool by the resource name.
// Heuristic rules are applied to get the resource whose name is similar to the given resource name.
// See the implementation of GetSingleResourceByName for details.
func (m *ResourceManager) GetSingleResourceByName(resourceName string) Resource {
	// We do not support an empty resource name.
	if resourceName == "" {
		return nil
	}

	// try to find a resource by full name
	resources := m.ResourceNameMap[resourceName]
	if len(resources) > 0 {
		return resources[rand.IntN(len(resources))]
	}

	// try to find a resource that matches in the last part of the name
	// For example, if the resource name is userName, we can get the resource by "name".
	resourceNameParts := utils.SplitIntoWords(resourceName)
	resources = m.ResourceNameMap[resourceNameParts[len(resourceNameParts)-1]]
	if len(resources) > 0 {
		return resources[rand.IntN(len(resources))]
	}

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
//	        "value": "value1"
//	    },
//	    {
//	        "name": "resource2",
//	        "value": 1.0
//	    }
//	]
//
// It returns an error if any.
// Note: for resources loaded from external dictionary, we do not store sub-resources.
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
	// To parse integer values as int64, we need to use the decoder, and set via decoder.UseInt64().
	var dictValues []struct {
		Name  string      `json:"name"`
		Value interface{} `json:"value"`
	}
	decoder := decoder.NewDecoder(string(bytes))
	decoder.UseInt64()
	err = decoder.Decode(&dictValues)
	if err != nil {
		log.Error().Err(err).Msg("[ResourceManager.LoadFromExternalDictFile] Failed to decode JSON")
		return err
	}

	// Populate ResourceManager maps
	succCnt := 0
	for _, dictValue := range dictValues {
		// parse value and create a new resource
		resourceName := dictValue.Name
		resourceValue := dictValue.Value
		resource, err := NewResourceFromValue(resourceValue)
		if err != nil {
			log.Warn().Msgf("[ResourceManager.LoadFromExternalDictFile] Failed to create resource: %s, err: %v", resourceName, err)
			continue
		}
		m.storeResource(resource, resourceName, false) // For resources loaded from external dictionary, we do not store sub-resources.
		succCnt++
	}
	log.Info().Msgf("[ResourceManager.LoadFromExternalDictFile] Loaded %d resources", succCnt)
	return nil
}

// StoreResourcesFromRawObjectBytes stores resources from raw object bytes.
// It returns an error if any.
// The raw object bytes should be a JSON object.
// The root resource name is the name of the root resource. It would be ignored if it is empty.
//
// Parameter `shouldStoreSubResources` indicates whether to store sub-resources.
// For example, if the raw object is:
//
//	{
//	    "name": "hi1",
//	    "value": {
//	        "name2": "hi2"
//	    }
//	}
//
// If `shouldStoreSubResources` is true, the resource "hi2" with name "name2" will be stored.
// In specific:
//   - for object type, all values from the object key-value pairs will be stored;
//   - for array type, all elements in the array will be stored.
func (m *ResourceManager) StoreResourcesFromRawObjectBytes(rawObjectBytes []byte, rootResourceName string, shouldStoreSubResources bool) error {
	// To parse integer values as int64, we need to use the decoder, and set via decoder.UseInt64().
	var jsonObject interface{}
	decoder := decoder.NewDecoder(string(rawObjectBytes))
	decoder.UseInt64()
	err := decoder.Decode(&jsonObject)
	if err != nil {
		log.Err(err).Msg("[ResourceManager.StoreResourcesFromRawObjectBytes] Failed to unmarshal JSON")
		return err
	}
	// Parse the object into a resource, for the convenience of post-processing.
	rootResource, err := NewResourceFromValue(jsonObject)
	if err != nil {
		log.Err(err).Msg("[ResourceManager.StoreResourcesFromRawObjectBytes] Failed to create resource from JSON object")
		return err
	}

	// Store the root resource.
	m.storeResource(rootResource, rootResourceName, shouldStoreSubResources)
	return nil
}

// storeResource stores a resource in the resource manager.
// If the resource name is not empty, it will not be stored in the resource name map, i.e., we cannot get it by name.
// Parameter `shouldStoreSubResources` indicates whether to store sub-resources.
// For example, if the raw object is:
//
//	{
//	    "name": "hi1",
//	    "value": {
//	        "name2": "hi2"
//	    }
//	}
//
// If `shouldStoreSubResources` is true, the resource "hi2" with name "name2" will be stored.
// In specific:
//   - for object type, all values from the object key-value pairs will be stored (resource name is the key);
//   - for array type, all elements in the array will be stored (heuristic rules are applied to current `resourceName` to get the name, e.g., "names" -> "name").
func (m *ResourceManager) storeResource(resource Resource, resourceName string, shouldStoreSubResources bool) {
	if isResourceEmpty(resource) {
		log.Warn().Msg("[ResourceManager.storeResource] Resource is empty")
		return
	}

	// Check if the resource is duplicate.
	resourceSet := m.ResourceName2HashSet[resourceName]
	if resourceSet == nil {
		resourceSet = make(map[uint64]struct{})
		m.ResourceName2HashSet[resourceName] = resourceSet
	}
	hashcode := resource.Hashcode()
	if _, ok := resourceSet[hashcode]; ok {
		return
	}
	resourceSet[hashcode] = struct{}{}

	// Store the resource in the resource manager.
	m.ResourceTypeMap[resource.Typ()] = append(m.ResourceTypeMap[resource.Typ()], resource)
	if resourceName != "" {
		m.ResourceNameMap[resourceName] = append(m.ResourceNameMap[resourceName], resource)
	}

	if !shouldStoreSubResources {
		return
	}
	switch resource.Typ() {
	case static.SimpleAPIPropertyTypeObject:
		for field, subResource := range resource.(*ResourceObject).Value {
			m.storeResource(subResource, field, shouldStoreSubResources)
		}
	case static.SimpleAPIPropertyTypeArray:
		// Heuristic rules to get the name of the array elements.
		arrayElementName := utils.GetArrayElementNameHeuristic(resourceName)
		for _, subResource := range resource.(*ResourceArray).Value {
			m.storeResource(subResource, arrayElementName, shouldStoreSubResources)
		}
	default:
		// Do nothing for primitive types.
	}
}

//isResourceEmpty checks if the resource is empty.
func isResourceEmpty(resource Resource) bool {
	if resource == nil {
		return true
	}
	switch resource.Typ() {
	case static.SimpleAPIPropertyTypeString:
		return resource.(*ResourceString).Value == ""
	case static.SimpleAPIPropertyTypeObject:
		return len(resource.(*ResourceObject).Value) == 0
	case static.SimpleAPIPropertyTypeArray:
		return len(resource.(*ResourceArray).Value) == 0
	default:
		return false
	}
}
