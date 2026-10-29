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

//go:embed testdata/spell_greater_heal_create_original_52f220.json
var greaterHealCreateOriginal52F220 []byte

type greaterHealCreateOriginalState52F220 struct {
	Caster   int       `json:"caster"`
	Target   int       `json:"target"`
	Mode     uint32    `json:"mode"`
	Spell    uint32    `json:"spell"`
	Aim      [2]uint32 `json:"aim"`
	HP       [3]int32  `json:"hp"`
	Rays     uint32    `json:"rays"`
	Messages uint32    `json:"messages"`
}

type greaterHealCreateOriginalEvent52F220 struct {
	Kind  string                               `json:"kind"`
	Args  []any                                `json:"args"`
	State greaterHealCreateOriginalState52F220 `json:"state"`
}

// Declared world-service observations are compared to executed GAME.EXE
// results. Expected argument caching, pointer reloads, stores and branching
// are never computed by this comparator.
func TestSpellGreaterHealCreateOriginal52F220(t *testing.T) {
	digest := sha256.Sum256(greaterHealCreateOriginal52F220)
	if hex.EncodeToString(digest[:]) != "82dfb7e603cce40e34beb0ccef0955515245cd8c22718a89fffaeca4f835d80b" {
		t.Fatal("sealed original create fixture")
	}
	type result struct {
		Return int32                                  `json:"return_value"`
		Events []greaterHealCreateOriginalEvent52F220 `json:"events"`
		State  greaterHealCreateOriginalState52F220   `json:"state"`
	}
	var fixture struct {
		Control uint32 `json:"control_word"`
		Cases   []struct {
			Input struct {
				greaterHealCreateOriginalState52F220
				Name         string `json:"name"`
				FlagsReturn  uint32 `json:"flags_return"`
				SearchReturn int    `json:"search_return"`
				EnemyReturn  uint32 `json:"enemy_return"`
				Mutation     struct {
					Stage  string `json:"stage"`
					At     int    `json:"at"`
					Action string `json:"action"`
				} `json:"mutation"`
			} `json:"input"`
			Result result `json:"result"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(greaterHealCreateOriginal52F220, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Control != 0xe3f || len(fixture.Cases) != 346 {
		t.Fatal("complete original create contexts")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Input.Name, func(t *testing.T) {
			in := c.Input
			state := in.greaterHealCreateOriginalState52F220
			aim := types.Ptf(math.Float32frombits(in.Aim[0]), math.Float32frombits(in.Aim[1]))
			pointBits := func(p types.Pointf) [2]uint32 { return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)} }
			snapshot := func() greaterHealCreateOriginalState52F220 { v := state; v.Aim = pointBits(aim); return v }
			events := make([]greaterHealCreateOriginalEvent52F220, 0)
			calls := make(map[string]int)
			changed := false
			event := func(kind string, args ...any) {
				events = append(events, greaterHealCreateOriginalEvent52F220{Kind: kind, Args: args, State: snapshot()})
				at := calls[kind]
				calls[kind]++
				m := in.Mutation
				if changed || m.Stage != kind || m.At != at {
					return
				}
				changed = true
				switch m.Action {
				case "caster-b":
					state.Caster = 1
				case "caster-nil":
					state.Caster = -1
				case "target-c":
					state.Target = 2
				case "target-nil":
					state.Target = -1
				case "aim":
					aim = types.Ptf(81.25, -91.5)
				case "mode":
					state.Mode = 1
				case "spell":
					state.Spell = 0x87654321
				default:
					t.Fatal("declared create mutation")
				}
			}
			h := greaterHealCreateHooks52F220[*greaterHealCreateOriginalState52F220, int]{
				loadMode:   func(r *greaterHealCreateOriginalState52F220) uint32 { return r.Mode },
				loadCaster: func(r *greaterHealCreateOriginalState52F220) int { return r.Caster + 1 },
				loadSpell:  func(r *greaterHealCreateOriginalState52F220) uint32 { return r.Spell },
				loadAim: func(r *greaterHealCreateOriginalState52F220) *types.Pointf {
					if r != &state {
						t.Fatal("record identity")
					}
					return &aim
				},
				spellFlags: func(id uint32) uint32 { event("flags", id); return in.FlagsReturn },
				searchTarget: func(p *types.Pointf, caster int, flags uint32, distance float32, mode int, self int) int {
					if p != &aim {
						t.Fatal("original aim alias")
					}
					event("search", pointBits(*p), caster-1, flags, math.Float32bits(distance), mode, self-1)
					return in.SearchReturn + 1
				},
				adjustHP: func(object int, amount int32) {
					if object < 1 || object > 3 || amount != 20 {
						t.Fatal("declared heal input")
					}
					event("adjust", object-1, amount)
					state.HP[object-1] += amount
				},
				storeTarget: func(r *greaterHealCreateOriginalState52F220, target int) { r.Target = target - 1 },
				isEnemy: func(caster, target int) bool {
					if in.EnemyReturn > 1 {
						t.Fatal("declared boolean enemy service")
					}
					event("enemy", caster-1, target-1)
					return in.EnemyReturn == 1
				},
				startRay: func(r *greaterHealCreateOriginalState52F220) {
					if r != &state {
						t.Fatal("ray record identity")
					}
					event("ray", "record")
					state.Rays++
				},
				noTarget: func(caster int) { event("message", caster-1, "ExecDur.c:GreaterHealNoTarget", 0); state.Messages++ },
			}
			value := spellGreaterHealCreate52F220(&state, h)
			got := result{Return: value, Events: events, State: snapshot()}
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
}
