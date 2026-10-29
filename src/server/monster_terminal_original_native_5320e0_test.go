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

type terminalOriginalState5320E0 struct {
	Update         int   `json:"update"`
	Ready          uint8 `json:"ready"`
	AlternateReady uint8 `json:"alternate_ready"`
}

type terminalOriginalEvent5320E0 struct {
	Kind  string                      `json:"kind"`
	Unit  int                         `json:"unit"`
	State terminalOriginalState5320E0 `json:"state"`
}

func TestMonsterTerminalOriginalNative5320E0(t *testing.T) {
	path := "testdata/monster_terminal_native_original_5320e0.json"
	if override := os.Getenv("NOX_MONSTER_TERMINAL_ORIGINAL_FIXTURE"); override != "" {
		path = override
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		ControlWord uint16 `json:"control_word"`
		Rows        []struct {
			Input struct {
				Name          string `json:"name"`
				Function      string `json:"function"`
				Ready         uint8  `json:"ready"`
				PopReturn     uint32 `json:"pop_return"`
				Mutation      string `json:"mutation"`
				MutationReady uint8  `json:"mutation_ready"`
			} `json:"input"`
			Result struct {
				ReturnByte uint8                         `json:"return_byte"`
				Events     []terminalOriginalEvent5320E0 `json:"events"`
				State      terminalOriginalState5320E0   `json:"state"`
			} `json:"result"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.ControlWord != 0xe3f || len(fixture.Rows) != 1350 {
		t.Fatal("original terminal fixture identity/profile/count")
	}
	// All three original roots have the same complete instruction sequence,
	// apart from their relative PopAction displacement. Compare that terminal
	// component against the existing injectable block-finish implementation.
	// Public DYING and both block C paths are checked separately. Raw EAX and
	// char bytes stay in the fixture; this component API returns handled-bool,
	// and the ordinary action dispatcher discards the public char result.
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			if row.Input.Function != "shield-finish" && row.Input.Function != "weapon-block" && row.Input.Function != "dying" {
				t.Fatal("unknown original root")
			}
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
				p.Before[i], p.After[i] = byte(0x29+i), byte(0xf1-i)
			}
			unit := &p.Unit
			unit.ObjClass, unit.UpdateData = object.ClassMonster, unsafe.Pointer(&p.Primary)
			p.Primary.Field120_3, p.Alternate.Field120_3 = row.Input.Ready, 0x53
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(unit)) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff || uintptr(unsafe.Pointer(&p.Alternate)) <= 0xffffffff) {
				t.Fatal("original comparator requires C-owned links above4GiB")
			}
			expected := *p
			state := func() terminalOriginalState5320E0 {
				index := 0
				if unit.UpdateData == unsafe.Pointer(&p.Alternate) {
					index = 1
				} else if unit.UpdateData != unsafe.Pointer(&p.Primary) {
					t.Fatal("unknown native update identity")
				}
				return terminalOriginalState5320E0{index, p.Primary.Field120_3, p.Alternate.Field120_3}
			}
			events := make([]terminalOriginalEvent5320E0, 0)
			pop := func() int {
				events = append(events, terminalOriginalEvent5320E0{"pop", 0, state()})
				switch row.Input.Mutation {
				case "none":
				case "ready":
					p.Primary.Field120_3 = row.Input.MutationReady
				case "replace-update":
					unit.UpdateData = unsafe.Pointer(&p.Alternate)
					p.Alternate.Field120_3 = row.Input.MutationReady
				default:
					t.Fatal("unknown original service mutation")
				}
				return int(row.Input.PopReturn)
			}
			if !monsterActionBlockTerminal5320E0(unit, pop) {
				t.Fatal("valid monster terminal was not handled")
			}
			if !reflect.DeepEqual(events, row.Result.Events) || state() != row.Result.State {
				t.Fatalf("original terminal mismatch\nevents=%+v want=%+v\nstate=%+v want=%+v", events, row.Result.Events, state(), row.Result.State)
			}
			expected.Primary.Field120_3, expected.Alternate.Field120_3 = row.Result.State.Ready, row.Result.State.AlternateReady
			if row.Result.State.Update == 1 {
				expected.Unit.UpdateData = unsafe.Pointer(&p.Alternate)
			}
			raw := func(v *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
			if !bytes.Equal(raw(p), raw(&expected)) {
				t.Fatal("native object/update records or canaries changed outside declared service effects")
			}
		})
	}
}
