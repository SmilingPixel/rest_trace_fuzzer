package trace

// TraceDB represents a database for traces.
// Structs that implement this interface should be able to store and retrieve traces.
type TraceDB interface {

	// SelectByIDs selects traces by IDs.
	// If any trace of target ID does not exist, length of the result will be less than the length of the input.
	SelectByIDs(ids []string) ([]*SimplifiedTrace, error)

	// BatchUpsert inserts or updates traces.
	// If the trace already exists, it will be updated.
	BatchUpsert(traces []*SimplifiedTrace) error

	// InsertAndReturn inserts a trace and returns the inserted trace.
	// If the trace already exists, it will not be inserted.
	InsertAndReturn(trace *SimplifiedTrace) (*SimplifiedTrace, error)

	// BatchInsertAndReturn inserts traces and returns the inserted traces.
	BatchInsertAndReturn(traces []*SimplifiedTrace) ([]*SimplifiedTrace, error)
}

// InMemoryTraceDB is an in-memory implementation of TraceDB.
type InMemoryTraceDB struct {

	// Traces is a list of traces.
	// TODO: performance optimization: use a better structure instead of a list. @xunzhou24
	Traces []*SimplifiedTrace
}

// NewInMemoryTraceDB creates a new InMemoryTraceDB.
func NewInMemoryTraceDB() *InMemoryTraceDB {
	return &InMemoryTraceDB{}
}

// SelectByIDs selects traces by IDs.
func (db *InMemoryTraceDB) SelectByIDs(ids []string) ([]*SimplifiedTrace, error) {
	idsSet := make(map[string]struct{})
	for _, id := range ids {
		idsSet[id] = struct{}{}
	}
	res := make([]*SimplifiedTrace, 0)
	for _, trace := range db.Traces {
		if _, ok := idsSet[trace.TraceID]; ok {
			res = append(res, trace)
		}
	}
	return res, nil
}

// BatchUpsert inserts or updates traces.
func (db *InMemoryTraceDB) BatchUpsert(traces []*SimplifiedTrace) error {
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

// BatchInsertAndReturn inserts traces and returns the inserted traces.
func (db *InMemoryTraceDB) BatchInsertAndReturn(traces []*SimplifiedTrace) ([]*SimplifiedTrace, error) {
	newlyInsertedTraces := make([]*SimplifiedTrace, 0)
	for _, trace := range traces {
		exist, err := db.SelectByIDs([]string{trace.TraceID})
		if err != nil {
			return nil, err
		}
		if len(exist) > 0 {
			continue
		}
		db.Traces = append(db.Traces, trace)
		newlyInsertedTraces = append(newlyInsertedTraces, trace)
	}
	return newlyInsertedTraces, nil
}

// InsertAndReturn inserts a trace and returns the inserted trace.
func (db *InMemoryTraceDB) InsertAndReturn(trace *SimplifiedTrace) (*SimplifiedTrace, error) {
	exist, err := db.SelectByIDs([]string{trace.TraceID})
	if err != nil {
		return nil, err
	}
	if len(exist) > 0 {
		return nil, nil
	}
	db.Traces = append(db.Traces, trace)
	return trace, nil
}
