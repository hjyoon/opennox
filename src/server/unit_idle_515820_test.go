package server

import (
	"fmt"
	"reflect"
	"testing"
)

const unitIdleHighHandle515820 = uint64(0x7fb3078c7430)

type unitIdleTestWorld515820 struct {
	events   []string
	faultAt  int
	classLow uint8
	flags    uint32
}

func (w *unitIdleTestWorld515820) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *unitIdleTestWorld515820) hooks() unitIdleHooks515820[uint64] {
	return unitIdleHooks515820[uint64]{
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
		pushAction: func(unit uint64, action uint32) {
			w.observe(fmt.Sprintf("push:%x:%d", unit, action))
		},
	}
}

func TestUnitIdle515820ExactOrderAndNativeWidthHandle(t *testing.T) {
	w := &unitIdleTestWorld515820{classLow: unitIdleMonsterClassLow515820}
	unitIdle515820(unitIdleHighHandle515820, w.hooks())

	want := []string{
		"class:7fb3078c7430",
		"flags:7fb3078c7430",
		"clear:7fb3078c7430",
		"push:7fb3078c7430:0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact oracle order %q", w.events, want)
	}
}

func TestUnitIdle515820ExactGates(t *testing.T) {
	tests := []struct {
		name       string
		unit       uint64
		classLow   uint8
		flags      uint32
		wantEvents []string
	}{
		{name: "null object"},
		{
			name: "non monster",
			unit: unitIdleHighHandle515820, classLow: 0xfd,
			wantEvents: []string{"class:7fb3078c7430"},
		},
		{
			name: "blocked flag",
			unit: unitIdleHighHandle515820, classLow: 0x82, flags: unitIdleBlockedFlag515820,
			wantEvents: []string{"class:7fb3078c7430", "flags:7fb3078c7430"},
		},
		{
			name: "unrelated high flags",
			unit: unitIdleHighHandle515820, classLow: 0x02, flags: 0xffff7fff,
			wantEvents: []string{
				"class:7fb3078c7430", "flags:7fb3078c7430",
				"clear:7fb3078c7430", "push:7fb3078c7430:0",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := &unitIdleTestWorld515820{classLow: tc.classLow, flags: tc.flags}
			unitIdle515820(tc.unit, w.hooks())
			if !reflect.DeepEqual(w.events, tc.wantEvents) {
				t.Fatalf("events = %q, want %q", w.events, tc.wantEvents)
			}
		})
	}
}

func TestUnitIdle515820AllFaultPrefixes(t *testing.T) {
	baseline := &unitIdleTestWorld515820{classLow: unitIdleMonsterClassLow515820}
	unitIdle515820(unitIdleHighHandle515820, baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := &unitIdleTestWorld515820{
				faultAt: faultAt, classLow: unitIdleMonsterClassLow515820,
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				unitIdle515820(unitIdleHighHandle515820, w.hooks())
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
