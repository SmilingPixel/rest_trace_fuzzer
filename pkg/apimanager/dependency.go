package apimanager

// APIDependencyGraph represents the dependencies between different APIs.
// It is a map from an API method to a list of API methods that it depends on.
type APIDependencyGraph map[SimpleAPIMethod][]SimpleAPIMethod

// NewAPIDependencyGraph creates a new APIDependencyGraph.
func NewAPIDependencyGraph() *APIDependencyGraph {
	return &APIDependencyGraph{}
}
