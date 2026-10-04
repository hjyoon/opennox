package server

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterLookAtTestWorld5125A0 struct {
	events     []string
	class      uint8
	flags      uint32
	x, y       float32
	pushResult uint64
}

func (w *monsterLookAtTestWorld5125A0) hooks() monsterLookAtHooks5125A0[uint64, uint64] {
	return monsterLookAtHooks5125A0[uint64, uint64]{
		directionToAngle: func(direction int32) uint32 {
			w.events = append(w.events, fmt.Sprintf("angle:%d", direction))
			return 160
		},
		loadClassLow: func(unit uint64) uint8 {
			w.events = append(w.events, fmt.Sprintf("class:%x", unit))
			return w.class
		},
		loadFlags: func(unit uint64) uint32 {
			w.events = append(w.events, fmt.Sprintf("flags:%x", unit))
			return w.flags
		},
		loadCosine: func(angle uint32) float32 {
			w.events = append(w.events, fmt.Sprintf("cosine:%d", angle))
			return 0.5
		},
		loadSine: func(angle uint32) float32 {
			w.events = append(w.events, fmt.Sprintf("sine:%d", angle))
			return -0.25
		},
		loadX: func(unit uint64) float32 {
			w.events = append(w.events, fmt.Sprintf("x:%x", unit))
			return w.x
		},
		loadY: func(unit uint64) float32 {
			w.events = append(w.events, fmt.Sprintf("y:%x", unit))
			return w.y
		},
		pushAction: func(unit uint64, action uint32) uint64 {
			w.events = append(w.events, fmt.Sprintf("push:%x:%d", unit, action))
			w.x, w.y = 1000, 2000 // Existing action cancellation can change the unit.
			return w.pushResult
		},
		storeArgBits: func(action uint64, index int, bits uint32) {
			w.events = append(w.events, fmt.Sprintf("store:%x:%d:%08x", action, index, bits))
		},
		actionResult: func(action uint64) uintptr { return uintptr(action) },
	}
}

func TestMonsterLookAt5125A0ExactOrderAndNativeHandles(t *testing.T) {
	w := &monsterLookAtTestWorld5125A0{class: 2, x: 12.5, y: -8.25, pushResult: 0x7f3527c00120}
	result := monsterLookAt5125A0(uint64(0x7f3527bef8a0), -7, w.hooks())
	want := []string{
		"angle:-7", "class:7f3527bef8a0", "flags:7f3527bef8a0",
		"cosine:160", "x:7f3527bef8a0", "sine:160", "y:7f3527bef8a0",
		"push:7f3527bef8a0:25",
		"store:7f3527c00120:0:418c0000", "store:7f3527c00120:1:c12c0000",
	}
	if result != uintptr(w.pushResult) || !reflect.DeepEqual(w.events, want) {
		t.Fatalf("result/events = %x/%q want %x/%q", result, w.events, w.pushResult, want)
	}
}

func TestMonsterLookAt5125A0GateAndPushRejectionOrder(t *testing.T) {
	for _, tc := range []struct {
		name   string
		class  uint8
		flags  uint32
		result uintptr
		events []string
	}{
		{"nonmonster skips flags", 0xfd, 0x8000, 160, []string{"angle:0", "class:1"}},
		{"dead skips coordinates", 2, 0x8000, 160, []string{"angle:0", "class:1", "flags:1"}},
		{"push rejection retains computed coordinates", 2, 0, 0, []string{
			"angle:0", "class:1", "flags:1", "cosine:160", "x:1", "sine:160", "y:1", "push:1:25",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &monsterLookAtTestWorld5125A0{class: tc.class, flags: tc.flags}
			if result := monsterLookAt5125A0(uint64(1), 0, w.hooks()); result != tc.result || !reflect.DeepEqual(w.events, tc.events) {
				t.Fatalf("result/events = %x/%q want %x/%q", result, w.events, tc.result, tc.events)
			}
		})
	}
}

