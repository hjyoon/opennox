package server

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"testing"

	"github.com/opennox/libs/types"
)

type scorchOriginalStep537AF0 struct {
	Entry     string
	IDs       []uint32
	Kind      int
	Quest     bool
	Allocated bool
	FPS       uint32
	Seconds   int
	Pos       [2]uint32
}

type scorchOriginalCase537AF0 struct {
	Name       string
	Initial    uint32
	Cache, IDs [3]uint32
	Sequence   []scorchOriginalStep537AF0
	Executions []struct {
		Ready  uint32
		Cache  [3]uint32
		Trace  json.RawMessage
		Return *uint32
	}
}

func TestScorch537AF0FrozenOriginal(t *testing.T) {
	data, err := os.ReadFile("testdata/scorch_original_v1.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != "c79b3d5ce09b2c333322f321f5ada07f822eb126dfbcd0dd752b4c5d3f1fb8ed" {
		t.Fatalf("unexpected original corpus hash %s", got)
	}
	z, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var original struct {
		OriginalSHA string         `json:"original_sha256"`
		Code        map[string]int `json:"executed_bytes"`
		Executions  map[string]int
		Guards      bool   `json:"guards_verified"`
		ControlWord string `json:"control_word"`
		Rows        []scorchOriginalCase537AF0
	}
	if err := json.NewDecoder(z).Decode(&original); err != nil {
		t.Fatal(err)
	}
	if original.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		original.Code["make"] != 216 || original.Code["initialize"] != 64 || !original.Guards ||
		original.ControlWord != "0xe3f" || original.Executions["make"] != 62 || original.Executions["initialize"] != 6 ||
		len(original.Rows) != 64 {
		t.Fatal("unexpected original Scorch corpus scope")
	}
	for _, row := range original.Rows {
		t.Run(row.Name, func(t *testing.T) {
			cache := scorchTypeCache537BD0{ready: row.Initial, ids: row.Cache}
			for i, step := range row.Sequence {
				want := row.Executions[i]
				trace := make([][]any, 0)
				typeIndex := func(name string) int {
					ind := slices.Index(scorchTypeIDs537AF0[:], name)
					ids := row.IDs[:]
					if step.IDs != nil {
						ids = step.IDs
					}
					value := ids[ind]
					trace = append(trace, []any{"lookup", name, value})
					return int(value)
				}
				switch step.Entry {
				case "initialize":
					if got := uint32(cache.initialize(typeIndex)); want.Return == nil || got != *want.Return {
						t.Fatalf("step %d initializer returned %d, want %v", i, got, want.Return)
					}
				case "make":
					pos := types.Ptf(math.Float32frombits(step.Pos[0]), math.Float32frombits(step.Pos[1]))
					makeScorch537AF0(pos, step.Kind, &cache, makeScorchHooks537AF0[bool]{
						typeIndex: typeIndex,
						newObject: func(ind int) bool {
							trace = append(trace, []any{"new", uint32(ind)})
							return step.Allocated
						},
						createObjectAt: func(_ bool, pos types.Pointf) {
							trace = append(trace, []any{"create", math.Float32bits(pos.X), math.Float32bits(pos.Y)})
						},
						gameFlag: func(flag uint32) int32 {
							trace = append(trace, []any{"flag", flag})
							if step.Quest {
								return 1
							}
							return 0
						},
						randomInt: func(min, max int) int {
							result := step.Seconds
							if min == 0 && max == 0 {
								result = 0
							}
							trace = append(trace, []any{"random", min, max, result})
							return result
						},
						loadGameFPS: func() uint32 { return step.FPS },
						setDecay: func(_ bool, frames uint32) {
							trace = append(trace, []any{"decay", frames})
						},
					})
				default:
					t.Fatalf("unexpected original entry %q", step.Entry)
				}
				got, err := json.Marshal(trace)
				if err != nil {
					t.Fatal(err)
				}
				if cache.ready != want.Ready || cache.ids != want.Cache || !bytes.Equal(got, want.Trace) {
					t.Fatalf("step %d: cache %+v and trace %s, want %d/%v and %s", i, cache, got, want.Ready, want.Cache, want.Trace)
				}
			}
		})
	}
}
