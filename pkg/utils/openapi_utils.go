package utils

import (

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
			props := s.Value.Properties
			for propName, propScheme := range props {
				log.Info().Msgf("[flattenSchema] start to process property %s", propName)
				if propScheme.Value.Type.Includes("object") {
					newQue = append(newQue, propScheme)
					schemas[propName] = propScheme
				} else if propScheme.Value.Type.Includes("array") {
					newQue = append(newQue, propScheme.Value.Items)
					schemas[propName] = propScheme.Value.Items
				} else {
					schemas[propName] = propScheme
				}
			}
		}
		que = newQue
	}
	return schemas, nil
}
