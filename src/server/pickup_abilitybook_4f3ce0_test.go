package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type pickupAbilityBookTestObject4F3CE0 struct {
	name     string
	flagsLow uint8
}

type pickupAbilityBookTestWorld4F3CE0 struct {
	gameFlagsResult int32
	arg3, arg4      int32
	defaultResult   int32
	events          []string
	faultAt         int
	afterUse        func(*pickupAbilityBookTestWorld4F3CE0)
}

func pickupAbilityBookTestName4F3CE0(obj *pickupAbilityBookTestObject4F3CE0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *pickupAbilityBookTestWorld4F3CE0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *pickupAbilityBookTestWorld4F3CE0) hooks() pickupAbilityBookHooks4F3CE0[*pickupAbilityBookTestObject4F3CE0] {
	return pickupAbilityBookHooks4F3CE0[*pickupAbilityBookTestObject4F3CE0]{
		gameFlagsCheck: func(flags uint32) int32 {
			w.event(fmt.Sprintf("game-flags=%08x", flags))
			return w.gameFlagsResult
		},
		useByNetCode: func(owner, item *pickupAbilityBookTestObject4F3CE0) {
			w.event("use:" + pickupAbilityBookTestName4F3CE0(owner) + ":" + pickupAbilityBookTestName4F3CE0(item))
			if w.afterUse != nil {
				w.afterUse(w)
			}
		},
		loadItemFlagsLow: func(item *pickupAbilityBookTestObject4F3CE0) uint8 {
			value := item.flagsLow
			w.event(fmt.Sprintf("flags:%s=%02x", pickupAbilityBookTestName4F3CE0(item), value))
			return value
		},
		loadArg4: func() int32 {
			w.event(fmt.Sprintf("arg4=%08x", uint32(w.arg4)))
			return w.arg4
		},
		loadArg3: func() int32 {
			w.event(fmt.Sprintf("arg3=%08x", uint32(w.arg3)))
			return w.arg3
		},
		defaultPickup: func(owner, item *pickupAbilityBookTestObject4F3CE0, arg3, arg4 int32) int32 {
			w.event(fmt.Sprintf("default:%s:%s:%08x:%08x", pickupAbilityBookTestName4F3CE0(owner), pickupAbilityBookTestName4F3CE0(item), uint32(arg3), uint32(arg4)))
			return w.defaultResult
		},
		audio: func(sound uint32, owner *pickupAbilityBookTestObject4F3CE0, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", sound, pickupAbilityBookTestName4F3CE0(owner), kind, code))
		},
	}
}

func pickupAbilityBookSuccessBuild4F3CE0() (*pickupAbilityBookTestWorld4F3CE0, *pickupAbilityBookTestObject4F3CE0, *pickupAbilityBookTestObject4F3CE0) {
	owner := &pickupAbilityBookTestObject4F3CE0{name: "owner"}
	item := &pickupAbilityBookTestObject4F3CE0{name: "item"}
	w := &pickupAbilityBookTestWorld4F3CE0{
		gameFlagsResult: -1,
		arg3:            math.MinInt32,
		arg4:            math.MaxInt32,
		defaultResult:   math.MinInt32,
	}
	return w, owner, item
}

func pickupAbilityBookSuccessTrace4F3CE0() []string {
	return []string{
		"game-flags=00001800",
		"use:owner:item",
		"flags:item=00",
		"arg4=7fffffff",
		"arg3=80000000",
		"default:owner:item:80000000:7fffffff",
		"audio:826:owner:0:00000000",
	}
}

func verifyPickupAbilityBookFaultPrefixes4F3CE0(t *testing.T, want []string) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, owner, item := pickupAbilityBookSuccessBuild4F3CE0()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			pickupAbilityBook4F3CE0(owner, item, w.hooks())
		})
	}
}

func TestPickupAbilityBook4F3CE0ForwardsFourArgsAndPreservesResult(t *testing.T) {
	w, owner, item := pickupAbilityBookSuccessBuild4F3CE0()
	if got := pickupAbilityBook4F3CE0(owner, item, w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	want := pickupAbilityBookSuccessTrace4F3CE0()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyPickupAbilityBookFaultPrefixes4F3CE0(t, want)
}

func TestPickupAbilityBook4F3CE0UseCanDestroyItem(t *testing.T) {
	owner := &pickupAbilityBookTestObject4F3CE0{name: "owner"}
	item := &pickupAbilityBookTestObject4F3CE0{name: "item"}
	w := &pickupAbilityBookTestWorld4F3CE0{gameFlagsResult: 1, arg3: 3, arg4: 4, defaultResult: -1}
	w.afterUse = func(*pickupAbilityBookTestWorld4F3CE0) {
		item.flagsLow = pickupAbilityBookDestroyedFlagLow4F3CE0
	}
	if got := pickupAbilityBook4F3CE0(owner, item, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{"game-flags=00001800", "use:owner:item", "flags:item=20"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPickupAbilityBook4F3CE0ZeroSkipsUseAndAudio(t *testing.T) {
	owner := &pickupAbilityBookTestObject4F3CE0{name: "owner"}
	item := &pickupAbilityBookTestObject4F3CE0{name: "item"}
	w := &pickupAbilityBookTestWorld4F3CE0{arg3: -3, arg4: -4}
	if got := pickupAbilityBook4F3CE0(owner, item, w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"game-flags=00001800",
		"flags:item=00",
		"arg4=fffffffc",
		"arg3=fffffffd",
		"default:owner:item:fffffffd:fffffffc",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPickupAbilityBook4F3CE0HasNoNilItemGuard(t *testing.T) {
	w := &pickupAbilityBookTestWorld4F3CE0{}
	owner := &pickupAbilityBookTestObject4F3CE0{name: "owner"}
	defer func() {
		if recover() == nil {
			t.Fatal("nil item did not preserve the flags-load fault")
		}
		want := []string{"game-flags=00001800"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	}()
	pickupAbilityBook4F3CE0(owner, nil, w.hooks())
}
