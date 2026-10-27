// Package oracletest reads independently captured original-executable test
// observations. It is used by tests only; it performs no gameplay arithmetic.
package oracletest

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type startupObservations struct {
	OriginalSHA string `json:"original_sha256"`
	ControlWord uint16 `json:"control_word"`
	Fixtures    map[string]struct {
		SourceSHA string `json:"source_sha256"`
		RowCount  int    `json:"row_count"`
		Changes   []struct {
			Index  int                        `json:"index"`
			Fields map[string]json.RawMessage `json:"fields"`
		} `json:"changes"`
	} `json:"fixtures"`
}

var observationsCache sync.Map

func loadObservations(path string) (*startupObservations, error) {
	if value, ok := observationsCache.Load(path); ok {
		return value.(*startupObservations), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var fixture startupObservations
	if err := json.Unmarshal(data, &fixture); err != nil {
		return nil, err
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.ControlWord != 0xe3f || len(fixture.Fixtures) < 36 {
		return nil, fmt.Errorf("incomplete original startup observations")
	}
	value, _ := observationsCache.LoadOrStore(path, &fixture)
	return value.(*startupObservations), nil
}

// ReadNonItemStartup retains the frozen nearest-control input and applies only
// independently captured startup outputs. It rejects a changed input fixture,
// changed row count, or a correction to input/branch/identity fields.
func ReadNonItemStartup(path, observations string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fixture, err := loadObservations(observations)
	if err != nil {
		return nil, err
	}
	entry, ok := fixture.Fixtures["testdata/"+filepath.Base(path)]
	if !ok || fmt.Sprintf("%x", sha256.Sum256(data)) != entry.SourceSHA {
		return nil, fmt.Errorf("unmatched frozen historical input %q", path)
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, err
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(document["rows"], &rows); err != nil {
		return nil, err
	}
	if len(rows) != entry.RowCount {
		return nil, fmt.Errorf("changed historical row count for %q", path)
	}
	for _, change := range entry.Changes {
		if change.Index < 0 || change.Index >= len(rows) {
			return nil, fmt.Errorf("invalid original observation index for %q", path)
		}
		var row map[string]json.RawMessage
		if err := json.Unmarshal(rows[change.Index], &row); err != nil {
			return nil, err
		}
		for key, value := range change.Fields {
			switch key {
			case "forward_damage", "carry_bits", "wear", "events", "prefix_damage", "HP_damage":
				row[key] = value
			default:
				return nil, fmt.Errorf("original startup observation changes nonnumeric field %q", key)
			}
		}
		rows[change.Index], err = json.Marshal(row)
		if err != nil {
			return nil, err
		}
	}
	document["rows"], err = json.Marshal(rows)
	if err != nil {
		return nil, err
	}
	return json.Marshal(document)
}
