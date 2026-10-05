package opennox

import (
	"fmt"
	"math"
	"runtime"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func aiVisibleDependencyPure546E0A() (*server.Object, *server.Object, *server.MonsterUpdateData, *server.AIStackItem) {
	update := new(server.MonsterUpdateData)
	unit := &server.Object{PosVec: types.Ptf(11, 23), UpdateData: unsafe.Pointer(update)}
	target := &server.Object{PosVec: types.Ptf(-43, 17)}
	slot := &update.AIStack[1]
	*slot = server.AIStackItem{Action: 48, Field5: 0x12345678}
	slot.SetArgs(types.Ptf(121, 231), unsafe.Pointer(target), uint32(0xfedcba98))
	return unit, target, update, slot
}

func aiVisibleDependencyPointBits546E0A(point types.Pointf) [2]uint32 {
	return [2]uint32{math.Float32bits(point.X), math.Float32bits(point.Y)}
}

func TestAIVisibleLocationDependency546E0ATruthAndCallOrder(t *testing.T) {
	for _, tc := range []struct {
		name      string
		hasTarget bool
		interacts bool
		clearRay  bool
		want      bool
	}{
		{"nil-clear", false, false, true, false},
		{"nil-blocked", false, false, false, true},
		{"unseen-clear", true, false, true, false},
		{"unseen-blocked", true, false, false, true},
		{"seen-clear", true, true, true, true},
		{"seen-blocked", true, true, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, target, update, slot := aiVisibleDependencyPure546E0A()
			if !tc.hasTarget {
				slot.Args[2] = 0
			}
			before := *slot
			wantPoint := slot.ArgPos(0)
			if tc.interacts {
				wantPoint = target.PosVec
			}
			var calls []string
			got := aiDependencyObjectAtVisibleLocation546E0A(unit, slot,
				func(from, selected *server.Object, flags int) bool {
					calls = append(calls, "interact")
					if from != unit || selected != target || flags != 0 {
						t.Fatal("interaction identity or original flags differ")
					}
					return tc.interacts
				},
				func(from, to types.Pointf, flags server.MapTraceFlags) bool {
					calls = append(calls, "ray")
					if from != unit.PosVec || to != wantPoint || flags != server.MapTraceFlag1 {
						t.Fatalf("ray arguments differ: %v -> %v flags=%v", from, to, flags)
					}
					return tc.clearRay
				})
			wantCalls := []string{"ray"}
			if tc.hasTarget {
				wantCalls = []string{"interact", "ray"}
			}
			if got != tc.want || !slices.Equal(calls, wantCalls) {
				t.Fatalf("result=%t calls=%v want=%t/%v", got, calls, tc.want, wantCalls)
			}
			wantSlot := before
			if tc.interacts {
				wantSlot.Args[0], wantSlot.Args[1] = uintptr(math.Float32bits(wantPoint.X)), uintptr(math.Float32bits(wantPoint.Y))
			}
			if *slot != wantSlot {
				t.Fatalf("original slot payload differs: got=%+v want=%+v", slot, wantSlot)
			}
			runtime.KeepAlive(target)
			runtime.KeepAlive(update)
			runtime.KeepAlive(unit)
		})
	}
}

func TestAIVisibleLocationDependency546E0ACachedSlotLiveTarget(t *testing.T) {
	for _, interacts := range []bool{false, true} {
		for _, replaceUpdate := range []bool{false, true} {
			t.Run(fmt.Sprintf("interact-%t/replace-update-%t", interacts, replaceUpdate), func(t *testing.T) {
				unit, target, update, slot := aiVisibleDependencyPure546E0A()
				replacement := &server.Object{PosVec: types.Ptf(-89, 71)}
				liveUpdate := new(server.MonsterUpdateData)
				liveUpdate.AIStack[1] = server.AIStackItem{Action: 48, Args: [4]uintptr{9, 8, 0, 7}, Field5: 6}
				otherBefore := liveUpdate.AIStack[1]
				before := *slot
				liveOrigin, savedPoint := types.Ptf(51, -61), types.Ptf(81, 91)
				wantPoint := savedPoint
				if interacts {
					wantPoint = replacement.PosVec
				}
				var calls []string
				got := aiDependencyObjectAtVisibleLocation546E0A(unit, slot,
					func(from, selected *server.Object, flags int) bool {
						calls = append(calls, "interact")
						if from != unit || selected != target || flags != 0 {
							t.Fatal("callback did not receive entry target")
						}
						unit.PosVec = liveOrigin
						slot.SetArgs(savedPoint, unsafe.Pointer(replacement), uint32(0xfedcba98))
						if replaceUpdate {
							unit.UpdateData = unsafe.Pointer(liveUpdate)
						}
						return interacts
					},
					func(from, to types.Pointf, flags server.MapTraceFlags) bool {
						calls = append(calls, "ray")
						if from != liveOrigin || to != wantPoint || flags != server.MapTraceFlag1 || slot.ArgPos(0) != wantPoint || slot.Args[2] != uintptr(unsafe.Pointer(replacement)) {
							t.Fatalf("ray did not use live origin/cached slot target: %v -> %v slot=%+v", from, to, slot)
						}
						// A ray callback cannot cause a second target/visibility read.
						slot.Args[0], slot.Args[2] = 0x80000000, 0
						unit.PosVec = types.Ptf(-201, -202)
						return true
					})
				if got != interacts || !slices.Equal(calls, []string{"interact", "ray"}) || slot.Args[0] != 0x80000000 || slot.Args[2] != 0 || slot.Args[3] != before.Args[3] || slot.Field5 != before.Field5 || slot.Action != before.Action || liveUpdate.AIStack[1] != otherBefore {
					t.Fatalf("cached slot/result or callback prefix differs: result=%t calls=%v slot=%+v", got, calls, slot)
				}
				runtime.KeepAlive(target)
				runtime.KeepAlive(replacement)
				runtime.KeepAlive(update)
				runtime.KeepAlive(liveUpdate)
				runtime.KeepAlive(unit)
			})
		}
	}
}

