package parser

import (
	"os"
	"resttracefuzzer/pkg/apimanager"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

// APIDependencyRestlerParser represents a parser for API dependencies from Restler.
// It implements the APIDependencyParser interface.
type APIDependencyRestlerParser struct {
}

// NewAPIDependencyRestlerParser creates a new APIDependencyRestlerParser.
func NewAPIDependencyRestlerParser() *APIDependencyRestlerParser {
	return &APIDependencyRestlerParser{}
}

// ParseFromPath parses API dependencies from a given path.
func (p *APIDependencyRestlerParser) ParseFromPath(path string) (*apimanager.APIDependencyGraph, error) {
	data, err := os.ReadFile(path)
    if err != nil {
        log.Info().Msgf("[ParseFromPath] Error reading file: %v", err)
        return nil, err
    }

	type ProducerConsumerDetail []map[string]string
	type ParamInMap map[string]ProducerConsumerDetail
	type MethodMap map[string]ParamInMap
	type PathMap map[string]MethodMap
    var jsonMap PathMap
    if err := sonic.Unmarshal(data, &jsonMap); err != nil {
        log.Info().Msgf("[ParseFromPath] Error parsing JSON: %v", err)
        return nil, err
    }

	// jsonMap: path -> method -> paramIn -> producer_consumer_details
	// Example JSON format:
	// {
	// 	"/api/products/{productId}": {
	//     "GET": {
	//       "Path": [
	//         {
	//           "producer_endpoint": "/api/products",
	//           "producer_method": "GET",
	//           "producer_resource_name": "[0]/id",
	//           "consumer_param": "productId"
	//         }
	//       ],
	//       "Query": [
	//         {
	//           "producer_endpoint": "",
	//           "producer_method": "",
	//           "producer_resource_name": "",
	//           "consumer_param": "currencyCode"
	//         }
	//       ]
	//     }
	//   },
	//   ...
	// }

	dependencyGraph := apimanager.NewAPIDependencyGraph()
	for path, methods := range jsonMap {
		for method, paramInMap := range methods {
			for _, producerConsumerDetails := range paramInMap {
				for _, producerConsumerDetail := range producerConsumerDetails {
					if producerConsumerDetail["producer_endpoint"] == "" {
						continue
					}
					consumer := apimanager.SimpleAPIMethod{
						Endpoint: path,
						Method:   method,
					}
					producer := apimanager.SimpleAPIMethod{
						Endpoint: producerConsumerDetail["producer_endpoint"],
						Method:   producerConsumerDetail["producer_method"],
					}
					log.Info().Msgf("[ParseFromPath] Adding dependency from %v to %v", producer, consumer)
					dependencyGraph.AddDependency(consumer, producer)
				}
			}
		}
	}
	return dependencyGraph, nil
}
