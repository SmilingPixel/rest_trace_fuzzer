package static

// APIDependencyGraph represents the dependencies between different APIs.
// It is a map from an API method to a list of API methods that it depends on.
type APIDependencyGraph map[SimpleAPIMethod][]SimpleAPIMethod

// NewAPIDependencyGraph creates a new APIDependencyGraph.
func NewAPIDependencyGraph() *APIDependencyGraph {
	return &APIDependencyGraph{}
}

// AddDependency adds a dependency from a producer API method to a consumer API method.
func (g *APIDependencyGraph) AddDependency(producer, consumer SimpleAPIMethod) {
	if _, ok := (*g)[producer]; !ok {
		(*g)[producer] = make([]SimpleAPIMethod, 0)
	}
	(*g)[producer] = append((*g)[producer], consumer)
}
