package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	fleeUnitHandle515F70   = uint64(0x100000123)
	fleeTargetHandle515F70 = uint64(0x200000456)
	fleeUpdateHandle515F70 = uint64(0x300000789)
)

type fleeTrace515F70 struct {
	events    []string
	class     uint8
	flags     uint32
	update    uint64
	scheduled bool
	pushFail  map[uint32]bool
	frame     uint32
	posX      uint32
	posY      uint32
	args      map[uint64][4]uint32
}

func newFleeTrace515F70() *fleeTrace515F70 {
	return &fleeTrace515F70{
		class:    2,
		update:   fleeUpdateHandle515F70,
		pushFail: make(map[uint32]bool),
		frame:    0xfffffff0,
		posX:     0x7fa12345,
		posY:     0x80000000,
		args:     make(map[uint64][4]uint32),
	}
}

func (w *fleeTrace515F70) hooks() monsterScriptFleeHooks515F70[uint64, uint64, uint64] {
	return monsterScriptFleeHooks515F70[uint64, uint64, uint64]{
		loadClassLow: func(unit uint64) uint8 {
			if unit != fleeUnitHandle515F70 {
				panic("unit handle truncated")
			}
			w.events = append(w.events, "class")
			return w.class
		},
		loadFlags: func(unit uint64) uint32 {
			if unit != fleeUnitHandle515F70 {
				panic("unit handle truncated")
			}
			w.events = append(w.events, "flags")
			return w.flags
		},
		loadUpdate: func(unit uint64) uint64 {
			if unit != fleeUnitHandle515F70 {
				panic("unit handle truncated")
			}
			w.events = append(w.events, "update")
			return w.update
		},
		hasAction: func(update uint64, action uint32) bool {
			if update != fleeUpdateHandle515F70 || action != uint32(ai.ACTION_FLEE) {
				panic("update handle or action truncated")
			}
			w.events = append(w.events, "has-flee")
			return w.scheduled
		},
		pushAction: func(unit uint64, action uint32) uint64 {
			if unit != fleeUnitHandle515F70 {
				panic("unit handle truncated")
			}
			w.events = append(w.events, fmt.Sprintf("push:%d", action))
			if w.pushFail[action] {
				return 0
			}
			return uint64(0x400000000) | uint64(action)
		},
		loadFrame: func() uint32 {
			w.events = append(w.events, "frame")
			return w.frame
		},
		loadTargetPosXBits: func(target uint64) uint32 {
			if target != fleeTargetHandle515F70 {
				panic("target handle truncated")
			}
			w.events = append(w.events, "pos-x")
			return w.posX
		},
		loadTargetPosYBits: func(target uint64) uint32 {
			w.events = append(w.events, "pos-y")
			return w.posY
		},
		storeArgBits: func(item uint64, index int, bits uint32) {
			w.events = append(w.events, fmt.Sprintf("arg:%d:%d:%08x", uint32(item), index, bits))
			args := w.args[item]
			args[index] = bits
			w.args[item] = args
		},
	}
}

