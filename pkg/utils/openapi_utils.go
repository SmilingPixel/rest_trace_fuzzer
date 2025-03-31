package utils

import (
	"reflect"

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

// PrimitiveSchemaType2ReflectKind converts a primitive schema type to a reflect kind.
func PrimitiveSchemaType2ReflectKind(schemaType *openapi3.Types) reflect.Kind {
	log.Debug().Msgf("[PrimitiveSchemaType2ReflectKind] schemaType: %v", schemaType)
	switch {
	case schemaType.Includes(openapi3.TypeString):
		return reflect.String
	case schemaType.Includes(openapi3.TypeNumber):
		return reflect.Float64
	case schemaType.Includes(openapi3.TypeInteger):
		return reflect.Int64
	case schemaType.Includes(openapi3.TypeBoolean):
		return reflect.Bool
	default:
		return 0
	}
}
