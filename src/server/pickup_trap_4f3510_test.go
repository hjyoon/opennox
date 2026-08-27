package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type pickupTrapTestObject4F3510 struct {
	name     string
	classLow uint8
	netCode  uint32
}

type pickupTrapTestWorld4F3510 struct {
	hasOwnerResult int32
	arg3           int32
	arg4           int32
	defaultResult  int32
	events         []string
	faultAt        int
	afterHasOwner  func(*pickupTrapTestWorld4F3510)
	afterDefault   func(*pickupTrapTestWorld4F3510)
}

func pickupTrapTestName4F3510(obj *pickupTrapTestObject4F3510) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *pickupTrapTestWorld4F3510) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *pickupTrapTestWorld4F3510) hooks(owner *pickupTrapTestObject4F3510) pickupTrapHooks4F3510[*pickupTrapTestObject4F3510] {
	return pickupTrapHooks4F3510[*pickupTrapTestObject4F3510]{
		hasOwner: func(item, gotOwner *pickupTrapTestObject4F3510) int32 {
			w.event(fmt.Sprintf("has-owner:%s:%s=%08x", pickupTrapTestName4F3510(item), pickupTrapTestName4F3510(gotOwner), uint32(w.hasOwnerResult)))
			if w.afterHasOwner != nil {
				w.afterHasOwner(w)
			}
			return w.hasOwnerResult
		},
		loadArg4: func() int32 {
			w.event(fmt.Sprintf("arg4=%08x", uint32(w.arg4)))
			return w.arg4
		},
		loadArg3: func() int32 {
			w.event(fmt.Sprintf("arg3=%08x", uint32(w.arg3)))
			return w.arg3
		},
		defaultPickup: func(gotOwner, item *pickupTrapTestObject4F3510, arg3, arg4 int32) int32 {
			w.event(fmt.Sprintf("default:%s:%s:%08x:%08x", pickupTrapTestName4F3510(gotOwner), pickupTrapTestName4F3510(item), uint32(arg3), uint32(arg4)))
			if w.afterDefault != nil {
				w.afterDefault(w)
			}
			return w.defaultResult
		},
		loadOwnerClassLow: func(gotOwner *pickupTrapTestObject4F3510) uint8 {
			value := gotOwner.classLow
			w.event(fmt.Sprintf("class-low:%s=%02x", pickupTrapTestName4F3510(gotOwner), value))
			return value
		},
		loadOwnerNetCode: func(gotOwner *pickupTrapTestObject4F3510) uint32 {
			value := gotOwner.netCode
			w.event(fmt.Sprintf("net-code:%s=%08x", pickupTrapTestName4F3510(gotOwner), value))
			return value
		},
		audio: func(id uint32, gotOwner *pickupTrapTestObject4F3510, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", id, pickupTrapTestName4F3510(gotOwner), kind, code))
		},
	}
}

func verifyPickupTrapFaultPrefixes4F3510(
	t *testing.T,
	want []string,
	build func() (*pickupTrapTestWorld4F3510, *pickupTrapTestObject4F3510, *pickupTrapTestObject4F3510),
) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, owner, item := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			pickupTrap4F3510(owner, item, w.hooks(owner))
		})
	}
}

func TestPickupTrap4F3510SuccessForwardsFourArgsAndExactResult(t *testing.T) {
	build := func() (*pickupTrapTestWorld4F3510, *pickupTrapTestObject4F3510, *pickupTrapTestObject4F3510) {
		owner := &pickupTrapTestObject4F3510{name: "owner", classLow: 4, netCode: 0x11111111}
		item := &pickupTrapTestObject4F3510{name: "item"}
		w := &pickupTrapTestWorld4F3510{
			hasOwnerResult: math.MinInt32,
			arg3:           0x12345678,
			arg4:           math.MinInt32,
			defaultResult:  math.MinInt32,
		}
		w.afterHasOwner = func(w *pickupTrapTestWorld4F3510) {
			w.arg3 = math.MaxInt32
			w.arg4 = -2
		}
		w.afterDefault = func(*pickupTrapTestWorld4F3510) {
			owner.netCode = 0x22222222
		}
		return w, owner, item
	}
	w, owner, item := build()
	if got := pickupTrap4F3510(owner, item, w.hooks(owner)); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	want := []string{
		"has-owner:item:owner=80000000",
		"arg4=fffffffe",
		"arg3=7fffffff",
		"default:owner:item:7fffffff:fffffffe",
		"audio:824:owner:0:00000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyPickupTrapFaultPrefixes4F3510(t, want, build)
}

func TestPickupTrap4F3510DefaultRejectReturnsExactZeroWithoutAudio(t *testing.T) {
	owner := &pickupTrapTestObject4F3510{name: "owner", classLow: 4, netCode: 0x11111111}
	item := &pickupTrapTestObject4F3510{name: "item"}
	w := &pickupTrapTestWorld4F3510{hasOwnerResult: 1, arg3: -3, arg4: -4}
	if got := pickupTrap4F3510(owner, item, w.hooks(owner)); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"has-owner:item:owner=00000001",
		"arg4=fffffffc",
		"arg3=fffffffd",
		"default:owner:item:fffffffd:fffffffc",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPickupTrap4F3510RejectedPlayerUsesLiveClassAndNetCode(t *testing.T) {
	build := func() (*pickupTrapTestWorld4F3510, *pickupTrapTestObject4F3510, *pickupTrapTestObject4F3510) {
		owner := &pickupTrapTestObject4F3510{name: "owner", classLow: 0, netCode: 0x11111111}
		item := &pickupTrapTestObject4F3510{name: "item"}
		w := &pickupTrapTestWorld4F3510{}
		w.afterHasOwner = func(*pickupTrapTestWorld4F3510) {
			owner.classLow = 0x84
			owner.netCode = 0xfedcba98
		}
		return w, owner, item
	}
	w, owner, item := build()
	if got := pickupTrap4F3510(owner, item, w.hooks(owner)); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"has-owner:item:owner=00000000",
		"class-low:owner=84",
		"net-code:owner=fedcba98",
		"audio:925:owner:2:fedcba98",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyPickupTrapFaultPrefixes4F3510(t, want, build)
}

func TestPickupTrap4F3510RejectedNonPlayerStopsBeforeNetCode(t *testing.T) {
	owner := &pickupTrapTestObject4F3510{name: "owner", classLow: 0x80, netCode: 0xfedcba98}
	item := &pickupTrapTestObject4F3510{name: "item"}
	w := &pickupTrapTestWorld4F3510{arg3: 3, arg4: 4, defaultResult: 1}
	if got := pickupTrap4F3510(owner, item, w.hooks(owner)); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"has-owner:item:owner=00000000",
		"class-low:owner=80",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}
