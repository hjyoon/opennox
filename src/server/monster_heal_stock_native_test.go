package server

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"math"
	"reflect"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

//go:embed testdata/monster_heal_original_v1.json.gz
var monsterHealOriginal5411A0 []byte

type monsterHealOriginalInput struct {
	Name           string
	Entry          string
	Flags          uint32
	CanCast        uint32 `json:"can_cast"`
	Frame          uint32
	AlreadyCasting uint32 `json:"already_casting"`
	Status         uint32
	HP             uint16
	MaxHP          uint16 `json:"max_hp"`
	Quest          uint32
	Position       [2]uint32
	IterResult     int `json:"iter_result"`
	Same           bool
	Enemy          uint32
	HasHealth      bool `json:"has_health"`
	Visible        uint32
	SwapHealth     []uint16 `json:"swap_health"`
	StartWinner    int      `json:"start_winner"`
}

type monsterHealOriginalOutcome struct {
	Calls  [][]any
	Winner int
	Result int
	Rect   []uint32
}

func TestMonsterHealStockStartup5411A0(t *testing.T) {
	reader, err := gzip.NewReader(bytes.NewReader(monsterHealOriginal5411A0))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var oracle struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Input    monsterHealOriginalInput
			Observed monsterHealOriginalOutcome
		}
	}
	if err := json.NewDecoder(reader).Decode(&oracle); err != nil {
		t.Fatal(err)
	}
	if oracle.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(oracle.Rows) != 272 {
		t.Fatal("unexpected original heal corpus")
	}
	for index, row := range oracle.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			in, want := row.Input, row.Observed
			caster := &Object{ObjClass: object.ClassMonster}
			ally := &Object{HealthData: &HealthData{Cur: 4, Max: 11}}
			previous := &Object{HealthData: &HealthData{Cur: 3, Max: 11}}
			targets := []*Object{nil, caster, ally, previous}
			if in.Entry == "heal-self-or-ally" {
				testMonsterHealSelectionStock5411A0(t, in, want, targets, index%2 == 0)
				return
			}
			if in.Entry != "heal-candidate" {
				t.Fatalf("unknown entry %q", in.Entry)
			}
			candidate := ally
			if in.Same {
				candidate = caster
			}
			candidate.ObjFlags = object.Flags(in.Flags)
			if in.HasHealth {
				candidate.HealthData = &HealthData{Cur: in.HP, Max: in.MaxHP}
			} else {
				candidate.HealthData = nil
			}
			calls := [][]any{}
			state := MonsterHealSelection5411A0{Target: targets[in.StartWinner]}
			state.considerHealCandidate5412A0(candidate, caster, func(from, to *Object) bool {
				if from != caster || to != candidate {
					t.Fatal("enemy query lost object identity")
				}
				calls = append(calls, []any{"enemy"})
				return in.Enemy != 0
			}, func(from, to *Object, mask int) bool {
				if from != caster || to != candidate || mask != 0 {
					t.Fatal("visibility query lost arguments")
				}
				calls = append(calls, []any{"interact"})
				if len(in.SwapHealth) == 2 {
					candidate.HealthData = &HealthData{Cur: in.SwapHealth[0], Max: in.SwapHealth[1]}
				}
				return in.Visible != 0
			})
			if got := slices.Index(targets, state.Target); got != want.Winner || !reflect.DeepEqual(calls, want.Calls) {
				t.Fatalf("candidate selected index %d with calls %v, stock %d and %v", got, calls, want.Winner, want.Calls)
			}
		})
	}
}

func testMonsterHealSelectionStock5411A0(t *testing.T, in monsterHealOriginalInput, want monsterHealOriginalOutcome, targets []*Object, useEmptyStack bool) {
	t.Helper()
	caster := targets[1]
	update := &MonsterUpdateData{StatusFlags: object.MonsterStatus(in.Status)}
	if in.CanCast != 0 {
		update.StatusFlags |= object.MonStatusCanCastSpells
	}
	if in.AlreadyCasting != 0 {
		if useEmptyStack {
			update.AIStackInd = -1
			update.Field132 = uint32(ai.ACTION_CAST_DURATION_SPELL)
		} else {
			update.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
		}
	}
	caster.UpdateData = unsafe.Pointer(update)
	caster.ObjFlags = object.Flags(in.Flags)
	caster.HealthData = &HealthData{Cur: in.HP, Max: in.MaxHP}
	caster.PosVec = types.Ptf(math.Float32frombits(in.Position[0]), math.Float32frombits(in.Position[1]))
	state := MonsterHealSelection5411A0{Target: targets[3]}
	var rect []uint32
	calls := [][]any{}
	got := state.healSomeone5411A0(caster, monsterHealHooks5411A0{
		frame: func() uint32 { return in.Frame },
		quest: func() bool {
			calls = append(calls, []any{"quest"})
			return in.Quest != 0
		},
		eachInRect: func(area types.Rectf, visit func(*Object) bool) {
			calls = append(calls, []any{"iterate"})
			rect = []uint32{math.Float32bits(area.Min.X), math.Float32bits(area.Min.Y),
				math.Float32bits(area.Max.X), math.Float32bits(area.Max.Y)}
			if state.Target != nil {
				t.Fatal("search begins with stale target")
			}
			if item := targets[in.IterResult]; item != nil && !visit(item) {
				t.Fatal("candidate callback stopped the scan")
			}
		},
		isEnemy:     func(*Object, *Object) bool { return false },
		canInteract: func(*Object, *Object, int) bool { return true },
		cast: func(from *Object, id spell.ID, to *Object) {
			if from != caster {
				t.Fatal("cast lost caster identity")
			}
			calls = append(calls, []any{"cast", float64(slices.Index(targets, to)), float64(id)})
		},
	})
	wantCalls := [][]any{}
	for _, call := range want.Calls {
		if call[0] != "can-cast" && call[0] != "already-casting" {
			wantCalls = append(wantCalls, call)
		}
	}
	if winner := slices.Index(targets, state.Target); got != want.Result || winner != want.Winner ||
		!reflect.DeepEqual(calls, wantCalls) || !reflect.DeepEqual(rect, want.Rect) {
		t.Fatalf("result=%d winner=%d rect=%08x calls=%v, stock result=%d winner=%d rect=%08x calls=%v",
			got, winner, rect, calls, want.Result, want.Winner, want.Rect, wantCalls)
	}
}
