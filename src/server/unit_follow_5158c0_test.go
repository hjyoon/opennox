package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	unitFollowHighUnit5158C0   = uint64(0x7f4c6e001490)
	unitFollowHighTarget5158C0 = uint64(0x7f4c6df743d0)
	unitFollowHighAction5158C0 = uint64(0x7f4c70000030)
)

type unitFollowTestWorld5158C0 struct {
	events       []string
	faultAt      int
	classLow     uint8
	flags        uint32
	pushResult   uint64
	posXBits     uint32
	posYBits     uint32
	mutatePush   bool
	storedBits   [2]uint32
	storedTarget uint64
}

func (w *unitFollowTestWorld5158C0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *unitFollowTestWorld5158C0) hooks() unitFollowHooks5158C0[uint64, uint64] {
	return unitFollowHooks5158C0[uint64, uint64]{
		loadClassLow: func(unit uint64) uint8 {
			w.observe(fmt.Sprintf("class:%x", unit))
			return w.classLow
		},
		loadFlags: func(unit uint64) uint32 {
			w.observe(fmt.Sprintf("flags:%x", unit))
			return w.flags
		},
		clearActionStack: func(unit uint64) {
			w.observe(fmt.Sprintf("clear:%x", unit))
		},
		pushAction: func(unit uint64, action uint32) uint64 {
			w.observe(fmt.Sprintf("push:%x:%d", unit, action))
			if w.mutatePush {
				w.posXBits = 0x7fc00001
				w.posYBits = 0x80000000
			}
			return w.pushResult
		},
		loadPosXBits: func(target uint64) uint32 {
			w.observe(fmt.Sprintf("load-x:%x", target))
			return w.posXBits
		},
		storeArgBits: func(action uint64, index int, bits uint32) {
			w.observe(fmt.Sprintf("store-%d:%x:%08x", index, action, bits))
			w.storedBits[index] = bits
		},
		loadPosYBits: func(target uint64) uint32 {
			w.observe(fmt.Sprintf("load-y:%x", target))
			return w.posYBits
		},
		storeTarget: func(action uint64, index int, target uint64) {
			w.observe(fmt.Sprintf("store-target-%d:%x:%x", index, action, target))
			w.storedTarget = target
		},
	}
}

func newUnitFollowTestWorld5158C0() *unitFollowTestWorld5158C0 {
	return &unitFollowTestWorld5158C0{
		classLow:   unitFollowMonsterClassLow5158C0,
		pushResult: unitFollowHighAction5158C0,
		posXBits:   0x3f800001,
		posYBits:   0xc0200001,
	}
}

func TestUnitFollow5158C0ExactOrderAndNativeWidthHandles(t *testing.T) {
	w := newUnitFollowTestWorld5158C0()
	unitFollow5158C0(unitFollowHighUnit5158C0, unitFollowHighTarget5158C0, w.hooks())

	want := []string{
		"class:7f4c6e001490",
		"flags:7f4c6e001490",
		"clear:7f4c6e001490",
		"push:7f4c6e001490:3",
		"load-x:7f4c6df743d0",
		"store-0:7f4c70000030:3f800001",
		"load-y:7f4c6df743d0",
		"store-1:7f4c70000030:c0200001",
		"store-target-2:7f4c70000030:7f4c6df743d0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact oracle order %q", w.events, want)
	}
	if w.storedBits != [2]uint32{0x3f800001, 0xc0200001} || w.storedTarget != unitFollowHighTarget5158C0 {
		t.Fatalf("stored payload = %#v/%#x", w.storedBits, w.storedTarget)
	}
}

func TestUnitFollow5158C0ExactGates(t *testing.T) {
	tests := []struct {
		name       string
		unit       uint64
		target     uint64
		classLow   uint8
		flags      uint32
		pushResult uint64
		wantEvents []string
	}{
		{name: "null unit", target: unitFollowHighTarget5158C0},
		{name: "null target", unit: unitFollowHighUnit5158C0},
		{
			name: "non monster", unit: unitFollowHighUnit5158C0, target: unitFollowHighTarget5158C0,
			classLow: 0xfd, wantEvents: []string{"class:7f4c6e001490"},
		},
		{
			name: "same object", unit: unitFollowHighUnit5158C0, target: unitFollowHighUnit5158C0,
			classLow: 0x82, wantEvents: []string{"class:7f4c6e001490"},
		},
		{
			name: "blocked flag", unit: unitFollowHighUnit5158C0, target: unitFollowHighTarget5158C0,
			classLow: 0x02, flags: unitFollowBlockedFlag5158C0,
			wantEvents: []string{"class:7f4c6e001490", "flags:7f4c6e001490"},
		},
		{
			name: "push failure", unit: unitFollowHighUnit5158C0, target: unitFollowHighTarget5158C0,
			classLow: 0x02,
			wantEvents: []string{
				"class:7f4c6e001490", "flags:7f4c6e001490",
				"clear:7f4c6e001490", "push:7f4c6e001490:3",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newUnitFollowTestWorld5158C0()
			w.classLow, w.flags, w.pushResult = tc.classLow, tc.flags, tc.pushResult
			unitFollow5158C0(tc.unit, tc.target, w.hooks())
			if !reflect.DeepEqual(w.events, tc.wantEvents) {
				t.Fatalf("events = %q, want %q", w.events, tc.wantEvents)
			}
		})
	}
}

func TestUnitFollow5158C0LoadsLiveTargetAfterPush(t *testing.T) {
	w := newUnitFollowTestWorld5158C0()
	w.mutatePush = true
	unitFollow5158C0(unitFollowHighUnit5158C0, unitFollowHighTarget5158C0, w.hooks())
	if w.storedBits != [2]uint32{0x7fc00001, 0x80000000} {
		t.Fatalf("stored position bits = %#v, want post-push target values", w.storedBits)
	}
}

func TestUnitFollow5158C0AllFaultPrefixes(t *testing.T) {
	baseline := newUnitFollowTestWorld5158C0()
	unitFollow5158C0(unitFollowHighUnit5158C0, unitFollowHighTarget5158C0, baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newUnitFollowTestWorld5158C0()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				unitFollow5158C0(unitFollowHighUnit5158C0, unitFollowHighTarget5158C0, w.hooks())
			}()
			if recovered == nil {
				t.Fatal("fault sentinel was not recovered")
			}
			if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
				t.Fatalf("events = %q, want fault prefix %q", w.events, prefix)
			}
		})
	}
}
