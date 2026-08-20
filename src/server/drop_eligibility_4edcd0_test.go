package server

import (
	"fmt"
	"reflect"
	"testing"
)

type dropEligibilityTestObject4EDCD0 struct {
	name     string
	classLow uint8
	flags    uint32
}

type dropEligibilityTestWorld4EDCD0 struct {
	owner  *dropEligibilityTestObject4EDCD0
	item   *dropEligibilityTestObject4EDCD0
	events []string
	fault  int
}

func (w *dropEligibilityTestWorld4EDCD0) event(name string) {
	w.events = append(w.events, name)
	if w.fault != 0 && len(w.events) == w.fault {
		panic(name)
	}
}

func dropEligibilityTestName4EDCD0(obj *dropEligibilityTestObject4EDCD0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *dropEligibilityTestWorld4EDCD0) hooks() dropEligibilityHooks4EDCD0[
	*dropEligibilityTestObject4EDCD0,
	*dropEligibilityTestObject4EDCD0,
] {
	return dropEligibilityHooks4EDCD0[
		*dropEligibilityTestObject4EDCD0,
		*dropEligibilityTestObject4EDCD0,
	]{
		loadItemArg: func() *dropEligibilityTestObject4EDCD0 {
			w.event("item-arg:" + dropEligibilityTestName4EDCD0(w.item))
			return w.item
		},
		loadItemFlags: func(item *dropEligibilityTestObject4EDCD0) uint32 {
			w.event(fmt.Sprintf("item-flags:%s:%08x", dropEligibilityTestName4EDCD0(item), item.flags))
			return item.flags
		},
		loadOwnerArg: func() *dropEligibilityTestObject4EDCD0 {
			w.event("owner-arg:" + dropEligibilityTestName4EDCD0(w.owner))
			return w.owner
		},
		loadOwnerClassLow: func(owner *dropEligibilityTestObject4EDCD0) uint8 {
			w.event(fmt.Sprintf("owner-class:%s:%02x", dropEligibilityTestName4EDCD0(owner), owner.classLow))
			return owner.classLow
		},
	}
}

func TestDropEligibility4EDCD0ConstantsAndBranches(t *testing.T) {
	if dropEligibilityDestroyed4EDCD0 != 0x20 || dropEligibilityUnitClass4EDCD0 != 0x06 || dropEligibilityNoDrop4EDCD0 != 0x10000000 {
		t.Fatalf("constants = %#x/%#x/%#x", dropEligibilityDestroyed4EDCD0, dropEligibilityUnitClass4EDCD0, dropEligibilityNoDrop4EDCD0)
	}
	tests := []struct {
		name     string
		classLow uint8
		flags    uint32
		want     int32
	}{
		{name: "destroyed-precedes-no-drop", classLow: 0x06, flags: 0x10000020, want: 1},
		{name: "non-unit-owner", classLow: 0xf8, flags: 0x10000000, want: 1},
		{name: "player-without-no-drop", classLow: 0x04, flags: 0x80000000, want: 1},
		{name: "monster-without-no-drop", classLow: 0x02, flags: 0, want: 1},
		{name: "player-no-drop", classLow: 0x04, flags: 0x10000000, want: 0},
		{name: "monster-no-drop", classLow: 0x02, flags: 0x10000000, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := &dropEligibilityTestWorld4EDCD0{
				owner: &dropEligibilityTestObject4EDCD0{name: "owner", classLow: tc.classLow},
				item:  &dropEligibilityTestObject4EDCD0{name: "item", flags: tc.flags},
			}
			if got := dropEligibility4EDCD0(w.hooks()); got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestDropEligibility4EDCD0ShortCircuitAndFaultPrefixes(t *testing.T) {
	tests := []struct {
		name  string
		world func() *dropEligibilityTestWorld4EDCD0
		want  []string
	}{
		{
			name: "destroyed-does-not-load-owner",
			world: func() *dropEligibilityTestWorld4EDCD0 {
				return &dropEligibilityTestWorld4EDCD0{
					item: &dropEligibilityTestObject4EDCD0{name: "item", flags: 0x20},
				}
			},
			want: []string{"item-arg:item", "item-flags:item:00000020"},
		},
		{
			name: "non-unit-stops-after-class",
			world: func() *dropEligibilityTestWorld4EDCD0 {
				return &dropEligibilityTestWorld4EDCD0{
					owner: &dropEligibilityTestObject4EDCD0{name: "owner", classLow: 0x80},
					item:  &dropEligibilityTestObject4EDCD0{name: "item", flags: 0x10000000},
				}
			},
			want: []string{
				"item-arg:item",
				"item-flags:item:10000000",
				"owner-arg:owner",
				"owner-class:owner:80",
			},
		},
		{
			name: "unit-no-drop-full-prefix",
			world: func() *dropEligibilityTestWorld4EDCD0 {
				return &dropEligibilityTestWorld4EDCD0{
					owner: &dropEligibilityTestObject4EDCD0{name: "owner", classLow: 0x02},
					item:  &dropEligibilityTestObject4EDCD0{name: "item", flags: 0x10000000},
				}
			},
			want: []string{
				"item-arg:item",
				"item-flags:item:10000000",
				"owner-arg:owner",
				"owner-class:owner:02",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := tc.world()
			_ = dropEligibility4EDCD0(w.hooks())
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %v, want %v", w.events, tc.want)
			}
			for fault := 1; fault <= len(tc.want); fault++ {
				w := tc.world()
				w.fault = fault
				func() {
					defer func() {
						if got := recover(); got != tc.want[fault-1] {
							t.Fatalf("fault %d panic = %v, want %q", fault, got, tc.want[fault-1])
						}
					}()
					_ = dropEligibility4EDCD0(w.hooks())
				}()
				if !reflect.DeepEqual(w.events, tc.want[:fault]) {
					t.Fatalf("fault %d events = %v, want %v", fault, w.events, tc.want[:fault])
				}
			}
		})
	}
}
