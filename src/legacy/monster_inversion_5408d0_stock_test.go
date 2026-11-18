package legacy

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
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type monsterInversionInput5408D0 struct {
	Name, Entry    string
	Enabled        bool
	CanCast        bool `json:"can_cast"`
	Mutate         bool
	Action         uint32
	Frame          uint32
	Deadline       uint32
	PostRNGFrame   uint32 `json:"post_rng_frame"`
	Balance        float64
	Spells, Deny   []int
	Selected       int
	Missiles       [][3]uint32
	CallbackTarget int `json:"callback_target"`
}

type monsterInversionOutcome5408D0 struct {
	Returned          int
	Events            json.RawMessage
	Threat, Deadline  uint32
	AlternateDeadline uint32 `json:"alternate_deadline"`
	KeepsOriginal     bool   `json:"keeps_original_update"`
}

func TestMonsterInversionFrozenOriginalC5408D0(t *testing.T) {
	content, err := os.ReadFile("testdata/monster_inversion_original_v1.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if sum := fmt.Sprintf("%x", sha256.Sum256(content)); sum != "adc9ecdbc8d57e2aca3ddd71343b3e3ace3049c76c7686ed7d28249570318e65" {
		t.Fatalf("stock monster inversion hash = %s", sum)
	}
	reader, err := gzip.NewReader(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var reference struct {
		Cases   int
		CW      string `json:"startup_control_word"`
		Guards  bool   `json:"object_and_register_guards"`
		Entered map[string]int
		Rows    []struct {
			Input    monsterInversionInput5408D0
			Observed monsterInversionOutcome5408D0
		}
	}
	if err := json.NewDecoder(reader).Decode(&reference); err != nil {
		t.Fatal(err)
	}
	if len(reference.Rows) != 31 || reference.Cases != 31 || reference.CW != "0xe3f" || !reference.Guards ||
		reference.Entered["inversion"] != 24 || reference.Entered["magic-missile"] != 26 ||
		reference.Entered["can-cast"] != 23 || reference.Entered["cast-stack-gate"] != 20 {
		t.Fatal("unexpected original monster inversion proof")
	}
	for _, row := range reference.Rows {
		t.Run(row.Input.Name, func(t *testing.T) { compareMonsterInversion5408D0(t, row.Input, row.Observed) })
	}
}

func compareMonsterInversion5408D0(t *testing.T, in monsterInversionInput5408D0, want monsterInversionOutcome5408D0) {
	t.Helper()
	update := game4ConsumerValue(t, server.MonsterUpdateData{Field363: in.Deadline, Field362_0: 11, Field362_2: 13})
	alternate := game4ConsumerValue(t, server.MonsterUpdateData{})
	update.AIStack[0].Action = in.Action
	if in.CanCast {
		update.StatusFlags = object.MonStatusCanCastSpells
	}
	flags := unsafe.Slice(&update.Field372, server.SpellsMax+1)
	for _, id := range in.Spells {
		flags[id] = 0x08000000
	}
	unit := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(12.5, -3.75),
		UpdateData: unsafe.Pointer(update)})
	if in.Enabled {
		unit.ObjFlags = object.FlagEnabled
	}
	other := game4ConsumerValue(t, server.Object{})
	roles := []*server.Object{nil, unit, other}
	for _, data := range in.Missiles {
		missile := game4ConsumerValue(t, server.Object{ObjClass: object.Class(data[0]), ObjSubClass: object.SubClass(data[1])})
		roles = append(roles, missile)
	}
	for i, data := range in.Missiles {
		mu := game4ConsumerValue(t, server.MissileUpdateData{Target: roles[data[2]]})
		roles[i+3].UpdateData = unsafe.Pointer(mu)
	}
	role := func(obj *server.Object) int {
		if index := slices.Index(roles, obj); index >= 0 {
			return index
		}
		t.Fatalf("unrecognized native object %p", obj)
		return -1
	}

	threat := memmap.PtrUint32(0x5D4594, 2489156)
	before := *threat
	*threat = 77
	t.Cleanup(func() { *threat = before })
	frame, calls := in.Frame, 0
	events := make([][]any, 0)
	event := func(args ...any) { events = append(events, args) }
	var result int
	if in.Entry == "magic-missile" {
		event("target-filter", 3, in.CallbackTarget)
		result = role(monsterInversionMissileFixture540B60(roles[3], roles[in.CallbackTarget]))
		if *threat != 77 {
			event("threat-write", *threat)
		}
	} else {
		runtime := server.MonsterCastInversionRuntime5408D0{
			Frame: func() uint32 { event("frame", frame); return frame },
			Balance: func(name string) float64 {
				if name != "InversionRange" {
					t.Fatalf("range lookup = %s", name)
				}
				event("balance")
				return in.Balance
			},
			EachMissile: func(pos types.Pointf, radius float32, visit func(*server.Object) bool) {
				event("query", math.Float32bits(pos.X), math.Float32bits(pos.Y), math.Float32bits(radius))
				for i, missile := range roles[3:] {
					event("target-filter", i+3, 1)
					if !visit(missile) {
						t.Fatal("missile enumeration stopped early")
					}
				}
			},
			LoadThreat: func() uint32 { event("threat-read", *threat); return *threat },
			StoreThreat: func(value uint32) {
				event("threat-write", value)
				*threat = value
			},
			SpellAllowed: func(id spell.ID) bool {
				event("allow", int(id))
				return !slices.Contains(in.Deny, int(id))
			},
			Random: func(min, max int) int {
				event("rng", min, max)
				calls++
				if calls == 1 {
					return in.Selected
				}
				frame = in.PostRNGFrame
				return max
			},
			Cast: func(unitArg *server.Object, id spell.ID, target *server.Object) {
				if unitArg != unit || target != unit {
					t.Fatalf("inversion received caster %p and target %p", unitArg, target)
				}
				event("cast", int(id))
				if in.Mutate {
					unit.UpdateData = unsafe.Pointer(alternate)
					update.Field362_0, update.Field362_2 = 0xfffe, 0xffff
				}
			},
		}
		oldEntry := monsterInversionEntry5408D0
		monsterInversionEntry5408D0 = func(o *server.Object) bool {
			if o != unit {
				t.Fatalf("C inversion entry received %p, want %p", o, unit)
			}
			return server.MonsterCastInversionNative5408D0(o, runtime)
		}
		t.Cleanup(func() { monsterInversionEntry5408D0 = oldEntry })
		result = monsterInversionFixture5408D0(unit)
	}
	gotEvents, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	if result != want.Returned || *threat != want.Threat || update.Field363 != want.Deadline ||
		alternate.Field363 != want.AlternateDeadline || (unit.UpdateDataMonster() == update) != want.KeepsOriginal ||
		!bytes.Equal(gotEvents, want.Events) {
		t.Fatalf("return/threat/deadlines = %d/%d/%d/%d, want %d/%d/%d/%d\ntrace: %s\nstock: %s",
			result, *threat, update.Field363, alternate.Field363, want.Returned, want.Threat, want.Deadline, want.AlternateDeadline,
			gotEvents, want.Events)
	}
}
