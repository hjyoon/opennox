package server

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type deadStartOriginalState544D80 struct {
	Update      int       `json:"update"`
	Class       uint32    `json:"cls"`
	Subclass    uint32    `json:"subclass"`
	Flags       uint32    `json:"flags"`
	Field131    uint32    `json:"field131"`
	Motion      [6]uint32 `json:"motion"`
	Definitions [2]int    `json:"definitions"`
	Callbacks   [3]int    `json:"callbacks"`
	Durations   [2]uint32 `json:"durations"`
	Balance     [2]uint32 `json:"balance"`
}

type deadStartOriginalEvent544D80 struct {
	Kind  string                       `json:"kind"`
	Arg   int                          `json:"arg"`
	Args  []int32                      `json:"args"`
	State deadStartOriginalState544D80 `json:"state"`
}

func TestMonsterDeadStartOriginalNative544D80(t *testing.T) {
	path := "testdata/monster_dead_start_native_original_544d80.json"
	if p := os.Getenv("NOX_MONSTER_DEAD_START_ORIGINAL_FIXTURE"); p != "" {
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
				Name         string    `json:"name"`
				Field131     uint32    `json:"field131"`
				Subclass     uint32    `json:"subclass"`
				Flags        uint32    `json:"flags"`
				Definition   int       `json:"definition"`
				Callback     int       `json:"callback"`
				Zombie       []uint32  `json:"zombie"`
				Balance      [2]uint32 `json:"balance"`
				RandomReturn uint32    `json:"random_return"`
				Mutation     struct {
					Stage  string `json:"stage"`
					Action string `json:"action"`
				} `json:"mutation"`
			} `json:"input"`
			Result struct {
				ReturnByte uint8                          `json:"return_byte"`
				Events     []deadStartOriginalEvent544D80 `json:"events"`
				State      deadStartOriginalState544D80   `json:"state"`
			} `json:"result"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || (f.Control != 0xe3f && f.Control != 0x23f) || len(f.Rows) != 630 {
		t.Fatal("original DeadStart identity/profile/count")
	}
	// This handled-bool API compares state/calls/canaries, not the public char.
	// Original RNG debug file/line metadata remains in the raw reference and is
	// outside this typed RNG callback. Profile selection changes the original
	// expectation only; it does not change or replace the engine conversion.
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			type guarded struct {
				Before          [32]byte
				Unit            Object
				Primary         MonsterUpdateData
				Alternate       MonsterUpdateData
				Defs            [3]MonsterDef
				FunctionMarkers [3][8]byte
				After           [32]byte
			}
			p, free := alloc.New(guarded{})
			defer free()
			*p = guarded{}
			for i := range p.Before {
				p.Before[i], p.After[i] = byte(0x21+i), byte(0xed-i)
			}
			fn := []unsafe.Pointer{nil, unsafe.Pointer(&p.FunctionMarkers[0][0]), unsafe.Pointer(&p.FunctionMarkers[1][0]), unsafe.Pointer(&p.FunctionMarkers[2][0])}
			fnIndex := func(ptr unsafe.Pointer) int {
				for i, v := range fn {
					if v == ptr {
						return i
					}
				}
				t.Fatal("unknown native function identity")
				return -1
			}
			defIndex := func(ptr *MonsterDef) int {
				for i := range p.Defs {
					if ptr == &p.Defs[i] {
						return i
					}
				}
				t.Fatal("unknown native definition identity")
				return -1
			}
			unit := &p.Unit
			unit.ObjClass, unit.ObjSubClass, unit.ObjFlags = object.ClassMonster, object.SubClass(row.Input.Subclass), object.Flags(row.Input.Flags)
			unit.Field131 = row.Input.Field131
			unit.UpdateData = unsafe.Pointer(&p.Primary)
			unit.VelVec, unit.ForceVec, unit.Pos24 = types.Ptf(1, 2), types.Ptf(3, 4), types.Ptf(5, 6)
			p.Primary.MonsterDef, p.Alternate.MonsterDef = &p.Defs[row.Input.Definition], &p.Defs[2]
			p.Primary.Field123, p.Alternate.Field123 = 0xcccc1111, 0xdddd2222
			for i := range p.Defs {
				p.Defs[i].DeadFunc232 = fn[i+1]
			}
			p.Defs[row.Input.Definition].DeadFunc232 = fn[row.Input.Callback]
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff || uintptr(unsafe.Pointer(&p.Defs[0])) <= 0xffffffff || uintptr(fn[1]) <= 0xffffffff) {
				t.Fatal("native original comparison links must exceed4GiB")
			}
			expected := *p
			balance := row.Input.Balance
			state := func() deadStartOriginalState544D80 {
				index := 0
				if unit.UpdateData == unsafe.Pointer(&p.Alternate) {
					index = 1
				} else if unit.UpdateData != unsafe.Pointer(&p.Primary) {
					t.Fatal("unknown current native update identity")
				}
				return deadStartOriginalState544D80{Update: index, Class: uint32(unit.ObjClass), Subclass: uint32(unit.ObjSubClass), Flags: uint32(unit.ObjFlags), Field131: unit.Field131, Motion: [6]uint32{math.Float32bits(unit.VelVec.X), math.Float32bits(unit.VelVec.Y), math.Float32bits(unit.ForceVec.X), math.Float32bits(unit.ForceVec.Y), math.Float32bits(unit.Pos24.X), math.Float32bits(unit.Pos24.Y)}, Definitions: [2]int{defIndex(p.Primary.MonsterDef), defIndex(p.Alternate.MonsterDef)}, Callbacks: [3]int{fnIndex(p.Defs[0].DeadFunc232), fnIndex(p.Defs[1].DeadFunc232), fnIndex(p.Defs[2].DeadFunc232)}, Durations: [2]uint32{p.Primary.Field123, p.Alternate.Field123}, Balance: balance}
			}
			mutate := func(stage string) {
				m := row.Input.Mutation
				if m.Stage != stage {
					return
				}
				switch m.Action {
				case "update", "definition", "callback", "flags", "motion", "primary_duration", "alternate_duration", "class", "subclass", "field131", "balance", "all":
				default:
					t.Fatal("unknown original mutation")
				}
				if m.Action == "update" || m.Action == "all" {
					unit.UpdateData = unsafe.Pointer(&p.Alternate)
				}
				if m.Action == "definition" || m.Action == "all" {
					p.Primary.MonsterDef = &p.Defs[1]
				}
				if m.Action == "callback" || m.Action == "all" {
					p.Primary.MonsterDef.DeadFunc232 = fn[2]
				}
				if m.Action == "flags" || m.Action == "all" {
					unit.ObjFlags = object.Flags(0x530000a5)
				}
				if m.Action == "motion" || m.Action == "all" {
					unit.VelVec = types.Ptf(math.Float32frombits(0x80000000), math.Float32frombits(0x7fc00001))
					unit.ForceVec = types.Ptf(5, -5)
					unit.Pos24 = types.Ptf(0.5, -0.5)
				}
				if m.Action == "primary_duration" || m.Action == "all" {
					p.Primary.Field123 = 0x02030405
				}
				if m.Action == "alternate_duration" || m.Action == "all" {
					p.Alternate.Field123 = 0x05060708
				}
				if m.Action == "class" || m.Action == "all" {
					unit.ObjClass = object.Class(0x400000)
				}
				if m.Action == "subclass" || m.Action == "all" {
					unit.ObjSubClass = object.SubClass(0xfedcba98)
				}
				if m.Action == "field131" || m.Action == "all" {
					unit.Field131 = 0x3344
				}
				if m.Action == "balance" || m.Action == "all" {
					balance = [2]uint32{0x40200000, 0x40600000}
				}
			}
			events := make([]deadStartOriginalEvent544D80, 0)
			emit := func(kind, stage string, arg int, got *Object, args ...int32) {
				if got != unit {
					t.Fatal("service native unit identity")
				}
				words := make([]int32, len(args))
				copy(words, args)
				events = append(events, deadStartOriginalEvent544D80{Kind: kind, Arg: arg, Args: words, State: state()})
				mutate(stage)
			}
			queries := 0
			handled := new(Server).MonsterActionDeadStart544D80(unit, MonsterActionDeadRuntime544D80{
				CreateReleasedSoul: func(got *Object) { emit("soul", "soul", 0, got) },
				CanDeadFunc:        func(ptr unsafe.Pointer) bool { return fnIndex(ptr) > 0 },
				DeadFunc:           func(ptr unsafe.Pointer, got *Object) { emit("dead", "dead", fnIndex(ptr), got) },
				IsZombie: func(got *Object) bool {
					if queries >= len(row.Input.Zombie) {
						t.Fatal("extra zombie query")
					}
					answer := row.Input.Zombie[queries]
					queries++
					stage := "zombie1"
					if queries == 2 {
						stage = "zombie2"
					}
					emit("zombie", stage, queries, got)
					return answer != 0
				},
				ZombieDeadDuration: func(index int) float32 {
					if index < 0 || index >= 2 {
						t.Fatal("duration index")
					}
					stage := "balance0"
					if index == 1 {
						stage = "balance1"
					}
					emit("balance", stage, index, unit)
					return math.Float32frombits(balance[index])
				},
				RandomInt: func(minimum, maximum int) int {
					emit("random", "random", 0, unit, int32(minimum), int32(maximum))
					return int(row.Input.RandomReturn)
				},
				Unsupported: func(reason string, _ *Object) { t.Fatal("valid original context rejected", reason) },
			})
			if !handled || !reflect.DeepEqual(events, row.Result.Events) || state() != row.Result.State {
				t.Fatalf("profile=%x handled=%t events=%+v want%+v state=%+v want%+v", f.Control, handled, events, row.Result.Events, state(), row.Result.State)
			}
			want := row.Result.State
			expected.Unit.ObjClass, expected.Unit.ObjSubClass, expected.Unit.ObjFlags = object.Class(want.Class), object.SubClass(want.Subclass), object.Flags(want.Flags)
			expected.Unit.Field131 = want.Field131
			if want.Update == 1 {
				expected.Unit.UpdateData = unsafe.Pointer(&p.Alternate)
			}
			expected.Unit.VelVec = types.Ptf(math.Float32frombits(want.Motion[0]), math.Float32frombits(want.Motion[1]))
			expected.Unit.ForceVec = types.Ptf(math.Float32frombits(want.Motion[2]), math.Float32frombits(want.Motion[3]))
			expected.Unit.Pos24 = types.Ptf(math.Float32frombits(want.Motion[4]), math.Float32frombits(want.Motion[5]))
			expected.Primary.MonsterDef, expected.Alternate.MonsterDef = &p.Defs[want.Definitions[0]], &p.Defs[want.Definitions[1]]
			expected.Primary.Field123, expected.Alternate.Field123 = want.Durations[0], want.Durations[1]
			for i := range expected.Defs {
				expected.Defs[i].DeadFunc232 = fn[want.Callbacks[i]]
			}
			raw := func(v *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
			if !bytes.Equal(raw(p), raw(&expected)) {
				t.Fatal("complete native unit/update/definitions/function-markers/canaries changed")
			}
		})
	}
}
