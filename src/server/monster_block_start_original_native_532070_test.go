package server

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type blockOriginalState532070 struct {
	Update    int           `json:"update"`
	Subclass  uint32        `json:"subclass"`
	Frame     uint32        `json:"frame"`
	FPS       uint32        `json:"fps"`
	Indices   [2]int8       `json:"indices"`
	Deadlines [2][24]uint32 `json:"deadlines"`
}

type blockOriginalEvent532070 struct {
	Kind   string                   `json:"kind"`
	Unit   int                      `json:"unit"`
	Action uint32                   `json:"action,omitempty"`
	State  blockOriginalState532070 `json:"state"`
}

func TestMonsterBlockStartOriginalNative532070(t *testing.T) {
	path := "testdata/monster_block_start_native_original_532070.json"
	if p := os.Getenv("NOX_MONSTER_BLOCK_START_ORIGINAL_FIXTURE"); p != "" {
		path = p
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		OriginalSHA string `json:"original_sha256"`
		ControlWord uint16 `json:"control_word"`
		Rows        []struct {
			Input struct {
				Name      string `json:"name"`
				Shield    int    `json:"shield"`
				Frame     uint32 `json:"frame"`
				FPS       uint32 `json:"fps"`
				Deadline  uint32 `json:"deadline"`
				Subclass  uint32 `json:"subclass"`
				HeadIndex int8   `json:"head_index"`
				PopReturn uint32 `json:"pop_return"`
				PushNil   bool   `json:"push_nil"`
				Mutation  struct {
					Stage    string `json:"stage"`
					Action   string `json:"action"`
					Subclass uint32 `json:"subclass"`
					Deadline uint32 `json:"deadline"`
					Frame    uint32 `json:"frame"`
					FPS      uint32 `json:"fps"`
				} `json:"mutation"`
			} `json:"input"`
			Result struct {
				ReturnByte uint8                      `json:"return_byte"`
				Events     []blockOriginalEvent532070 `json:"events"`
				State      blockOriginalState532070   `json:"state"`
			} `json:"result"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.ControlWord != 0xe3f || len(f.Rows) != 485 {
		t.Fatal("original block-start identity/profile/count")
	}
	// This existing API returns handled-bool, not the public char register.
	// Compare original state/calls/canaries here; retain original EAX/char in
	// the unchanged reference for the separate actual public C composition.
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			type guarded struct {
				Before    [32]byte
				Unit      Object
				Primary   MonsterUpdateData
				Alternate MonsterUpdateData
				Pushed    AIStackItem
				After     [32]byte
			}
			p, free := alloc.New(guarded{})
			defer free()
			*p = guarded{}
			for i := range p.Before {
				p.Before[i], p.After[i] = byte(0x21+i), byte(0xed-i)
			}
			unit := &p.Unit
			unit.ObjClass, unit.ObjSubClass, unit.UpdateData = object.ClassMonster, object.SubClass(row.Input.Subclass), unsafe.Pointer(&p.Primary)
			p.Primary.AIStackInd, p.Alternate.AIStackInd = row.Input.HeadIndex, 2
			for _, ud := range []*MonsterUpdateData{&p.Primary, &p.Alternate} {
				for i := range ud.AIStack {
					ud.AIStack[i].Action = 21
					ud.AIStack[i].Args[0] = uintptr(0x24680000 + i)
				}
			}
			p.Primary.AIStack[row.Input.HeadIndex].Args[0], p.Alternate.AIStack[2].Args[0] = uintptr(row.Input.Deadline), 0x1234
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff || uintptr(unsafe.Pointer(&p.Alternate)) <= 0xffffffff || uintptr(unsafe.Pointer(&p.Pushed)) <= 0xffffffff) {
				t.Fatal("C-owned original comparison links must exceed4GiB")
			}
			expected := *p
			frame, fps := row.Input.Frame, row.Input.FPS
			state := func() blockOriginalState532070 {
				index := 0
				if unit.UpdateData == unsafe.Pointer(&p.Alternate) {
					index = 1
				} else if unit.UpdateData != unsafe.Pointer(&p.Primary) {
					t.Fatal("unknown update identity")
				}
				v := blockOriginalState532070{Update: index, Subclass: uint32(unit.ObjSubClass), Frame: frame, FPS: fps, Indices: [2]int8{p.Primary.AIStackInd, p.Alternate.AIStackInd}}
				for j, ud := range []*MonsterUpdateData{&p.Primary, &p.Alternate} {
					for i := range ud.AIStack {
						v.Deadlines[j][i] = ud.AIStack[i].ArgU32(0)
					}
				}
				return v
			}
			mutate := func(stage string) {
				m := row.Input.Mutation
				if m.Stage != stage {
					return
				}
				switch m.Action {
				case "unit_update", "index", "class", "deadline", "clock", "all":
				default:
					t.Fatal("unknown original mutation")
				}
				if m.Action == "unit_update" || m.Action == "all" {
					unit.UpdateData = unsafe.Pointer(&p.Alternate)
				}
				if m.Action == "index" || m.Action == "all" {
					p.Primary.AIStackInd = 2
				}
				if m.Action == "class" || m.Action == "all" {
					unit.ObjSubClass = object.SubClass(m.Subclass)
				}
				if m.Action == "deadline" || m.Action == "all" {
					p.Primary.AIStack[row.Input.HeadIndex].Args[0] = uintptr(m.Deadline)
				}
				if m.Action == "clock" || m.Action == "all" {
					frame, fps = m.Frame, m.FPS
				}
			}
			events := make([]blockOriginalEvent532070, 0)
			log := func(kind string, action uint32) {
				events = append(events, blockOriginalEvent532070{Kind: kind, Unit: 0, Action: action, State: state()})
				mutate(kind)
			}
			hooks := monsterActionBlockHooks532070{
				frame: func() uint32 { return frame }, tickRate: func() uint32 { return fps },
				testShield: func(got *Object) int {
					if got != unit {
						t.Fatal("shield unit identity")
					}
					log("shield", 0)
					return row.Input.Shield
				},
				pop: func() int { log("pop", 0); return int(row.Input.PopReturn) },
				push: func(action ai.ActionType, args ...any) *AIStackItem {
					if action != ai.ACTION_BLOCK_FINISH || len(args) != 0 {
						t.Fatal("original push arguments")
					}
					log("push", uint32(action))
					if row.Input.PushNil {
						return nil
					}
					return &p.Pushed
				},
			}
			if !monsterActionBlockAttack532070(unit, hooks) {
				t.Fatal("valid original block input was not handled")
			}
			if !reflect.DeepEqual(events, row.Result.Events) || state() != row.Result.State {
				t.Fatalf("original block-start mismatch\nevents=%+v want=%+v\nstate=%+v want=%+v", events, row.Result.Events, state(), row.Result.State)
			}
			expected.Unit.ObjSubClass = object.SubClass(row.Result.State.Subclass)
			if row.Result.State.Update == 1 {
				expected.Unit.UpdateData = unsafe.Pointer(&p.Alternate)
			}
			expected.Primary.AIStackInd, expected.Alternate.AIStackInd = row.Result.State.Indices[0], row.Result.State.Indices[1]
			for j, ud := range []*MonsterUpdateData{&expected.Primary, &expected.Alternate} {
				for i := range ud.AIStack {
					ud.AIStack[i].Args[0] = uintptr(row.Result.State.Deadlines[j][i])
				}
			}
			raw := func(v *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
			if !bytes.Equal(raw(p), raw(&expected)) {
				t.Fatal("object/update/stack/push frame/canaries changed outside original declared effects")
			}
		})
	}
}
