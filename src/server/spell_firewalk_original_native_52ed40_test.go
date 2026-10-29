package server

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

//go:embed testdata/spell_firewalk_native_original_startup_52ed40.json
var firewalkOriginalStartup52ED40 []byte

//go:embed testdata/spell_firewalk_native_original_nearest_52ed40.json
var firewalkOriginalNearest52ED40 []byte

type firewalkOriginalState52ED40 struct {
	Target    int          `json:"target"`
	Level     uint32       `json:"level"`
	Frame60   uint32       `json:"frame60"`
	Frame64   uint32       `json:"frame64"`
	Previous  [2]uint32    `json:"previous"`
	Anchor    [2]uint32    `json:"anchor"`
	Positions [2][2]uint32 `json:"positions"`
	Flags     [2]uint32    `json:"flags"`
	Radii     [2]uint32    `json:"radii"`
	Cache     [3]uint32    `json:"cache"`
	FPS       uint32       `json:"fps"`
}

type firewalkOriginalInput52ED40 struct {
	firewalkOriginalState52ED40
	Name     string    `json:"name"`
	NameIDs  [3]uint32 `json:"name_ids"`
	Random   [2]int    `json:"random"`
	Allocate [2]bool   `json:"allocate"`
	Mutation struct {
		Stage  string `json:"stage"`
		At     int    `json:"at"`
		Action string `json:"action"`
	} `json:"mutation"`
}

type firewalkOriginalEvent52ED40 struct {
	Kind  string                      `json:"kind"`
	Args  []any                       `json:"args"`
	State firewalkOriginalState52ED40 `json:"state"`
}

type firewalkOriginalResult52ED40 struct {
	Return int32                         `json:"return_value"`
	Events []firewalkOriginalEvent52ED40 `json:"events"`
	State  firewalkOriginalState52ED40   `json:"state"`
}

