package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const monsterScriptPauseHighUnit516090 = uint64(0x7f9c86c94f90)

type monsterScriptPauseTrace516090 struct {
	events   []string
	classLow uint8
	flags    uint32
	frame    uint32
	nilPush  map[uint32]bool
	args     map[uint64][4]uint32
}

func newMonsterScriptPauseTrace516090() *monsterScriptPauseTrace516090 {
	return &monsterScriptPauseTrace516090{
		classLow: uint8(object.ClassMonster),
		frame:    0xfffffff0,
		nilPush:  make(map[uint32]bool),
		args:     make(map[uint64][4]uint32),
	}
}

func (w *monsterScriptPauseTrace516090) hooks() monsterScriptPauseHooks516090[uint64, uint64] {
	return monsterScriptPauseHooks516090[uint64, uint64]{
		loadClassLow: func(unit uint64) uint8 {
			w.events = append(w.events, fmt.Sprintf("class:%x", unit))
			return w.classLow
		},
		loadFlags: func(unit uint64) uint32 {
			w.events = append(w.events, fmt.Sprintf("flags:%x", unit))
			return w.flags
		},
		pushAction: func(unit uint64, action uint32) uint64 {
			w.events = append(w.events, fmt.Sprintf("push:%x:%d", unit, action))
			if w.nilPush[action] {
				return 0
			}
			return 0x700000000 + uint64(action)
		},
		loadFrame: func() uint32 {
			w.events = append(w.events, "frame")
			return w.frame
		},
		storeArgBits: func(item uint64, index int, bits uint32) {
			w.events = append(w.events, fmt.Sprintf("arg:%x:%d:%08x", item, index, bits))
			args := w.args[item]
			args[index] = bits
			w.args[item] = args
		},
	}
}

func TestMonsterScriptPause516090ExactOrderAndNativeWidth(t *testing.T) {
	w := newMonsterScriptPauseTrace516090()
	monsterScriptPause516090(monsterScriptPauseHighUnit516090, 0x30, w.hooks())
	want := []string{
		"class:7f9c86c94f90", "flags:7f9c86c94f90",
		"push:7f9c86c94f90:32", "arg:700000020:0:00000001",
		"push:7f9c86c94f90:1", "frame", "arg:700000001:0:00000020",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if got := w.args[0x700000020]; got != [4]uint32{uint32(ai.ACTION_WAIT)} {
		t.Fatalf("REPORT args = %#v", got)
	}
	if got := w.args[0x700000001]; got != [4]uint32{0x20} {
		t.Fatalf("WAIT args = %#v", got)
	}
}

func TestMonsterScriptPause516090GatesAndPushFailures(t *testing.T) {
	for _, tc := range []struct {
		name     string
		unit     uint64
		classLow uint8
		flags    uint32
		fail     ai.ActionType
		want     []string
	}{
		{name: "nil unit", classLow: uint8(object.ClassMonster)},
		{name: "not monster", unit: monsterScriptPauseHighUnit516090, want: []string{"class:7f9c86c94f90"}},
		{name: "dead", unit: monsterScriptPauseHighUnit516090, classLow: uint8(object.ClassMonster), flags: monsterScriptPauseDeadFlag516090,
			want: []string{"class:7f9c86c94f90", "flags:7f9c86c94f90"}},
		{name: "report push fails", unit: monsterScriptPauseHighUnit516090, classLow: uint8(object.ClassMonster), fail: ai.ACTION_REPORT,
			want: []string{"class:7f9c86c94f90", "flags:7f9c86c94f90", "push:7f9c86c94f90:32", "push:7f9c86c94f90:1", "frame", "arg:700000001:0:fffffff1"}},
		{name: "wait push fails", unit: monsterScriptPauseHighUnit516090, classLow: uint8(object.ClassMonster), fail: ai.ACTION_WAIT,
			want: []string{"class:7f9c86c94f90", "flags:7f9c86c94f90", "push:7f9c86c94f90:32", "arg:700000020:0:00000001", "push:7f9c86c94f90:1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newMonsterScriptPauseTrace516090()
			w.classLow, w.flags = tc.classLow, tc.flags
			if tc.fail != 0 {
				w.nilPush[uint32(tc.fail)] = true
			}
			monsterScriptPause516090(tc.unit, 1, w.hooks())
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %v, want %v", w.events, tc.want)
			}
		})
	}
}

func TestMonsterScriptPause516090NativeStack(t *testing.T) {
	s, unit, update := monsterScriptHitFixture515A30(t)
	s.SetFrame(100)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("object pointer is not above 4 GiB: %p", unit)
	}

	s.MonsterScriptPause516090(unit, 10)
	if update.AIStackInd != 2 ||
		update.AIStack[0].Type() != ai.ACTION_WAIT ||
		update.AIStack[1].Type() != ai.ACTION_REPORT || update.AIStack[1].ArgU32(0) != uint32(ai.ACTION_WAIT) ||
		update.AIStack[2].Type() != ai.ACTION_WAIT || update.AIStack[2].ArgU32(0) != 110 || !s.AI.StackChanged {
		t.Fatalf("pause stack = %#v index %d changed %t", update.AIStack[:3], update.AIStackInd, s.AI.StackChanged)
	}
	runtime.KeepAlive(unit)
}

func TestMonsterScriptPause516090RejectsMissingUpdate(t *testing.T) {
	s := &Server{}
	unit := &Object{ObjClass: object.ClassMonster}
	s.MonsterScriptPause516090(unit, 10)
}
