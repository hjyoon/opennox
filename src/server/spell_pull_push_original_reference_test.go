package server

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sync"
	"testing"

	"github.com/opennox/libs/types"
)

// Both packages embed the same sealed original blob; no runtime asset path is needed.
//
//go:embed testdata/pull_push_original_52bfa0_v1.json.gz
var pullPushForceOriginal []byte

type pullPushForceReferenceCase struct {
	Input struct {
		Name        string
		Cast        int
		Coefficient uint64 `json:"coefficient_bits"`
		Level       int32
	}
	Events []struct {
		Kind string
		Args json.RawMessage
	}
}

var pullPushReferenceOnce sync.Once
var pullPushReferenceRows []pullPushForceReferenceCase
var pullPushReferenceError error

func pullPushForceReferences(t *testing.T) []pullPushForceReferenceCase {
	t.Helper()
	pullPushReferenceOnce.Do(func() {
		if fmt.Sprintf("%x", sha256.Sum256(pullPushForceOriginal)) != "d5304b3b03c61e9ec103ee4df9b034d04cee2265ceb042ab1053d33ca50272d7" {
			pullPushReferenceError = fmt.Errorf("sealed GZIP differs")
			return
		}
		z, err := gzip.NewReader(bytes.NewReader(pullPushForceOriginal))
		if err != nil {
			pullPushReferenceError = err
			return
		}
		raw, err := io.ReadAll(z)
		if closeErr := z.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			pullPushReferenceError = err
			return
		}
		if fmt.Sprintf("%x", sha256.Sum256(raw)) != "cc1f7ed74c0e41f1fc729f0555b642c86cabbfa16cd303283dece417ac670b89" {
			pullPushReferenceError = fmt.Errorf("sealed raw reference differs")
			return
		}
		pullPushReferenceError = json.Unmarshal(raw, &pullPushReferenceRows)
		if pullPushReferenceError == nil && len(pullPushReferenceRows) != 566 {
			pullPushReferenceError = fmt.Errorf("wrong original case count")
		}
	})
	if pullPushReferenceError != nil {
		t.Fatal(pullPushReferenceError)
	}
	return pullPushReferenceRows
}
func pullPushReferenceForceBits(t *testing.T, c pullPushForceReferenceCase) uint32 {
	t.Helper()
	if len(c.Events) != 4 || c.Events[1].Kind != "radial" {
		t.Fatal("original radial transcript missing")
	}
	var args []uint32
	if err := json.Unmarshal(c.Events[1].Args, &args); err != nil || len(args) != 9 {
		t.Fatal("original radial arguments", err)
	}
	return args[5]
}
func pullPushOriginalForce52BFA0(t *testing.T, cast int, coefficient float64, level int32) float32 {
	t.Helper()
	for _, c := range pullPushForceReferences(t) {
		if c.Input.Cast == cast && c.Input.Coefficient == math.Float64bits(coefficient) && c.Input.Level == level {
			return math.Float32frombits(pullPushReferenceForceBits(t, c))
		}
	}
	t.Fatalf("numeric input absent from sealed original:%d/%016x/%d", cast, math.Float64bits(coefficient), level)
	return 0
}
func TestPullPushOriginalForceReference(t *testing.T) {
	for _, c := range pullPushForceReferences(t) {
		t.Run(c.Input.Name, func(t *testing.T) {
			var observed uint32
			balance := func(string) float64 { return math.Float64frombits(c.Input.Coefficient) }
			push := func(_ types.Pointf, _, _, force float32, _ *Object, _, _ int) { observed = math.Float32bits(force) }
			audio := func(int32, *Object) {}
			var got int32
			if c.Input.Cast == 0 {
				got = pullCast52BFA0(66, nil, new(Object), c.Input.Level, pullCastDeps52BFA0{balance: balance, pushUnits: push, castAudio: audio})
			} else {
				got = pushCast52C000(66, nil, new(Object), c.Input.Level, pushCastDeps52C000{balance: balance, pushUnits: push, castAudio: audio})
			}
			if wanted := pullPushReferenceForceBits(t, c); got != 1 || observed != wanted {
				t.Fatalf("original PC53/ToZero force=%08x; got=%08x return=%d", wanted, observed, got)
			}
		})
	}
}
