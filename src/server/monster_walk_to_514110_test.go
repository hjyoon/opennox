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

const monsterWalkToHighHandle514110 = uint64(0x7f9641bd3030)

type monsterWalkToTestWorld514110 struct {
	events        []string
	flags         uint32
	classLow      uint8
	pushResults   []uint64
	pushResultInd int
}

func (w *monsterWalkToTestWorld514110) hooks() monsterWalkToHooks514110[uint64, uint64] {
	return monsterWalkToHooks514110[uint64, uint64]{
		loadFlags: func(unit uint64) uint32 {
			w.events = append(w.events, fmt.Sprintf("flags:%x", unit))
			return w.flags
		},
		loadClassLow: func(unit uint64) uint8 {
			w.events = append(w.events, fmt.Sprintf("class:%x", unit))
			return w.classLow
		},
		clearActionStack: func(unit uint64) {
			w.events = append(w.events, fmt.Sprintf("clear:%x", unit))
		},
		pushAction: func(unit uint64, action uint32) uint64 {
			w.events = append(w.events, fmt.Sprintf("push:%x:%d", unit, action))
			var result uint64
			if w.pushResultInd < len(w.pushResults) {
				result = w.pushResults[w.pushResultInd]
			}
			w.pushResultInd++
			return result
		},
		storeArgBits: func(action uint64, index int, bits uint32) {
			w.events = append(w.events, fmt.Sprintf("store:%x:%d:%08x", action, index, bits))
		},
	}
}

func TestMonsterWalkTo514110ExactOrderAndNativeWidthHandles(t *testing.T) {
	w := &monsterWalkToTestWorld514110{
		classLow:    monsterWalkToMonsterClass514110,
		pushResults: []uint64{0x7f9642001010, 0x7f9642002020},
	}
	monsterWalkTo514110(monsterWalkToHighHandle514110, 0x450e4000, 0x45430000, w.hooks())

	want := []string{
		"flags:7f9641bd3030",
		"class:7f9641bd3030",
		"clear:7f9641bd3030",
		"push:7f9641bd3030:32",
		"store:7f9642001010:0:00000008",
		"push:7f9641bd3030:8",
		"store:7f9642002020:0:450e4000",
		"store:7f9642002020:1:45430000",
		"store:7f9642002020:2:00000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact oracle order %q", w.events, want)
	}
}

func TestMonsterWalkTo514110PreservesGatesAndFullStackBehavior(t *testing.T) {
	tests := []struct {
		name        string
		flags       uint32
		classLow    uint8
		pushResults []uint64
		want        []string
	}{
		{
			name:  "dead skips class load",
			flags: monsterWalkToBlockedFlag514110,
			want:  []string{"flags:7f9641bd3030"},
		},
		{
			name:     "non monster",
			classLow: 0xfd,
			want: []string{
				"flags:7f9641bd3030",
				"class:7f9641bd3030",
			},
		},
		{
			name:        "report push full still attempts move",
			classLow:    monsterWalkToMonsterClass514110,
			pushResults: []uint64{0, 0},
			want: []string{
				"flags:7f9641bd3030",
				"class:7f9641bd3030",
				"clear:7f9641bd3030",
				"push:7f9641bd3030:32",
				"push:7f9641bd3030:8",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := &monsterWalkToTestWorld514110{
				flags: tc.flags, classLow: tc.classLow, pushResults: tc.pushResults,
			}
			monsterWalkTo514110(monsterWalkToHighHandle514110, 1, 2, w.hooks())
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %q, want %q", w.events, tc.want)
			}
		})
	}
}

func TestMonsterWalkToNative514110BuildsExactActionStack(t *testing.T) {
	s := newMoverStateServer(t)
	update := new(MonsterUpdateData)
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_WAIT)}
	unit := &Object{
		ObjClass:     object.ClassMonster,
		UpdateData:   unsafe.Pointer(update),
		serverHandle: s.handle,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	x := math.Float32frombits(0x450e4000)
	y := math.Float32frombits(0x45430000)
	monsterWalkToNative514110(unit, x, y)

	if update.AIStackInd != 1 {
		t.Fatalf("AIStackInd = %d, want 1", update.AIStackInd)
	}
	report := &update.AIStack[0]
	if report.Type() != ai.ACTION_REPORT || report.Args != [4]uintptr{uintptr(ai.ACTION_FAR_MOVE_TO), 0, 0, 0} {
		t.Fatalf("REPORT = %s/%#v, want REPORT/FAR_MOVE_TO", report.Type(), report.Args)
	}
	move := &update.AIStack[1]
	if move.Type() != ai.ACTION_FAR_MOVE_TO {
		t.Fatalf("move action = %s, want FAR_MOVE_TO", move.Type())
	}
	if move.Args != [4]uintptr{uintptr(math.Float32bits(x)), uintptr(math.Float32bits(y)), 0, 0} {
		t.Fatalf("FAR_MOVE_TO args = %#v", move.Args)
	}
	if !s.AI.StackChanged {
		t.Fatal("native action-stack callbacks did not mark the stack changed")
	}
	runtime.KeepAlive(unit)
}

func TestMonsterWalkToNative514110PreservesOriginalFaultContract(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("null object did not preserve the original initial-load fault")
		}
	}()
	monsterWalkToNative514110(nil, 1, 2)
}
