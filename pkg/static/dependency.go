package static

// APIDependencyGraph represents the dependencies between different APIs.
// It is a map from an API method to a list of API methods that depend on it, i.e., mapping from producer to consumer.
// The graph is mainly used to choose a consumer API method when extending a test scenario (a request sequence).
type APIDependencyGraph struct {
	Graph map[SimpleAPIMethod][]SimpleAPIMethod
}

// NewAPIDependencyGraph creates a new APIDependencyGraph.
func NewAPIDependencyGraph() *APIDependencyGraph {
	return &APIDependencyGraph{
		Graph: make(map[SimpleAPIMethod][]SimpleAPIMethod),
	}
}

// AddDependency adds a dependency from a producer API method to a consumer API method.
func (g *APIDependencyGraph) AddDependency(producer, consumer SimpleAPIMethod) {
	if _, ok := g.Graph[producer]; !ok {
		g.Graph[producer] = make([]SimpleAPIMethod, 0)
	}
	g.Graph[producer] = append(g.Graph[producer], consumer)
}
