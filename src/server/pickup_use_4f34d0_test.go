package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type pickupUseTestObject4F34D0 struct {
	name  string
	flags uint32
}

type pickupUseTestWorld4F34D0 struct {
	useResult    int32
	defaultValue int32
	events       []string
	faultAt      int
	afterUse     func()
}

func pickupUseTestName4F34D0(obj *pickupUseTestObject4F34D0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *pickupUseTestWorld4F34D0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *pickupUseTestWorld4F34D0) hooks() pickupUseHooks4F34D0[*pickupUseTestObject4F34D0] {
	return pickupUseHooks4F34D0[*pickupUseTestObject4F34D0]{
		useByNetCode: func(owner, item *pickupUseTestObject4F34D0) int32 {
			w.event(fmt.Sprintf("use:%s:%s=%08x", pickupUseTestName4F34D0(owner), pickupUseTestName4F34D0(item), uint32(w.useResult)))
			if w.afterUse != nil {
				w.afterUse()
			}
			return w.useResult
		},
		loadFlagsLow: func(item *pickupUseTestObject4F34D0) uint8 {
			if item == nil {
				w.event("flags:nil:fault")
				panic("nil flags")
			}
			value := uint8(item.flags)
			w.event(fmt.Sprintf("flags:%s=%02x", item.name, value))
			return value
		},
		defaultPickup: func(owner, item *pickupUseTestObject4F34D0, arg3, arg4 int32) int32 {
			w.event(fmt.Sprintf("default:%s:%s:%08x:%08x", pickupUseTestName4F34D0(owner), pickupUseTestName4F34D0(item), uint32(arg3), uint32(arg4)))
			return w.defaultValue
		},
	}
}

func pickupUseOrdinaryTrace4F34D0() []string {
	return []string{
		"use:owner:item=80000000",
		"flags:item=00",
		"default:owner:item:80000000:7fffffff",
	}
}

func TestPickupUse4F34D0ExactTraceResultAndFaultPrefixes(t *testing.T) {
	want := pickupUseOrdinaryTrace4F34D0()
	build := func() (*pickupUseTestWorld4F34D0, *pickupUseTestObject4F34D0, *pickupUseTestObject4F34D0) {
		w := &pickupUseTestWorld4F34D0{useResult: math.MinInt32, defaultValue: math.MaxInt32}
		return w, &pickupUseTestObject4F34D0{name: "owner"}, &pickupUseTestObject4F34D0{name: "item"}
	}

	w, owner, item := build()
	if got := pickupUse4F34D0(owner, item, math.MinInt32, math.MaxInt32, w.hooks()); got != math.MaxInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MaxInt32))
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}

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
			pickupUse4F34D0(owner, item, math.MinInt32, math.MaxInt32, w.hooks())
		})
	}
}

func TestPickupUse4F34D0UseMutationDestroyedShortCircuits(t *testing.T) {
	owner := &pickupUseTestObject4F34D0{name: "owner"}
	item := &pickupUseTestObject4F34D0{name: "item", flags: 0x10000000}
	w := &pickupUseTestWorld4F34D0{useResult: math.MaxInt32, defaultValue: math.MinInt32}
	w.afterUse = func() { item.flags |= uint32(pickupUseDestroyedFlagsLow4F34D0) }

	if got := pickupUse4F34D0(owner, item, -17, -23, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{"use:owner:item=7fffffff", "flags:item=20"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPickupUse4F34D0OnlyDestroyedLowBitSkipsDefault(t *testing.T) {
	for _, tc := range []struct {
		name  string
		flags uint32
		want  int32
	}{
		{name: "high-bit-only", flags: 0x20000000, want: math.MinInt32},
		{name: "destroyed-low-bit", flags: 0x00000020, want: 1},
		{name: "both", flags: 0x20000020, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &pickupUseTestWorld4F34D0{useResult: -7, defaultValue: math.MinInt32}
			owner := &pickupUseTestObject4F34D0{name: "owner"}
			item := &pickupUseTestObject4F34D0{name: "item", flags: tc.flags}
			if got := pickupUse4F34D0(owner, item, -17, -23, w.hooks()); got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			wantCalls := 3
			if tc.want == 1 {
				wantCalls = 2
			}
			if len(w.events) != wantCalls {
				t.Fatalf("events = %v, want %d calls", w.events, wantCalls)
			}
		})
	}
}

func TestPickupUse4F34D0HasNoWrapperNilGuard(t *testing.T) {
	w := &pickupUseTestWorld4F34D0{useResult: 1}
	owner := &pickupUseTestObject4F34D0{name: "owner"}
	defer func() {
		if got := recover(); got != "nil flags" {
			t.Fatalf("panic = %v, want nil flags", got)
		}
		want := []string{"use:owner:nil=00000001", "flags:nil:fault"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	}()
	pickupUse4F34D0(owner, nil, 0, 0, w.hooks())
}
