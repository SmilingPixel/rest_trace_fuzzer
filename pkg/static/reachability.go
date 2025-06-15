package static


import "slices"

// ReachabilityMap is a map of reachability information.
// It maps an external API to a list of internal APIs that are reachable from it, and vice versa.
// Note: key of the map is not simple types, but a struct, so ReachabilityMap cannot be json.Marshal
type ReachabilityMap struct {
	// External to Internal reachability
	External2Internal map[SimpleAPIMethod][]InternalServiceEndpoint

	// Internal to External reachability
	Internal2External map[InternalServiceEndpoint][]SimpleAPIMethod

}

// NewReachabilityMap creates a new ReachabilityMap.
func NewReachabilityMap() *ReachabilityMap {
	external2Internal := make(map[SimpleAPIMethod][]InternalServiceEndpoint)
	internal2External := make(map[InternalServiceEndpoint][]SimpleAPIMethod)
	return &ReachabilityMap{
		External2Internal: external2Internal,
		Internal2External: internal2External,
	}
}

// AddReachability adds reachability information to the map.
func (r *ReachabilityMap) AddReachability(external SimpleAPIMethod, internal InternalServiceEndpoint) {
	// Add external to internal reachability
	r.External2Internal[external] = append(r.External2Internal[external], internal)

	// Add internal to external reachability
	r.Internal2External[internal] = append(r.Internal2External[internal], external)
}

// RemoveReachability removes reachability information from the map.
func (r *ReachabilityMap) RemoveReachability(external SimpleAPIMethod, internal InternalServiceEndpoint) {
	// Remove external to internal reachability
	if internalAPIs, ok := r.External2Internal[external]; ok {
		for i, api := range internalAPIs {
			if api == internal {
				r.External2Internal[external] = slices.Delete(internalAPIs, i, i+1)
				break
			}
		}
	}

	// Remove internal to external reachability
	if externalAPIs, ok := r.Internal2External[internal]; ok {
		for i, api := range externalAPIs {
			if api == external {
				r.Internal2External[internal] = slices.Delete(externalAPIs, i, i+1)
				break
			}
		}
	}
}

// GetInternalsByExternal returns the list of internal APIs that are reachable from the given external API.
// It returns the list of internal APIs and a boolean indicating whether the external API is found in the map.
func (r *ReachabilityMap) GetInternalsByExternal(external SimpleAPIMethod) ([]InternalServiceEndpoint, bool) {
	if internal, ok := r.External2Internal[external]; ok {
		return internal, true
	}
	return nil, false
}

// GetExternalsByInternal returns the list of external APIs that are reachable from the given internal API.
// It returns the list of external APIs and a boolean indicating whether the internal API is found in the map.
func (r *ReachabilityMap) GetExternalsByInternal(internal InternalServiceEndpoint) ([]SimpleAPIMethod, bool) {
	if external, ok := r.Internal2External[internal]; ok {
		return external, true
	}
	return nil, false
}