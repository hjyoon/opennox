package legacy

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"bytes"
	"encoding/binary"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type trapUndeadState struct {
	Flags   uint32       `json:"flags"`
	Status  uint32       `json:"status"`
	Next    uint32       `json:"next"`
	Collide int          `json:"collide_ref"`
	Traps   [2][7]uint32 `json:"traps"`
	Undead  [2]int       `json:"undead_refs"`
	Spells  [2]uint32    `json:"spell_flags"`
	Frame   uint32       `json:"frame"`
	FPS     uint32       `json:"fps"`
}
type trapUndeadInput struct {
	Name     string `json:"name"`
	Kind     int    `json:"kind"`
	Mutation string `json:"mutation"`
	trapUndeadState
}
type trapUndeadEvent struct {
	Service string  `json:"service"`
	Args    []int64 `json:"args"`
	trapUndeadState
}
type trapUndeadRow struct {
	Input     trapUndeadInput   `json:"input"`
	Output    trapUndeadState   `json:"output"`
	Events    []trapUndeadEvent `json:"events"`
	Preserved bool              `json:"nonoutput_bytes_preserved"`
	Source    int               `json:"source_ref"`
	Fault     bool              `json:"fault"`
}

func trapUndeadRun(t *testing.T, in trapUndeadInput, sourceRef int, want trapUndeadState) (trapUndeadState, []trapUndeadEvent, any) {
	t.Helper()
	var source *server.Object
	var traps [2]*server.TrapDoorCollideData
	var undead [2]*server.UndeadKillerCollideData
	var spells [2]*server.DurSpell
	var frames, before [7][]byte
	sizes := []uintptr{unsafe.Sizeof(server.Object{}), unsafe.Sizeof(server.TrapDoorCollideData{}), unsafe.Sizeof(server.TrapDoorCollideData{}), unsafe.Sizeof(server.UndeadKillerCollideData{}), unsafe.Sizeof(server.UndeadKillerCollideData{}), unsafe.Sizeof(server.DurSpell{}), unsafe.Sizeof(server.DurSpell{})}
	for n, size := range sizes {
		frames[n] = radialCFrameOriginal(t, size)
		clear(frames[n][32 : 32+size])
	}
	source = (*server.Object)(unsafe.Pointer(&frames[0][32]))
	for n := range traps {
		traps[n] = (*server.TrapDoorCollideData)(unsafe.Pointer(&frames[1+n][32]))
		undead[n] = (*server.UndeadKillerCollideData)(unsafe.Pointer(&frames[3+n][32]))
		spells[n] = (*server.DurSpell)(unsafe.Pointer(&frames[5+n][32]))
	}

	spellRef := func(n int) *server.DurSpell {
		if n < 0 {
			return nil
		}
		return spells[n]
	}
	recordRef := func(n int) unsafe.Pointer {
		if n < 0 {
			return nil
		}
		if in.Kind == 0 {
			return unsafe.Pointer(traps[n])
		}
		return unsafe.Pointer(undead[n])
	}
	*source = server.Object{ObjFlags: object.Flags(in.Flags), Field5: in.Status, Field34: in.Next, CollideData: recordRef(in.Collide)}
	for n := range traps {
		*(*[7]uint32)(unsafe.Pointer(traps[n])) = in.Traps[n]
		undead[n].Spell = spellRef(in.Undead[n])
		spells[n].Flags88 = in.Spells[n]
	}
	for n := range frames {
		before[n] = append([]byte(nil), frames[n]...)
	}
	snapshot := func() trapUndeadState {
		s := trapUndeadState{Flags: uint32(source.ObjFlags), Status: source.Field5, Next: source.Field34, Collide: -1, Frame: in.Frame, FPS: in.FPS}
		if source.CollideData != nil {
			for n := range traps {
				if source.CollideData == recordRef(n) {
					s.Collide = n
				}
			}
			if s.Collide < 0 {
				t.Fatal("unknown collision record")
			}
		}
		for n := range traps {
			s.Traps[n] = *(*[7]uint32)(unsafe.Pointer(traps[n]))
			s.Spells[n] = spells[n].Flags88
			s.Undead[n] = -1
			if undead[n].Spell != nil {
				for k, p := range spells {
					if undead[n].Spell == p {
						s.Undead[n] = k
					}
				}
				if s.Undead[n] < 0 {
					t.Fatal("unknown spell record")
				}
			}
		}
		return s
	}
	events := make([]trapUndeadEvent, 0)
	event := func(service string, args ...int64) {
		events = append(events, trapUndeadEvent{Service: service, Args: args, trapUndeadState: snapshot()})
	}
	applied := false
	mutate := func(service string) {
		prefix := service + ":"
		if applied || !strings.HasPrefix(in.Mutation, prefix) {
			return
		}
		applied = true
		switch strings.TrimPrefix(in.Mutation, prefix) {
		case "clock":
			in.Frame = 0xfffffffc
		case "fps":
			in.FPS = 0x80000001
		case "flags":
			source.ObjFlags = 0x80000000
		case "status":
			source.Field5 = 0x10000002
		case "next":
			source.Field34 = 0x12345678
		case "deadline":
			traps[0].NextFrame = 0xfffffffc
		case "replacement-deadline":
			traps[1].NextFrame = 99
		case "collide-swap":
			source.CollideData = recordRef(1)
		case "collide-nil":
			source.CollideData = nil
		case "spell-swap":
			undead[0].Spell = spells[1]
		case "cancel":
			spells[0].Flags88 = 1
		case "activate":
			traps[0].Activated = 99
		default:
			t.Fatal("unknown mutation")
		}
	}
	h := server.TrapDoorUpdateRawRuntime53DE80{
		Frame: func() uint32 { return in.Frame }, FPS: func() uint32 { return in.FPS },
		SetOnOff: func(o *server.Object, on bool) {
			if o != source || !on {
				t.Fatal("on-off args")
			}
			event("on", 0, 1)
			o.ObjFlags |= object.FlagEnabled
			mutate("on")
		},
		UnsetXStatus: func(o *server.Object, v uint32) {
			if o != source {
				t.Fatal("unset object")
			}
			event("unset", 0, int64(v))
			o.Field5 &^= v
			mutate("unset")
		},
		SetXStatus: func(o *server.Object, v uint32) {
			if o != source {
				t.Fatal("set object")
			}
			event("set", 0, int64(v))
			o.Field5 |= v
			mutate("set")
		},
		AudioEvent: func(id uint32, o *server.Object) {
			if o != source {
				t.Fatal("audio object")
			}
			event("audio", int64(id), 0, 0, 0)
			mutate("audio")
		},
	}
	u := server.UndeadKillerUpdateRawRuntime53E190{Frame: h.Frame, DelayedDelete: func(o *server.Object) {
		if o != source {
			t.Fatal("delete object")
		}
		event("delete", 0)
		mutate("delete")
	}}
	arg := source
	if sourceRef < 0 {
		arg = nil
	}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		oldT, oldU := trapDoorUpdateRawRuntime53DE80, undeadKillerUpdateRawRuntime53E190
		defer func() { trapDoorUpdateRawRuntime53DE80 = oldT; undeadKillerUpdateRawRuntime53E190 = oldU }()
		trapDoorUpdateRawRuntime53DE80 = func() server.TrapDoorUpdateRawRuntime53DE80 { return h }
		undeadKillerUpdateRawRuntime53E190 = func() server.UndeadKillerUpdateRawRuntime53E190 { return u }
		trapDoorUndeadNativeProbe(in.Kind, arg)
	}()
	putptr := func(b []byte, off uintptr, p unsafe.Pointer) {
		if unsafe.Sizeof(uintptr(0)) == 8 {
			binary.LittleEndian.PutUint64(b[32+off:], uint64(uintptr(p)))
		} else {
			binary.LittleEndian.PutUint32(b[32+off:], uint32(uintptr(p)))
		}
	}
	expected := [7][]byte{}
	for n := range before {
		expected[n] = append([]byte(nil), before[n]...)
	}
	for n, off := range []uintptr{unsafe.Offsetof(source.ObjFlags), unsafe.Offsetof(source.Field5), unsafe.Offsetof(source.Field34)} {
		binary.LittleEndian.PutUint32(expected[0][32+off:], []uint32{want.Flags, want.Status, want.Next}[n])
	}
	putptr(expected[0], unsafe.Offsetof(source.CollideData), recordRef(want.Collide))
	for n := range traps {
		for k, v := range want.Traps[n] {
			binary.LittleEndian.PutUint32(expected[1+n][32+4*k:], v)
		}
		putptr(expected[3+n], 0, unsafe.Pointer(spellRef(want.Undead[n])))
		binary.LittleEndian.PutUint32(expected[5+n][32+unsafe.Offsetof(spells[n].Flags88):], want.Spells[n])
	}
	for n := range frames {
		if !bytes.Equal(expected[n], frames[n]) {
			t.Fatalf("native object/data/spell frame/canary %d", n)
		}
	}

	return snapshot(), events, recovered
}