func TestMonsterLookAt5125A0PreservesFinalCoordinateSpill(t *testing.T) {
	w := &monsterLookAtTestWorld5125A0{class: 2, x: -7, y: 7, pushResult: 2}
	hooks := w.hooks()
	hooks.loadCosine = func(uint32) float32 { return math.Float32frombits(0x3f3504f3) }
	hooks.loadSine = func(uint32) float32 { return math.Float32frombits(0xbf3504f3) }
	var got [2]uint32
	hooks.storeArgBits = func(_ uint64, index int, bits uint32) { got[index] = bits }
	monsterLookAt5125A0(uint64(1), 0, hooks)
	if want := [2]uint32{0x3d918bf0, 0xbd918bf0}; got != want {
		t.Fatalf("coordinate spill=%08x want %08x; each product must not spill early", got, want)
	}
}

func TestMonsterLookAtNative5125A0StackBoundaries(t *testing.T) {
	for _, mode := range []string{"idle", "wait", "dead head", "full"} {
		t.Run(mode, func(t *testing.T) {
			srv := newMoverStateServer(t)
			update := new(MonsterUpdateData)
			unit := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update), serverHandle: srv.handle, PosVec: types.Ptf(-7, 7), Direction1: 64, Direction2: 64}
			update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{9, 8, 7, 6}}
			switch mode {
			case "idle":
				update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
			case "dead head":
				update.AIStack[0].Action = uint32(ai.ACTION_DEAD)
			case "full":
				update.AIStackInd = int8(len(update.AIStack) - 1)
			}
			oldStack, oldIndex := update.AIStack, update.AIStackInd
			result := MonsterLookAt5125A0(unit, 8)
			if mode == "dead head" || mode == "full" {
				if result != 0 || update.AIStack != oldStack || update.AIStackInd != oldIndex || srv.AI.StackChanged {
					t.Fatal("rejected push changed the stack or returned an action")
				}
				return
			}
			wantIndex := int8(1)
			if mode == "idle" {
				wantIndex = 0
			}
			if update.AIStackInd != wantIndex || result != uintptr(unsafe.Pointer(&update.AIStack[wantIndex])) || !srv.AI.StackChanged {
				t.Fatalf("native result/index = %x/%d", result, update.AIStackInd)
			}
			if mode == "wait" && update.AIStack[0] != oldStack[0] {
				t.Fatal("LookAt cleared the existing WAIT")
			}
			head := update.AIStackHead()
			if head.Type() != ai.ACTION_FACE_LOCATION || head.Args != [4]uintptr{0x3d918bf0, 0x41612318, 0, 0} || unit.Direction1 != 64 || unit.Direction2 != 64 {
				t.Fatalf("face action/direct directions = %+v/%d/%d", head, unit.Direction1, unit.Direction2)
			}
		})
	}
}

func TestMonsterLookAtNative5125A0EarlyGatesNeedNoUpdateOrServer(t *testing.T) {
	for _, tc := range []struct {
		name  string
		class object.Class
		flags object.Flags
	}{
		{"player", object.ClassPlayer, 0},
		{"nonmonster high class bit", object.ClassImmobile, 0},
		{"dead monster", object.ClassMonster, object.FlagDead},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit := &Object{ObjClass: tc.class, ObjFlags: tc.flags}
			if result := MonsterLookAt5125A0(unit, 0); result != 160 {
				t.Fatalf("gated residual=%d want 160", result)
			}
		})
	}
}

func TestMonsterLookAtNative5125A0FaultAndDirectionContracts(t *testing.T) {
	for _, direction := range []int32{0, -1, 9} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			defer func() {
				failure := recover()
				if failure == nil || direction != 0 && !strings.Contains(fmt.Sprint(failure), "direction") {
					t.Fatalf("fault=%v: direction mapping must precede the null object load", failure)
				}
			}()
			MonsterLookAt5125A0(nil, direction)
		})
	}
}
