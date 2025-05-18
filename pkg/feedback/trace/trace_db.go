package trace

import (
	"fmt"
	"os"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

// TraceDB represents a database for traces.
// Structs that implement this interface should be able to store and retrieve traces.
type TraceDB interface {

	// SelectByIDs selects traces by IDs.
	// If any trace of target ID does not exist, length of the result will be less than the length of the input.
	SelectByIDs(ids []string) ([]*SimplifiedTrace, error)

	// Upsert inserts or updates a trace.
	// If the trace already exists, it will be updated.
	Upsert(trace *SimplifiedTrace) error

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

// Upsert inserts or updates a trace.
func (db *InMemoryTraceDB) Upsert(trace *SimplifiedTrace) error {
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
	return nil
}

// BatchUpsert inserts or updates traces.
func (db *InMemoryTraceDB) BatchUpsert(traces []*SimplifiedTrace) error {
	for _, trace := range traces {
		if err := db.Upsert(trace); err != nil {
			return err
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

// RawTraceFileSaver is a file-based implementation of TraceDB.
// It saves traces to files in a specified directory.
type RawTraceFileSaver struct {
	// DirPath is the directory path where traces are saved.
	DirPath string
}

// NewRawTraceFileSaver creates a new RawTraceFileSaver.
func NewRawTraceFileSaver(dirPath string) *RawTraceFileSaver {
	// Create the directory if it does not exist.
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			log.Err(err).Msgf("[NewRawTraceFileSaver] Failed to create directory: %s", err)
		}
	}
	return &RawTraceFileSaver{
		DirPath: dirPath,
	}
}

// InsertAndReturn inserts a trace and returns the inserted trace.
func (s *RawTraceFileSaver) InsertAndReturn(trace *SimplifiedTrace) (*SimplifiedTrace, error) {
	if trace == nil {
		return nil, fmt.Errorf("trace is nil")
	}
	if err := s.saveToFile(trace); err != nil {
		log.Err(err).Msgf("[RawTraceFileSaver.InsertAndReturn] Failed to save trace to file")
		return nil, fmt.Errorf("failed to save trace to file: %w", err)
	}
	return trace, nil
}

// BatchInsertAndReturn inserts traces and returns the inserted traces.
func (s *RawTraceFileSaver) BatchInsertAndReturn(traces []*SimplifiedTrace) ([]*SimplifiedTrace, error) {
	res := make([]*SimplifiedTrace, 0)
	for _, trace := range traces {
		if insertRes, err := s.InsertAndReturn(trace); insertRes != nil && err == nil {
			res = append(res, insertRes)
		}
	}
	return res, nil
}

// SelectByIDs selects traces by IDs.
// It is not implemented for RawTraceFileSaver.
func (s *RawTraceFileSaver) SelectByIDs(ids []string) ([]*SimplifiedTrace, error) {
	return nil, fmt.Errorf("not implemented")
}

// Upsert inserts or updates a trace.
func (s *RawTraceFileSaver) Upsert(trace *SimplifiedTrace) error {
	if trace == nil {
		return fmt.Errorf("trace is nil")
	}
	if err := s.saveToFile(trace); err != nil {
		log.Err(err).Msgf("[RawTraceFileSaver.Upsert] Failed to save trace to file")
		return fmt.Errorf("failed to save trace to file: %w", err)
	}
	return nil
}

// BatchUpsert inserts or updates traces.
func (s *RawTraceFileSaver) BatchUpsert(traces []*SimplifiedTrace) error {
	for _, trace := range traces {
		if err := s.Upsert(trace); err != nil {
			log.Err(err).Msgf("[RawTraceFileSaver.BatchUpsert] Failed to upsert trace")
			return fmt.Errorf("failed to upsert trace: %w", err)
		}
	}
	return nil
}

// saveToFile saves a trace to a file.
// The file is named by the trace ID and is saved in the specified directory.
func (s *RawTraceFileSaver) saveToFile(trace *SimplifiedTrace) error {
	if trace == nil {
		return fmt.Errorf("trace is nil")
	}
	traceId := trace.TraceID
	// Save the trace into a file named by traceId under the directory.
	filePath := fmt.Sprintf("%s/%s.json", s.DirPath, traceId)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Err(err).Msgf("[RawTraceFileSaver.saveToFile] Failed to open file")
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Write the trace to the file.
	traceBytes, err := sonic.Marshal(trace)
	if err != nil {
		log.Err(err).Msgf("[RawTraceFileSaver.saveToFile] Failed to marshal trace")
		return fmt.Errorf("failed to marshal trace: %w", err)
	}
	if _, err := file.Write(traceBytes); err != nil {
		log.Err(err).Msgf("[RawTraceFileSaver.saveToFile] Failed to write trace to file")
		return fmt.Errorf("failed to write trace to file: %w", err)
	}

	return nil
}
