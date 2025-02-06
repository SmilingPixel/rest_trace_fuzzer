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
	schemas := make(map[string]*openapi3.SchemaRef)
	if schema == nil {
		log.Info().Msg("Schema is nil")
		return schemas, nil
	}
	// schemas = append(schemas, schema)

	// BFS
	que := make([]*openapi3.SchemaRef, 0)
	que = append(que, schema)
	for len(que) > 0 {
		newQue := make([]*openapi3.SchemaRef, 0)
		for _, s := range que {
			switch {
			case s.Value.Type.Includes("object"):
				for propName, propSchema := range s.Value.Properties {
					newQue = append(newQue, propSchema)
					schemas[propName] = propSchema
				}
			case s.Value.Type.Includes("array"):
				newQue = append(newQue, s.Value.Items)
				schemas[s.Value.Title] = s.Value.Items
			default:
				schemas[s.Value.Title] = s
			}

		}
		que = newQue
	}
	return schemas, nil
}

// IncludePrimitiveType checks if the types include primitive types.
func IncludePrimitiveType(types *openapi3.Types) bool {
	return types.Includes("string") || types.Includes("number") || types.Includes("integer") || types.Includes("boolean")
}

// GenerateDefaultValueForPrimitiveSchemaType generates a default value for a primitive schema type.
func GenerateDefaultValueForPrimitiveSchemaType(schemaType *openapi3.Types) interface{} {
	log.Debug().Msgf("[GenerateDefaultValueForPrimitiveSchemaType] schemaType: %v", schemaType)
	switch {
	case schemaType.Includes("string"):
		return "114-514"
	case schemaType.Includes("number"):
		return 114.514
	case schemaType.Includes("integer"):
		return 114514
	case schemaType.Includes("boolean"):
		return true
	default:
		return nil
	}
}

// GenerateRandomValueForPrimitiveSchemaType generates a random value for a primitive schema type.
func GenerateRandomValueForPrimitiveSchemaType(schemaType *openapi3.Types) interface{} {
	log.Debug().Msgf("[GenerateRandomValueForPrimitiveSchemaType] schemaType: %v", schemaType)
	switch {
	case schemaType.Includes("string"):
		randLength := rand.IntN(114) + 1
		return RandStringBytes(randLength)
	case schemaType.Includes("number"):
		return rand.Float64() + float64(rand.IntN(114514))
	case schemaType.Includes("integer"):
		return rand.IntN(114514)
	case schemaType.Includes("boolean"):
		return rand.IntN(2) == 1
	default:
		return nil
	}
}