func TestTrapDoorUndeadOriginalNativeC(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/trap_door_undead_updates_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA      string          `json:"original_sha256"`
		CW       uint32          `json:"original_startup_FPCW"`
		Contexts int             `json:"contexts"`
		Rows     []trapUndeadRow `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.CW != 0xe3f || f.Contexts != 696 || len(f.Rows) != 696 {
		t.Fatal("original provenance")
	}
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			out, events, fault := trapUndeadRun(t, row.Input, 0, row.Output)
			if !row.Preserved || fault != nil {
				t.Fatal("original/native fault or canary", fault)
			}
			if !reflect.DeepEqual(out, row.Output) {
				t.Errorf("state differs: got %+v want %+v", out, row.Output)
			}
			if !reflect.DeepEqual(events, row.Events) {
				t.Errorf("service order/arguments/state differ: got %+v want %+v", events, row.Events)
			}
		})
	}
}

func TestTrapDoorUndeadOriginalFaultsNativeC(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/trap_door_undead_updates_faults_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Rows []trapUndeadRow `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Rows) != 10 {
		t.Fatal("fault provenance")
	}
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			out, events, fault := trapUndeadRun(t, row.Input, row.Source, row.Output)
			if (fault != nil) != row.Fault || !reflect.DeepEqual(out, row.Output) || !reflect.DeepEqual(events, row.Events) {
				t.Fatal("fault/gate/prefix/state differs", fault)
			}
		})
	}
}
