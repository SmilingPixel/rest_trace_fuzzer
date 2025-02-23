package strategy

const WEIGHT_MAP_STRATEGY_PARAM_PLACEHOLDER = -1

// WeightMapStrategy defines the interface for different weight map strategies.
type WeightMapStrategy interface {
    // GetWeight returns the weight value for a given key.
    GetWeight(key string) int

    // GetWeightWithParam returns the weight value for a given key and function parameter.
    GetWeightWithParam(key string, param int) int

	// GetMapWithParam returns the weight map with a function parameter.
	GetMapWithParam(param int) map[string]int
}

// ConstantWeightMapStrategy is a strategy for constant weight maps.
type ConstantWeightMapStrategy struct {
    weights map[string]int
}

// NewConstantWeightMapStrategy creates a new ConstantWeightMapStrategy.
func NewConstantWeightMapStrategy(weights map[string]int) *ConstantWeightMapStrategy {
    return &ConstantWeightMapStrategy{weights: weights}
}

// GetWeight returns the weight value for a given key.
func (s *ConstantWeightMapStrategy) GetWeight(key string) int {
    return s.weights[key]
}

// GetWeightWithParam returns the weight value for a given key and function parameter.
// For a constant weight map, the parameter is ignored, and it's recommended to use WEIGHT_MAP_STRATEGY_PARAM_PLACEHOLDER as the placeholder.
func (s *ConstantWeightMapStrategy) GetWeightWithParam(key string, param int) int {
    return s.GetWeight(key)
}

// GetMapWithParam returns the weight map with a function parameter.
// For a constant weight map, the parameter is ignored, and it's recommended to use WEIGHT_MAP_STRATEGY_PARAM_PLACEHOLDER as the placeholder.
func (s *ConstantWeightMapStrategy) GetMapWithParam(param int) map[string]int {
	return s.weights
}

// VariableWeightMapStrategy is a strategy for variable weight maps.
type VariableWeightMapStrategy struct {
    weights map[string]func(int) int
}

// NewVariableWeightMapStrategy creates a new VariableWeightMapStrategy.
func NewVariableWeightMapStrategy(weights map[string]func(int) int) *VariableWeightMapStrategy {
    return &VariableWeightMapStrategy{weights: weights}
}

// GetWeight returns the weight value for a given key.
func (s *VariableWeightMapStrategy) GetWeight(key string) int {
    if weightFunc, exists := s.weights[key]; exists {
        return weightFunc(1) // Default parameter value
    }
    return 0
}

// GetWeightWithParam returns the weight value for a given key and function parameter.
func (s *VariableWeightMapStrategy) GetWeightWithParam(key string, param int) int {
    if weightFunc, exists := s.weights[key]; exists {
        return weightFunc(param)
    }
    return 0
}

// GetMapWithParam returns the weight map with a function parameter.
func (s *VariableWeightMapStrategy) GetMapWithParam(param int) map[string]int {
	weightMap := make(map[string]int)
	for key, weightFunc := range s.weights {
		weightMap[key] = weightFunc(param)
	}
	return weightMap
}
