package utils

import (
	"math/rand/v2"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
)

// flattenSchema flattens a schema to a list of schemas.
// It returns a map from the schema name to the schema.
//
// TODO: support openapi3 oneOf, anyOf, allOf, etc. @xunzhou24
func FlattenSchema(schema *openapi3.SchemaRef) (map[string]*openapi3.SchemaRef, error) {
	name2schema := make(map[string]*openapi3.SchemaRef)
	if schema == nil {
		log.Info().Msg("Schema is nil")
		return name2schema, nil
	}
	// schemas = append(schemas, schema)

	type schemaQueueItem struct {
		name   string
		schema *openapi3.SchemaRef
	}

	// BFS
	que := make([]schemaQueueItem, 0)
	que = append(que, schemaQueueItem{name: schema.Ref, schema: schema})
	for len(que) > 0 {
		newQue := make([]schemaQueueItem, 0)
		for _, s := range que {
			switch {
			case s.schema.Value.Type.Includes(openapi3.TypeObject):
				for propName, propSchema := range s.schema.Value.Properties {
					newQue = append(newQue, schemaQueueItem{name: propName, schema: propSchema})
					name2schema[propName] = propSchema
				}
			case s.schema.Value.Type.Includes(openapi3.TypeArray):
				// Array element would not be seen as a whole,
				// so we do not store array itself, just flatten it instead.
				newQue = append(newQue, schemaQueueItem{name: s.name, schema: s.schema.Value.Items})
			default:
				if s.name != "" {
					name2schema[s.name] = s.schema
				}
			}
		}
		que = newQue
	}
	return name2schema, nil
}

// IncludePrimitiveType checks if the types include primitive types.
func IncludePrimitiveType(types *openapi3.Types) bool {
	return types.Includes(openapi3.TypeString) || types.Includes(openapi3.TypeNumber) || types.Includes(openapi3.TypeInteger) || types.Includes(openapi3.TypeBoolean)
}

// GenerateDefaultValueForPrimitiveSchemaType generates a default value for a primitive schema type.
func GenerateDefaultValueForPrimitiveSchemaType(schemaType *openapi3.Types) interface{} {
	log.Debug().Msgf("[GenerateDefaultValueForPrimitiveSchemaType] schemaType: %v", schemaType)
	switch {
	case schemaType.Includes(openapi3.TypeString):
		return "114-514"
	case schemaType.Includes(openapi3.TypeNumber):
		return 114.514
	case schemaType.Includes(openapi3.TypeInteger):
		return 114514
	case schemaType.Includes(openapi3.TypeBoolean):
		return true
	default:
		return nil
	}
}

// GenerateRandomValueForPrimitiveSchemaType generates a random value for a primitive schema type.
func GenerateRandomValueForPrimitiveSchemaType(schemaType *openapi3.Types) interface{} {
	log.Debug().Msgf("[GenerateRandomValueForPrimitiveSchemaType] schemaType: %v", schemaType)
	switch {
	case schemaType.Includes(openapi3.TypeString):
		randLength := rand.IntN(114) + 1
		return RandStringBytes(randLength)
	case schemaType.Includes(openapi3.TypeNumber):
		return rand.Float64() + float64(rand.IntN(114514))
	case schemaType.Includes(openapi3.TypeInteger):
		return rand.IntN(114514)
	case schemaType.Includes(openapi3.TypeBoolean):
		return rand.IntN(2) == 1
	default:
		return nil
	}
}
