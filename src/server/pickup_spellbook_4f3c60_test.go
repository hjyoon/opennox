package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type pickupSpellBookTestObject4F3C60 struct {
	name        string
	flagsLow    uint8
	subClassLow uint8
}

type pickupSpellBookTestWorld4F3C60 struct {
	gameFlagsResult int32
	arg3, arg4      int32
	defaultResult   int32
	events          []string
	faultAt         int
	afterUse        func(*pickupSpellBookTestWorld4F3C60)
	afterDefault    func(*pickupSpellBookTestWorld4F3C60)
}

func pickupSpellBookTestName4F3C60(obj *pickupSpellBookTestObject4F3C60) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *pickupSpellBookTestWorld4F3C60) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *pickupSpellBookTestWorld4F3C60) hooks() pickupSpellBookHooks4F3C60[*pickupSpellBookTestObject4F3C60] {
	return pickupSpellBookHooks4F3C60[*pickupSpellBookTestObject4F3C60]{
		gameFlagsCheck: func(flags uint32) int32 {
			w.event(fmt.Sprintf("game-flags=%08x", flags))
			return w.gameFlagsResult
		},
		useByNetCode: func(owner, item *pickupSpellBookTestObject4F3C60) {
			w.event("use:" + pickupSpellBookTestName4F3C60(owner) + ":" + pickupSpellBookTestName4F3C60(item))
			if w.afterUse != nil {
				w.afterUse(w)
			}
		},
		loadItemFlagsLow: func(item *pickupSpellBookTestObject4F3C60) uint8 {
			value := item.flagsLow
			w.event(fmt.Sprintf("flags:%s=%02x", pickupSpellBookTestName4F3C60(item), value))
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
		defaultPickup: func(owner, item *pickupSpellBookTestObject4F3C60, arg3, arg4 int32) int32 {
			w.event(fmt.Sprintf("default:%s:%s:%08x:%08x", pickupSpellBookTestName4F3C60(owner), pickupSpellBookTestName4F3C60(item), uint32(arg3), uint32(arg4)))
			if w.afterDefault != nil {
				w.afterDefault(w)
			}
			return w.defaultResult
		},
		loadItemSubClassLow: func(item *pickupSpellBookTestObject4F3C60) uint8 {
			value := item.subClassLow
			w.event(fmt.Sprintf("subclass:%s=%02x", pickupSpellBookTestName4F3C60(item), value))
			return value
		},
		audio: func(sound uint32, owner *pickupSpellBookTestObject4F3C60, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", sound, pickupSpellBookTestName4F3C60(owner), kind, code))
		},
	}
}

func pickupSpellBookSuccessBuild4F3C60() (*pickupSpellBookTestWorld4F3C60, *pickupSpellBookTestObject4F3C60, *pickupSpellBookTestObject4F3C60) {
	owner := &pickupSpellBookTestObject4F3C60{name: "owner"}
	item := &pickupSpellBookTestObject4F3C60{name: "item", subClassLow: 0x80}
	w := &pickupSpellBookTestWorld4F3C60{
		gameFlagsResult: -1,
		arg3:            math.MinInt32,
		arg4:            math.MaxInt32,
		defaultResult:   math.MinInt32,
	}
	w.afterDefault = func(*pickupSpellBookTestWorld4F3C60) {
		item.subClassLow = 0x81
	}
	return w, owner, item
}

func pickupSpellBookSuccessTrace4F3C60() []string {
	return []string{
		"game-flags=00001800",
		"use:owner:item",
		"flags:item=00",
		"arg4=7fffffff",
		"arg3=80000000",
		"default:owner:item:80000000:7fffffff",
		"subclass:item=81",
		"audio:826:owner:0:00000000",
	}
}

func verifyPickupSpellBookFaultPrefixes4F3C60(t *testing.T, want []string) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, owner, item := pickupSpellBookSuccessBuild4F3C60()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			pickupSpellBook4F3C60(owner, item, w.hooks())
		})
	}
}

func TestPickupSpellBook4F3C60ForwardsFourArgsAndReloadsSubclass(t *testing.T) {
	w, owner, item := pickupSpellBookSuccessBuild4F3C60()
	if got := pickupSpellBook4F3C60(owner, item, w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	want := pickupSpellBookSuccessTrace4F3C60()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyPickupSpellBookFaultPrefixes4F3C60(t, want)
}

func TestPickupSpellBook4F3C60UseCanDestroyItem(t *testing.T) {
	owner := &pickupSpellBookTestObject4F3C60{name: "owner"}
	item := &pickupSpellBookTestObject4F3C60{name: "item"}
	w := &pickupSpellBookTestWorld4F3C60{gameFlagsResult: 1, arg3: 3, arg4: 4, defaultResult: -1}
	w.afterUse = func(*pickupSpellBookTestWorld4F3C60) {
		item.flagsLow = pickupSpellBookDestroyedFlagLow4F3C60
	}
	if got := pickupSpellBook4F3C60(owner, item, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{"game-flags=00001800", "use:owner:item", "flags:item=20"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPickupSpellBook4F3C60ZeroSkipsUseSubclassAndAudio(t *testing.T) {
	owner := &pickupSpellBookTestObject4F3C60{name: "owner"}
	item := &pickupSpellBookTestObject4F3C60{name: "item", subClassLow: 1}
	w := &pickupSpellBookTestWorld4F3C60{arg3: -3, arg4: -4}
	if got := pickupSpellBook4F3C60(owner, item, w.hooks()); got != 0 {
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

func TestPickupSpellBook4F3C60InventoryAudioWithoutSpellBit(t *testing.T) {
	owner := &pickupSpellBookTestObject4F3C60{name: "owner"}
	item := &pickupSpellBookTestObject4F3C60{name: "item", subClassLow: 0x80}
	w := &pickupSpellBookTestWorld4F3C60{defaultResult: 7}
	if got := pickupSpellBook4F3C60(owner, item, w.hooks()); got != 7 {
		t.Fatalf("result = %d, want 7", got)
	}
	if got := w.events[len(w.events)-1]; got != "audio:828:owner:0:00000000" {
		t.Fatalf("last event = %q, want inventory audio", got)
	}
}

func TestPickupSpellBook4F3C60HasNoNilItemGuard(t *testing.T) {
	w := &pickupSpellBookTestWorld4F3C60{}
	owner := &pickupSpellBookTestObject4F3C60{name: "owner"}
	defer func() {
		if recover() == nil {
			t.Fatal("nil item did not preserve the flags-load fault")
		}
		want := []string{"game-flags=00001800"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	}()
	pickupSpellBook4F3C60(owner, nil, w.hooks())
}