func TestAIVisibleLocationDependency546E0ARawFloatWords(t *testing.T) {
	for index, bits := range [][2]uint32{
		{0x80000000, 0x7fc12345},
		{0x7fa12345, 0x80000000},
		{0x7f800000, 0xff800000},
		{0x00000001, 0x80000001},
	} {
		for _, interacts := range []bool{false, true} {
			t.Run(fmt.Sprintf("words-%d/interact-%t", index, interacts), func(t *testing.T) {
				unit, target, update, slot := aiVisibleDependencyPure546E0A()
				target.PosVec = types.Ptf(math.Float32frombits(bits[0]), math.Float32frombits(bits[1]))
				unit.PosVec = target.PosVec
				high := uintptr(0x12345678)
				high <<= 32
				slot.Args[0], slot.Args[1] = high|uintptr(bits[0]), high|uintptr(bits[1])
				before := *slot
				rayCalls := 0
				got := aiDependencyObjectAtVisibleLocation546E0A(unit, slot,
					func(from, selected *server.Object, flags int) bool {
						if from != unit || selected != target || flags != 0 {
							t.Fatal("raw words changed native target/interaction flags")
						}
						return interacts
					},
					func(from, to types.Pointf, flags server.MapTraceFlags) bool {
						rayCalls++
						if aiVisibleDependencyPointBits546E0A(from) != bits || aiVisibleDependencyPointBits546E0A(to) != bits || flags != server.MapTraceFlag1 {
							t.Fatalf("float word bits changed: from=%08x to=%08x", aiVisibleDependencyPointBits546E0A(from), aiVisibleDependencyPointBits546E0A(to))
						}
						return true
					})
				wantSlot := before
				if interacts {
					// Original DWORD stores zero-extend copied words, but
					// a failed interaction leaves saved uintptr words intact.
					wantSlot.Args[0], wantSlot.Args[1] = uintptr(bits[0]), uintptr(bits[1])
				}
				if got != interacts || rayCalls != 1 || *slot != wantSlot {
					t.Fatalf("raw payload/result differs: result=%t rays=%d got=%+v want=%+v", got, rayCalls, slot, wantSlot)
				}
				runtime.KeepAlive(target)
				runtime.KeepAlive(update)
				runtime.KeepAlive(unit)
			})
		}
	}
}

func TestAIVisibleLocationDependency546E0AFaultPrefix(t *testing.T) {
	for _, name := range []string{
		"nil-slot", "nil-interaction", "interaction-panic",
		"live-target-cleared", "nil-ray-after-copy", "nil-unit",
	} {
		t.Run(name, func(t *testing.T) {
			unit, target, update, slot := aiVisibleDependencyPure546E0A()
			argUnit, argSlot := unit, slot
			var calls []string
			canInteract := func(from, selected *server.Object, flags int) bool {
				calls = append(calls, "interact")
				if from != unit || selected != target || flags != 0 {
					t.Fatal("faulting interaction prefix changed")
				}
				switch name {
				case "interaction-panic":
					panic("original interaction callback fault")
				case "live-target-cleared":
					slot.Args[2] = 0
				}
				return true
			}
			traceRay := func(types.Pointf, types.Pointf, server.MapTraceFlags) bool {
				calls = append(calls, "ray")
				return false
			}
			wantCalls := []string{"interact"}
			switch name {
			case "nil-slot":
				argSlot, wantCalls = nil, nil
			case "nil-interaction":
				canInteract, wantCalls = nil, nil
			case "nil-ray-after-copy":
				traceRay = nil
			case "nil-unit":
				argUnit, wantCalls = nil, nil
				slot.Args[2] = 0
			}
			wantSlot := *slot
			switch name {
			case "live-target-cleared":
				wantSlot.Args[2] = 0
			case "nil-ray-after-copy":
				wantSlot.Args[0], wantSlot.Args[1] = uintptr(math.Float32bits(target.PosVec.X)), uintptr(math.Float32bits(target.PosVec.Y))
			}
			panicked, returned := false, false
			func() {
				defer func() { panicked = recover() != nil }()
				aiDependencyObjectAtVisibleLocation546E0A(argUnit, argSlot, canInteract, traceRay)
				returned = true
			}()
			if !panicked || returned || !slices.Equal(calls, wantCalls) || *slot != wantSlot {
				t.Fatalf("original fault prefix swallowed/reordered: panic=%t returned=%t calls=%v want=%v slot=%+v want=%+v", panicked, returned, calls, wantCalls, slot, wantSlot)
			}
			runtime.KeepAlive(target)
			runtime.KeepAlive(update)
			runtime.KeepAlive(unit)
		})
	}
}
