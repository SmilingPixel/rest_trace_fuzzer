package trace


// TraceDB represents a database for traces.
// Structs that implement this interface should be able to store and retrieve traces.
// By default, we store Jaeger-style traces.
type TraceDB interface {
	// TODO: Define the methods of the TraceDB interface. @xunzhou24

	// SelectByIDs selects traces by IDs.
	// If any trace of target ID does not exist, length of the result will be less than the length of the input.
	SelectByIDs(ids []string) ([]*SimplifiedJaegerTrace, error)

	// Upsert inserts or updates traces.
	// If the trace already exists, it will be updated.
	Upsert(traces []*SimplifiedJaegerTrace) error
}


// InMemoryTraceDB is an in-memory implementation of TraceDB.
type InMemoryTraceDB struct {

	// Traces is a list of traces.
	// TODO: performance optimization: use a better structure instead of a list. @xunzhou24
	Traces []*SimplifiedJaegerTrace
}

// NewInMemoryTraceDB creates a new InMemoryTraceDB.
func NewInMemoryTraceDB() *InMemoryTraceDB {
	return &InMemoryTraceDB{}
}

// SelectByIDs selects traces by IDs.
func (db *InMemoryTraceDB) SelectByIDs(ids []string) ([]*SimplifiedJaegerTrace, error) {
	idsSet := make(map[string]struct{})
	for _, id := range ids {
		idsSet[id] = struct{}{}
	}
	res := make([]*SimplifiedJaegerTrace, 0)
	for _, trace := range db.Traces {
		if _, ok := idsSet[trace.TraceID]; ok {
			res = append(res, trace)
		}
	}
	return res, nil
}

// Upsert inserts or updates traces.
func (db *InMemoryTraceDB) Upsert(traces []*SimplifiedJaegerTrace) error {
	for _, trace := range traces {
		queriedTrace, err := db.SelectByIDs([]string{trace.TraceID})
		if err != nil {
			return err
		}
		if len(queriedTrace) == 0 {
			db.Traces = append(db.Traces, trace)
		} else {
			// Update the trace.
			for i, t := range db.Traces {
				if t.TraceID == trace.TraceID {
					db.Traces[i] = trace
					break
				}
			}
		}
	}
	return nil
}

