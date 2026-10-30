package server

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type dyingEndOriginalState544CA0 struct {
	Update   int       `json:"update"`
	Class    uint32    `json:"cls"`
	Subclass uint32    `json:"subclass"`
	Flags    [2]uint32 `json:"flags"`
}

type dyingEndOriginalEvent544CA0 struct {
	Kind  string                      `json:"kind"`
	Unit  int                         `json:"unit"`
	State dyingEndOriginalState544CA0 `json:"state"`
}

func TestMonsterDyingEndOriginalNative544CA0(t *testing.T) {
	path := "testdata/monster_dying_end_native_original_544ca0.json"
	if p := os.Getenv("NOX_MONSTER_DYING_END_ORIGINAL_FIXTURE"); p != "" {
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
				Name           string `json:"name"`
				Zombie         uint32 `json:"zombie"`
				Flags          uint32 `json:"flags"`
				AlternateFlags uint32 `json:"alternate_flags"`
				Mutation       struct {
					Stage  string `json:"stage"`
					Action string `json:"action"`
					Flags  uint32 `json:"flags"`
				} `json:"mutation"`
			} `json:"input"`
			Result struct {
				Events []dyingEndOriginalEvent544CA0 `json:"events"`
				State  dyingEndOriginalState544CA0   `json:"state"`
			} `json:"result"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.Control != 0xe3f || len(f.Rows) != 133 {
		t.Fatal("original dying-end identity/profile/count")
	}
	// The public original ABI is void. This existing handled-bool API compares
	// only observable state, arguments, service order and complete canaries.
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			type guarded struct {
				Before    [32]byte
				Unit      Object
				Primary   MonsterUpdateData
				Alternate MonsterUpdateData
				After     [32]byte
			}
			p, free := alloc.New(guarded{})
			defer free()
			*p = guarded{}
			for i := range p.Before {
				p.Before[i], p.After[i] = byte(0x21+i), byte(0xed-i)
			}
			unit := &p.Unit
			unit.ObjClass, unit.ObjSubClass, unit.UpdateData = object.ClassMonster, 0xa5, unsafe.Pointer(&p.Primary)
			p.Primary.StatusFlags, p.Alternate.StatusFlags = object.MonsterStatus(row.Input.Flags), object.MonsterStatus(row.Input.AlternateFlags)
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff || uintptr(unsafe.Pointer(&p.Alternate)) <= 0xffffffff) {
				t.Fatal("native original comparison links must exceed4GiB")
			}
			expected := *p
			state := func() dyingEndOriginalState544CA0 {
				index := 0
				if unit.UpdateData == unsafe.Pointer(&p.Alternate) {
					index = 1
				} else if unit.UpdateData != unsafe.Pointer(&p.Primary) {
					t.Fatal("unexpected native update identity")
				}
				return dyingEndOriginalState544CA0{Update: index, Class: uint32(unit.ObjClass), Subclass: uint32(unit.ObjSubClass), Flags: [2]uint32{uint32(p.Primary.StatusFlags), uint32(p.Alternate.StatusFlags)}}
			}
			mutate := func(stage string) {
				m := row.Input.Mutation
				if m.Stage != stage {
					return
				}
				switch m.Action {
				case "update", "primary", "alternate", "class", "subclass", "all":
				default:
					t.Fatal("unknown declared original mutation")
				}
				if m.Action == "update" || m.Action == "all" {
					unit.UpdateData = unsafe.Pointer(&p.Alternate)
				}
				if m.Action == "primary" || m.Action == "all" {
					p.Primary.StatusFlags = object.MonsterStatus(m.Flags)
				}
				if m.Action == "alternate" || m.Action == "all" {
					p.Alternate.StatusFlags = object.MonsterStatus(m.Flags)
				}
				if m.Action == "class" || m.Action == "all" {
					unit.ObjClass = object.Class(0x400000)
				}
				if m.Action == "subclass" || m.Action == "all" {
					unit.ObjSubClass = object.SubClass(0xfedcba98)
				}
			}
			events := make([]dyingEndOriginalEvent544CA0, 0)
			emit := func(kind string, got *Object) {
				if got != unit {
					t.Fatal("service native unit identity")
				}
				events = append(events, dyingEndOriginalEvent544CA0{Kind: kind, Unit: 0, State: state()})
				mutate(kind)
			}
			handled := new(Server).MonsterActionDyingEnd544CA0(unit, MonsterActionDyingRuntime544C40{
				IsZombie:         func(got *Object) bool { emit("zombie", got); return row.Input.Zombie != 0 },
				ZombieBurnDelete: func(got *Object) { emit("burn", got) },
				Unsupported:      func(reason string, _ *Object) { t.Fatal("unexpected valid original context rejected", reason) },
			})
			if !handled || !reflect.DeepEqual(events, row.Result.Events) || state() != row.Result.State {
				t.Fatalf("handled=%t events=%+v want%+v state=%+v want%+v", handled, events, row.Result.Events, state(), row.Result.State)
			}
			expected.Unit.ObjClass = object.Class(row.Result.State.Class)
			expected.Unit.ObjSubClass = object.SubClass(row.Result.State.Subclass)
			if row.Result.State.Update == 1 {
				expected.Unit.UpdateData = unsafe.Pointer(&p.Alternate)
			}
			expected.Primary.StatusFlags = object.MonsterStatus(row.Result.State.Flags[0])
			expected.Alternate.StatusFlags = object.MonsterStatus(row.Result.State.Flags[1])
			raw := func(v *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
			if !bytes.Equal(raw(p), raw(&expected)) {
				t.Fatal("complete native unit/primary/alternate/canary comparison")
			}
		})
	}
}
