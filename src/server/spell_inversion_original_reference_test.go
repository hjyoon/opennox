package server

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/opennox/libs/types"
	"io"
	"math"
	"testing"
)

//go:embed testdata/inversion_original_52beb0_v1.json.gz
var inversionOriginalRangeReference []byte

type inversionOriginalRangeCase struct {
	Input struct {
		Name      string
		Caster    int
		RangeBits uint64 `json:"range_bits"`
	}
	Spills []uint32 `json:"binary32_spills"`
	Return *uint32  `json:"return_u32"`
	Fault  map[string]any
}

func inversionOriginalRangeRows(t *testing.T) []inversionOriginalRangeCase {
	t.Helper()
	if fmt.Sprintf("%x", sha256.Sum256(inversionOriginalRangeReference)) != "09ef7b359c4b988988f2a4ceb472375e432d0f1ab89f1f3f1fea2e50eac38ead" {
		t.Fatal("original compressed Inversion reference changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(inversionOriginalRangeReference))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "9027a202f77f366bc2e61716708d5f2a509b131f5bd685d0741a75e77f912eb2" {
		t.Fatal("original raw Inversion reference changed")
	}
	var rows []inversionOriginalRangeCase
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 79 {
		t.Fatal(len(rows))
	}
	return rows
}

// inversionOriginalRangeBits52BEB0 reads the observed original word, independently
// of the implementation under test. Existing cases retain their original inputs.
func inversionOriginalRangeBits52BEB0(t *testing.T, value float64) uint32 {
	t.Helper()
	for _, r := range inversionOriginalRangeRows(t) {
		if r.Input.RangeBits == math.Float64bits(value) {
			if len(r.Spills) != 1 {
				t.Fatal("original spill count")
			}
			return r.Spills[0]
		}
	}
	t.Fatalf("no original range reference for %016x", math.Float64bits(value))
	return 0
}

// Compare the observed range passed to enumeration and the original nil-caster
// fault prefix. Complete object/service traces are checked through the legacy C ABI.
func TestInversionOriginalRangeReference52BEB0(t *testing.T) {
	for _, c := range inversionOriginalRangeRows(t) {
		t.Run(c.Input.Name, func(t *testing.T) {
			caster := &Object{}
			if c.Input.Caster == -1 {
				caster = nil
			}
			balanced, queries, audio := 0, 0, 0
			var got int32
			var caught any
			func() {
				defer func() { caught = recover() }()
				got = inversionCast52BEB0(38, nil, caster, inversionCastDeps52BEB0{
					balance: func(string) float64 { balanced++; return math.Float64frombits(c.Input.RangeBits) },
					eachMissile: func(_ types.Pointf, radius float32, _ func(*Object) bool) {
						queries++
						if len(c.Spills) != 1 || math.Float32bits(radius) != c.Spills[0] {
							t.Fatalf("original radius=%08x, got=%08x", c.Spills[0], math.Float32bits(radius))
						}
					},
					changeOwner: func(*Object, *Object) { t.Fatal("numeric projection unexpectedly enumerated") },
					castAudio:   func(int32, *Object) { audio++ },
				})
			}()
			if balanced != 1 {
				t.Fatal("balance order", balanced)
			}
			if c.Fault != nil {
				if caught == nil || queries != 0 || audio != 0 || c.Return != nil {
					t.Fatal("original nil-caster prefix differs", caught, queries, audio)
				}
			} else if caught != nil || c.Return == nil || uint32(got) != *c.Return || queries != 1 || audio != 1 {
				t.Fatal("original range/result/audio projection differs", got, caught, queries, audio)
			}
		})
	}
}
