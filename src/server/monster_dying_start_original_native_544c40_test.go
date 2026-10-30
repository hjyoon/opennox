package server

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type dyingStartOriginalView544C40 struct {
	Definition  int    `json:"definition"`
	Sound       int    `json:"sound"`
	ScriptFlags uint32 `json:"script_flags"`
	ScriptFunc  int32  `json:"script_func"`
}

type dyingStartOriginalState544C40 struct {
	Update    int                             `json:"update"`
	Class     uint32                          `json:"cls"`
	Subclass  uint32                          `json:"subclass"`
	Flags     uint32                          `json:"flags"`
	Position  [2]float32                      `json:"position"`
	Updates   [2]dyingStartOriginalView544C40 `json:"updates"`
	Callbacks [3]int                          `json:"callbacks"`
	Sounds    [2][16]uint32                   `json:"sounds"`
}

type dyingStartOriginalEvent544C40 struct {
	Kind  string                        `json:"kind"`
	Args  []uint32                      `json:"args"`
	State dyingStartOriginalState544C40 `json:"state"`
}

// The existing handled-bool API compares original state, call arguments/order
// and whole native records. Raw EAX is retained outside that API contract.
// The49 standalone original getter rows are exercised by actual C compositions.
func TestMonsterDyingStartOriginalNative544C40(t *testing.T) {
	path := "testdata/monster_dying_start_native_original_544c40.json"
	if p := os.Getenv("NOX_MONSTER_DYING_START_ORIGINAL_FIXTURE"); p != "" {
		path = p
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA     string `json:"original_sha256"`
		Control uint16 `json:"control_word"`
		Rows    []struct {
			Input struct {
				Name     string `json:"name"`
				Entry    string `json:"entry"`
				Update   int    `json:"update"`
				Class    uint32 `json:"cls"`
				Sound    int    `json:"sound"`
				Callback int    `json:"callback"`
				Mutation struct {
					Stage  string `json:"stage"`
					Action string `json:"action"`
					Value  int    `json:"value"`
				} `json:"mutation"`
			} `json:"input"`
			Result struct {
				Events []dyingStartOriginalEvent544C40 `json:"events"`
				State  dyingStartOriginalState544C40   `json:"state"`
			} `json:"result"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.Control != 0xe3f || len(f.Rows) != 499 {
		t.Fatal("original dying-start/getter identity/profile/count")
	}
	rootCount, getterCount := 0, 0
	for _, row := range f.Rows {
		if row.Input.Entry == "getter" {
			getterCount++
			continue
		}
		if row.Input.Entry != "dying" {
			t.Fatal("unknown original entry")
		}
		rootCount++
		t.Run(row.Input.Name, func(t *testing.T) {
			defer func() {
				if v := recover(); v != nil {
					t.Errorf("valid original dying context panic: %v", v)
				}
			}()
			type guarded struct {
				Before      [32]byte
				Unit        Object
				Updates     [2]MonsterUpdateData
				Definitions [3]MonsterDef
				Sounds      [2][16]uint32
				Marks       [5]byte
				After       [32]byte
			}
			p, free := alloc.New(guarded{})
			defer free()
			*p = guarded{}
			for i := range p.Before {
				p.Before[i], p.After[i] = byte(0x21+i), byte(0xed-i)
			}
			unit := &p.Unit
			entry, alternate := &p.Updates[row.Input.Update], &p.Updates[1-row.Input.Update]
			unit.ObjClass, unit.ObjSubClass, unit.ObjFlags = object.Class(row.Input.Class), 0xa5, 0x12000004
			unit.PosVec = types.Ptf(321, 654)
			unit.UpdateData = unsafe.Pointer(entry)
			functions := [6]unsafe.Pointer{}
			functionIDs := map[unsafe.Pointer]int{nil: 0}
			for i := range p.Marks {
				p.Marks[i] = byte(0x80 + i)
				functions[i+1] = unsafe.Pointer(&p.Marks[i])
				functionIDs[functions[i+1]] = i + 1
			}
			soundPointers := [2]unsafe.Pointer{unsafe.Pointer(&p.Sounds[0][0]), unsafe.Pointer(&p.Sounds[1][0])}
			for n := range p.Updates {
				ud := &p.Updates[n]
				ud.MonsterDef = &p.Definitions[2]
				ud.SoundSet122 = soundPointers[1]
				ud.ScriptDeath = ScriptCallback{Flags: 0x5a, Func: 19}
				if ud == entry {
					ud.MonsterDef = &p.Definitions[0]
					ud.SoundSet122 = nil
					if row.Input.Sound >= 0 {
						ud.SoundSet122 = soundPointers[row.Input.Sound]
					}
					ud.ScriptDeath = ScriptCallback{Flags: 0xa5, Func: 17}
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unsafe.Pointer(ud)) <= 0xffffffff || uintptr(unsafe.Pointer(ud.MonsterDef)) <= 0xffffffff || uintptr(soundPointers[n]) <= 0xffffffff || uintptr(functions[n+1]) <= 0xffffffff) {
					t.Fatal("native object/update/definition/sound/callback links must exceed4GiB")
				}
			}
			for n := range p.Definitions {
				id := n + 1
				if n == 0 {
					id = row.Input.Callback
				}
				p.Definitions[n].DieFunc228 = functions[id]
			}
			for n := range p.Sounds {
				for i := range p.Sounds[n] {
					p.Sounds[n][i] = 0xa5000000 + uint32(n)*0x100 + uint32(i)
				}
				p.Sounds[n][15] = 715 + uint32(n)
			}
			expected := *p
			state := func() dyingStartOriginalState544C40 {
				v := dyingStartOriginalState544C40{Update: -1, Class: uint32(unit.ObjClass), Subclass: uint32(unit.ObjSubClass), Flags: uint32(unit.ObjFlags), Position: [2]float32{unit.PosVec.X, unit.PosVec.Y}, Sounds: p.Sounds}
				for n := range p.Updates {
					ud := &p.Updates[n]
					if unit.UpdateData == unsafe.Pointer(ud) {
						v.Update = n
					}
					d, s := -1, -1
					for i := range p.Definitions {
						if ud.MonsterDef == &p.Definitions[i] {
							d = i
						}
					}
					for i, ptr := range soundPointers {
						if ud.SoundSet122 == ptr {
							s = i
						}
					}
					if d < 0 || (s < 0 && ud.SoundSet122 != nil) {
						t.Fatal("unknown native definition/sound identity")
					}
					v.Updates[n] = dyingStartOriginalView544C40{Definition: d, Sound: s, ScriptFlags: ud.ScriptDeath.Flags, ScriptFunc: ud.ScriptDeath.Func}
				}
				if v.Update < 0 {
					t.Fatal("unknown current native update")
				}
				for i := range p.Definitions {
					var ok bool
					v.Callbacks[i], ok = functionIDs[p.Definitions[i].DieFunc228]
					if !ok {
						t.Fatal("unknown callback identity")
					}
				}
				return v
			}
			mutate := func(stage string) {
				m := row.Input.Mutation
				if m.Stage != stage {
					return
				}
				a, value := m.Action, m.Value
				switch a {
				case "update", "entry-definition", "alternate-definition", "entry-function", "function-A", "function-B", "function-C", "entry-sound", "alternate-sound", "sound-id", "script", "class", "flags", "position", "all":
				default:
					t.Fatal("unknown original mutation", a)
				}
				if a == "update" || a == "all" {
					unit.UpdateData = unsafe.Pointer(alternate)
				}
				if a == "entry-definition" || a == "all" {
					entry.MonsterDef = &p.Definitions[1]
				}
				if a == "alternate-definition" || a == "all" {
					alternate.MonsterDef = &p.Definitions[1]
				}
				if a == "entry-function" {
					entry.MonsterDef.DieFunc228 = functions[value]
				}
				if a == "function-A" {
					p.Definitions[0].DieFunc228 = functions[value]
				}
				if a == "function-B" || a == "all" {
					p.Definitions[1].DieFunc228 = functions[value]
				}
				if a == "function-C" {
					p.Definitions[2].DieFunc228 = functions[value]
				}
				if a == "entry-sound" || a == "all" {
					entry.SoundSet122 = soundPointers[1]
				}
				if a == "alternate-sound" || a == "all" {
					alternate.SoundSet122 = soundPointers[0]
				}
				if a == "sound-id" || a == "all" {
					p.Sounds[0][15] = 0xfedc0000 | uint32(value)
				}
				if a == "script" || a == "all" {
					entry.ScriptDeath = ScriptCallback{Flags: 0x5a, Func: 27}
				}
				if a == "class" || a == "all" {
					unit.ObjClass = 0x400000
					unit.ObjSubClass = 0xfedcba98
				}
				if a == "flags" || a == "all" {
					unit.ObjFlags = 0x87654321
				}
				if a == "position" || a == "all" {
					unit.PosVec = types.Ptf(-37.5, 19.25)
				}
			}
			events := make([]dyingStartOriginalEvent544C40, 0)
			emit := func(kind string, args []uint32) {
				events = append(events, dyingStartOriginalEvent544C40{Kind: kind, Args: args, State: state()})
				mutate(kind)
			}
			handled := new(Server).MonsterActionDyingStart544C40(unit, MonsterActionDyingRuntime544C40{
				AudioEvent: func(id uint32, got *Object) {
					if got != unit {
						t.Fatal("native audio unit identity")
					}
					emit("audio", []uint32{id, 0, 0, 0})
				},
				ScriptCallback: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) {
					if block != &entry.ScriptDeath || caller != nil || trigger != unit || event != NoxEventMonsterDead {
						t.Fatal("native cached script block/caller/trigger and existing event7 adapter extension")
					}
					emit("script", []uint32{uint32(row.Input.Update), 0, 0})
				},
				CanDieFunc: func(fn unsafe.Pointer) bool { id, ok := functionIDs[fn]; return ok && id != 0 },
				DieFunc: func(fn unsafe.Pointer, got *Object) {
					id, ok := functionIDs[fn]
					if !ok || id == 0 || got != unit {
						t.Fatal("native selected die callback identity")
					}
					emit("callback", []uint32{uint32(id), 0})
				},
				Unsupported: func(reason string, _ *Object) { t.Fatal("valid original context rejected", reason) },
			})
			if !handled || !reflect.DeepEqual(events, row.Result.Events) || state() != row.Result.State {
				t.Fatalf("handled=%t events=%+v want=%+v state=%+v want=%+v", handled, events, row.Result.Events, state(), row.Result.State)
			}
			want := row.Result.State
			expected.Unit.ObjClass, expected.Unit.ObjSubClass, expected.Unit.ObjFlags = object.Class(want.Class), object.SubClass(want.Subclass), object.Flags(want.Flags)
			expected.Unit.PosVec = types.Ptf(want.Position[0], want.Position[1])
			expected.Unit.UpdateData = unsafe.Pointer(&p.Updates[want.Update])
			for n, v := range want.Updates {
				ud := &expected.Updates[n]
				ud.MonsterDef = &p.Definitions[v.Definition]
				ud.SoundSet122 = nil
				if v.Sound >= 0 {
					ud.SoundSet122 = soundPointers[v.Sound]
				}
				ud.ScriptDeath = ScriptCallback{Flags: v.ScriptFlags, Func: v.ScriptFunc}
			}
			for i, id := range want.Callbacks {
				expected.Definitions[i].DieFunc228 = functions[id]
			}
			expected.Sounds = want.Sounds
			raw := func(v *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
			if !bytes.Equal(raw(p), raw(&expected)) {
				t.Fatal("complete native object/two-updates/three-definitions/sounds/callback-marks/canaries comparison")
			}
		})
	}
	if rootCount != 450 || getterCount != 49 {
		t.Fatal("450 root component and49 separate actual-C getter rows")
	}
}