func TestMonsterScriptFleeFromOracleOrder515F70(t *testing.T) {
	w := newFleeTrace515F70()
	monsterScriptFleeFrom515F70(fleeUnitHandle515F70, fleeTargetHandle515F70, 0x30, w.hooks())
	want := []string{
		"class", "flags", "update", "has-flee",
		"push:32", "arg:32:0:00000018",
		"push:41", "frame", "arg:41:0:00000020",
		"push:24", "pos-x", "arg:24:0:7fa12345",
		"pos-y", "arg:24:2:00000000", "arg:24:1:80000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if w.args[0x400000020] != [4]uint32{24} ||
		w.args[0x400000029] != [4]uint32{0x20} ||
		w.args[0x400000018] != [4]uint32{0x7fa12345, 0x80000000, 0} {
		t.Fatalf("action arguments = %#v", w.args)
	}
}

func TestMonsterScriptFleeFromEligibility515F70(t *testing.T) {
	tests := []struct {
		name      string
		unit      uint64
		target    uint64
		class     uint8
		flags     uint32
		update    uint64
		scheduled bool
		want      []string
	}{
		{name: "nil unit", target: fleeTargetHandle515F70, class: 2},
		{name: "not monster", unit: fleeUnitHandle515F70, target: fleeTargetHandle515F70, want: []string{"class"}},
		{name: "dead", unit: fleeUnitHandle515F70, target: fleeTargetHandle515F70, class: 2, flags: 0x8000, want: []string{"class", "flags"}},
		{name: "nil update", unit: fleeUnitHandle515F70, target: fleeTargetHandle515F70, class: 2, want: []string{"class", "flags", "update"}},
		{name: "already fleeing", unit: fleeUnitHandle515F70, target: fleeTargetHandle515F70, class: 2, update: fleeUpdateHandle515F70, scheduled: true, want: []string{"class", "flags", "update", "has-flee"}},
		{name: "nil target", unit: fleeUnitHandle515F70, class: 2, update: fleeUpdateHandle515F70, want: []string{"class", "flags", "update", "has-flee"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newFleeTrace515F70()
			w.class, w.flags, w.update, w.scheduled = tc.class, tc.flags, tc.update, tc.scheduled
			monsterScriptFleeFrom515F70(tc.unit, tc.target, 10, w.hooks())
			if !reflect.DeepEqual(w.events, tc.want) || len(w.args) != 0 {
				t.Fatalf("rejected request mutated actions: events %v, args %#v", w.events, w.args)
			}
		})
	}
}

func TestMonsterScriptFleeFromPushFailures515F70(t *testing.T) {
	for _, action := range []ai.ActionType{ai.ACTION_REPORT, ai.DEPENDENCY_TIME, ai.ACTION_FLEE} {
		t.Run(action.String(), func(t *testing.T) {
			w := newFleeTrace515F70()
			w.pushFail[uint32(action)] = true
			monsterScriptFleeFrom515F70(fleeUnitHandle515F70, fleeTargetHandle515F70, 0x30, w.hooks())
			var pushed int
			for _, event := range w.events {
				if len(event) >= 5 && event[:5] == "push:" {
					pushed++
				}
			}
			if pushed != 3 || len(w.args) != 2 {
				t.Fatalf("push failure did not preserve later actions: events %v, args %#v", w.events, w.args)
			}
			for _, event := range w.events {
				if action == ai.DEPENDENCY_TIME && event == "frame" {
					t.Fatal("failed time push read the frame")
				}
				if action == ai.ACTION_FLEE && (event == "pos-x" || event == "pos-y") {
					t.Fatal("failed flee push read the target position")
				}
			}
		})
	}
}

func TestMonsterScriptFleeFromNative515F70(t *testing.T) {
	s, unit, update := monsterScriptHitFixture515A30(t)
	s.SetFrame(0xfffffff0)
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	target := &Object{PosVec: types.Ptf(math.Float32frombits(0x7fa12345), math.Float32frombits(0x80000000))}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unsafe.Pointer(target)) <= math.MaxUint32) {
		t.Fatalf("object pointers are not both above 4 GiB: %p %p", unit, target)
	}
	s.MonsterScriptFleeFrom515F70(unit, target, 0x30)
	if update.AIStackInd != 2 ||
		update.AIStack[0].Type() != ai.ACTION_REPORT || update.AIStack[0].ArgU32(0) != uint32(ai.ACTION_FLEE) ||
		update.AIStack[1].Type() != ai.DEPENDENCY_TIME || update.AIStack[1].ArgU32(0) != 0x20 ||
		update.AIStack[2].Type() != ai.ACTION_FLEE || update.AIStack[2].ArgU32(0) != 0x7fa12345 ||
		update.AIStack[2].ArgU32(1) != 0x80000000 || update.AIStack[2].ArgU32(2) != 0 || !s.AI.StackChanged {
		t.Fatalf("flee stack = %#v index %d changed %t", update.AIStack[:3], update.AIStackInd, s.AI.StackChanged)
	}
	runtime.KeepAlive(target)
	runtime.KeepAlive(unit)

	s.AI.StackChanged = false
	s.MonsterScriptFleeFrom515F70(unit, target, 10)
	if update.AIStackInd != 2 || s.AI.StackChanged {
		t.Fatal("already scheduled FLEE changed the stack")
	}
	unit.ObjFlags = object.FlagDead
	s.MonsterScriptFleeFrom515F70(unit, target, 10)
	if update.AIStackInd != 2 || s.AI.StackChanged {
		t.Fatal("dead monster changed the stack")
	}
}

func TestMonsterScriptFleeFromPreservesExistingStack515F70(t *testing.T) {
	s, unit, update := monsterScriptHitFixture515A30(t)
	s.SetFrame(3)
	target := &Object{PosVec: types.Ptf(10, -20)}
	s.MonsterScriptFleeFrom515F70(unit, target, -5)
	if update.AIStackInd != 3 || update.AIStack[0].Type() != ai.ACTION_WAIT ||
		update.AIStack[1].Type() != ai.ACTION_REPORT || update.AIStack[1].ArgU32(0) != uint32(ai.ACTION_FLEE) ||
		update.AIStack[2].Type() != ai.DEPENDENCY_TIME || update.AIStack[2].ArgU32(0) != 0xfffffffe ||
		update.AIStack[3].Type() != ai.ACTION_FLEE || update.AIStack[3].ArgU32(0) != math.Float32bits(10) ||
		update.AIStack[3].ArgU32(1) != math.Float32bits(-20) || update.AIStack[3].ArgU32(2) != 0 {
		t.Fatalf("stack after negative duration = %#v, index %d", update.AIStack[:4], update.AIStackInd)
	}
}