// The comparator supplies declared world services and records their inputs.
// All expected arithmetic, bit patterns, cache writes, reloads and ordering
// come from GAME.EXE; this test computes no expected movement or threshold.
func TestSpellFirewalkOriginalNative52ED40(t *testing.T) {
	point := func(v [2]uint32) types.Pointf {
		return types.Ptf(math.Float32frombits(v[0]), math.Float32frombits(v[1]))
	}
	bits := func(v types.Pointf) [2]uint32 {
		return [2]uint32{math.Float32bits(v.X), math.Float32bits(v.Y)}
	}
	for _, profile := range []struct {
		name string
		data []byte
		cw   uint32
		hash string
	}{{"startup", firewalkOriginalStartup52ED40, 0xe3f, "b550467728d657951ee71f80a9e76277c3e6254cec8b8b4ae2d423949b924a0e"}, {"nearest", firewalkOriginalNearest52ED40, 0x23f, "2bfefcb5e2e887b51d7f85bc874e0368b4cacd1175343e2318c9659a444f888d"}} {
		t.Run(profile.name, func(t *testing.T) {
			digest := sha256.Sum256(profile.data)
			if hex.EncodeToString(digest[:]) != profile.hash {
				t.Fatal("sealed original fixture digest")
			}
			var fixture struct {
				Control uint32 `json:"control_word"`
				Cases   []struct {
					Input  firewalkOriginalInput52ED40  `json:"input"`
					Result firewalkOriginalResult52ED40 `json:"result"`
				} `json:"cases"`
			}
			if err := json.Unmarshal(profile.data, &fixture); err != nil {
				t.Fatal(err)
			}
			if fixture.Control != profile.cw || len(fixture.Cases) != 520 {
				t.Fatal("sealed original context set")
			}
			for _, c := range fixture.Cases {
				t.Run(c.Input.Name, func(t *testing.T) {
					in := c.Input
					state := in.firewalkOriginalState52ED40
					calls := make(map[string]int)
					events := make([]firewalkOriginalEvent52ED40, 0)
					changed := false
					event := func(kind string, args ...any) {
						events = append(events, firewalkOriginalEvent52ED40{Kind: kind, Args: args, State: state})
						at := calls[kind]
						calls[kind]++
						m := in.Mutation
						if changed || m.Stage != kind || m.At != at {
							return
						}
						changed = true
						switch m.Action {
						case "target-b":
							state.Target = 1
						case "target-position":
							state.Positions[0] = bits(types.Ptf(81.25, -91.5))
						case "previous":
							state.Previous = bits(types.Ptf(3.25, -4.5))
						case "anchor":
							state.Anchor = bits(types.Ptf(7.25, -8.5))
						case "level":
							state.Level = 4
						case "frame60":
							state.Frame60 = 11
						case "frame64":
							state.Frame64 = 999
						case "fps":
							state.FPS = 60
						case "cache":
							state.Cache = [3]uint32{91, 92, 93}
						default:
							t.Fatal("unknown declared mutation")
						}
					}
					allocate := func(typeID uint32) int {
						at := calls["allocate"]
						if at >= len(in.Allocate) {
							t.Fatal("extra allocation")
						}
						value := 0
						if in.Allocate[at] {
							value = 3 + at
						}
						index := -1
						if value != 0 {
							index = at
						}
						event("allocate", typeID, index)
						return value
					}
					create := func(object int, p types.Pointf) { b := bits(p); event("create", object-3, b[0], b[1]) }
					audio := func(p types.Pointf) { b := bits(p); event("audio", 46, b[0], b[1]) }
					decay := func(object int, duration uint32) { event("decay", object-3, duration) }
					h := firewalkHooks52ED40[*firewalkOriginalState52ED40, int]{
						loadTarget:    func(r *firewalkOriginalState52ED40) int { return r.Target + 1 },
						loadFlags:     func(o int) uint32 { return state.Flags[o-1] },
						loadFrame60:   func(r *firewalkOriginalState52ED40) uint32 { return r.Frame60 },
						loadFrame64:   func(r *firewalkOriginalState52ED40) uint32 { return r.Frame64 },
						storeFrame64:  func(r *firewalkOriginalState52ED40, v uint32) { r.Frame64 = v },
						loadPosition:  func(o int) types.Pointf { return point(state.Positions[o-1]) },
						loadRadius:    func(o int) float32 { return math.Float32frombits(state.Radii[o-1]) },
						loadPrevious:  func(r *firewalkOriginalState52ED40) types.Pointf { return point(r.Previous) },
						storePrevious: func(r *firewalkOriginalState52ED40, v types.Pointf) { r.Previous = bits(v) },
						loadAnchor:    func(r *firewalkOriginalState52ED40) types.Pointf { return point(r.Anchor) },
						storeAnchor:   func(r *firewalkOriginalState52ED40, v types.Pointf) { r.Anchor = bits(v) },
						loadLevel:     func(r *firewalkOriginalState52ED40) uint32 { return r.Level },
						randomInt: func(minimum, maximum int) int {
							at := calls["random"]
							if minimum != 0 || maximum < 0 || maximum > 2 || at >= len(in.Random) {
								t.Fatal("random input")
							}
							choice := in.Random[at]
							if choice > maximum {
								choice = maximum
							}
							event("random", minimum, maximum, choice)
							return choice
						},
						loadTypeID:  func(i int) uint32 { return state.Cache[i] },
						storeTypeID: func(i int, v uint32) { state.Cache[i] = v },
						nameID: func(name string) uint32 {
							i := -1
							for j, n := range [3]string{"SmallFlame", "MediumFlame", "Flame"} {
								if n == name {
									i = j
								}
							}
							if i < 0 {
								t.Fatal("type name")
							}
							event("name", name)
							return in.NameIDs[i]
						},
						allocateFlame: allocate, createFlame: create, audioFlame: audio, decayFlame: decay,
						tickRate:        func() uint32 { return state.FPS },
						roundTowardZero: profile.cw == 0xe3f,
						// Compatibility for the unchanged old monolithic callout.
						// It selects by name in production; no original cache access
						// or movement arithmetic is synthesized in this adapter.
						spawnFlame: func(kind int, p types.Pointf) {
							object := allocate(in.NameIDs[kind])
							if object != 0 {
								create(object, p)
								audio(p)
								decay(object, 25*state.FPS)
							}
						},
					}
					got := firewalkOriginalResult52ED40{Return: spellFirewalkUpdate52ED40(&state, h), Events: events, State: state}
					// Canonical JSON removes Go's concrete integer types from the
					// heterogeneous declared-service argument lists.
					canonical := func(v any) any {
						b, err := json.Marshal(v)
						if err != nil {
							t.Fatal(err)
						}
						var out any
						if err = json.Unmarshal(b, &out); err != nil {
							t.Fatal(err)
						}
						return out
					}
					if !reflect.DeepEqual(canonical(got), canonical(c.Result)) {
						a, _ := json.Marshal(got)
						b, _ := json.Marshal(c.Result)
						t.Fatalf("native=%s\noriginal=%s", a, b)
					}
				})
			}
		})
	}
}
