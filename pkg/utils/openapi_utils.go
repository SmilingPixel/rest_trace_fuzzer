package utils

import (
	"fmt"

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

// GenerateJsonTemplateFromSchema generates a JSON template from a schema.
// It returns a json object.
//
// For primitive types, the method fills a default value.
//
// Deprecated: Use [resttracefuzzer/pkg/casemanager.PopulateCaseOperation] instead.
func GenerateJsonTemplateFromSchema(schema *openapi3.SchemaRef) (map[string]interface{}, error) {
	if schema == nil || schema.Value == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	result := make(map[string]interface{})

	for propName, propSchema := range schema.Value.Properties {
		switch {
		case propSchema.Value.Type.Includes("object"):
			subResult, err := GenerateJsonTemplateFromSchema(propSchema)
			if err != nil {
				return nil, err
			}
			result[propName] = subResult

		case propSchema.Value.Type.Includes("array"):
			subResult, err := GenerateJsonTemplateFromSchema(propSchema.Value.Items)
			if err != nil {
				return nil, err
			}
			// TODO: control the array size @xunzhou24
			result[propName] = []interface{}{subResult}

		default:
			// primitive types
			result[propName] = GenerateDefaultValueForPrimitiveSchemaType(propSchema.Value.Type)
		}
	}

	return result, nil
}

// GenerateDefaultValueForPrimitiveSchemaType generates a placeholder value for a primitive schema type.
//
// TODO: deprecate it when strategy-based generation is implemented. @xunzhou24
// TODO: support multiple types, see https://swagger.io/docs/specification/v3_0/describing-parameters/ @xunzhou24
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
