package server

import (
	"fmt"
	"reflect"
	"testing"
)

type crownCollideTestObject4EBB50 struct {
	name  string
	flags uint32
	class uint8
}

type crownCollideTestState4EBB50 struct {
	events       []string
	pickupResult uint32
	onFlags      func()
	onClass      func()
}

func (s *crownCollideTestState4EBB50) hooks() crownCollideHooks4EBB50[*crownCollideTestObject4EBB50] {
	return crownCollideHooks4EBB50[*crownCollideTestObject4EBB50]{
		loadFlags: func(obj *crownCollideTestObject4EBB50) uint32 {
			s.events = append(s.events, "flags:"+obj.name)
			value := obj.flags
			if s.onFlags != nil {
				s.onFlags()
			}
			return value
		},
		loadClassLow: func(obj *crownCollideTestObject4EBB50) uint8 {
			s.events = append(s.events, "class:"+obj.name)
			value := obj.class
			if s.onClass != nil {
				s.onClass()
			}
			return value
		},
		pickup: func(who, crown *crownCollideTestObject4EBB50, flag1, flag2 int32) uint32 {
			s.events = append(s.events, fmt.Sprintf(
				"pickup:%s:%s:%d:%d",
				who.name,
				objectNameCrownCollide4EBB50(crown),
				flag1,
				flag2,
			))
			return s.pickupResult
		},
	}
}

func objectNameCrownCollide4EBB50(obj *crownCollideTestObject4EBB50) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func TestCrownCollide4EBB50GuardOrderAndOriginalTargetResult(t *testing.T) {
	tests := []struct {
		name       string
		target     *crownCollideTestObject4EBB50
		wantEvents []string
	}{
		{name: "nil target"},
		{
			name:       "disabled flag",
			target:     &crownCollideTestObject4EBB50{name: "disabled", flags: 0x20, class: 0x04},
			wantEvents: []string{"flags:disabled"},
		},
		{
			name:       "dead flag",
			target:     &crownCollideTestObject4EBB50{name: "dead", flags: 0x8000, class: 0x04},
			wantEvents: []string{"flags:dead"},
		},
		{
			name:       "both blocked flags",
			target:     &crownCollideTestObject4EBB50{name: "both", flags: 0x8020, class: 0x04},
			wantEvents: []string{"flags:both"},
		},
		{
			name:       "non player",
			target:     &crownCollideTestObject4EBB50{name: "unit", flags: 0x4000, class: 0x02},
			wantEvents: []string{"flags:unit", "class:unit"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := &crownCollideTestState4EBB50{pickupResult: 0xffffffff}
			got := crownCollide4EBB50(
				(*crownCollideTestObject4EBB50)(nil),
				tc.target,
				struct{ unread [0]func() }{},
				state.hooks(),
			)
			if got.target != tc.target || got.pickupAttempted || got.pickupResult != 0 {
				t.Fatalf("result = %+v, want untouched target %p", got, tc.target)
			}
			if !reflect.DeepEqual(state.events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", state.events, tc.wantEvents)
			}
		})
	}
}

func TestCrownCollide4EBB50PickupArgumentsAndFullResult(t *testing.T) {
	crown := &crownCollideTestObject4EBB50{name: "crown"}
	target := &crownCollideTestObject4EBB50{name: "player", flags: 0x4000, class: 0x84}
	state := &crownCollideTestState4EBB50{pickupResult: 0xf1234567}

	got := crownCollide4EBB50(crown, target, (*uint32)(nil), state.hooks())
	if got.target != target || !got.pickupAttempted || got.pickupResult != 0xf1234567 {
		t.Fatalf("result = %+v", got)
	}
	want := []string{"flags:player", "class:player", "pickup:player:crown:1:1"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestCrownCollide4EBB50UsesLiveClassAfterFlags(t *testing.T) {
	crown := &crownCollideTestObject4EBB50{name: "crown"}
	target := &crownCollideTestObject4EBB50{name: "target"}
	state := &crownCollideTestState4EBB50{pickupResult: 7}
	state.onFlags = func() {
		target.class = crownCollidePlayerClass4EBB50
		target.flags = crownCollideBlockedFlags4EBB50
	}

	got := crownCollide4EBB50(crown, target, "unread collision", state.hooks())
	if !got.pickupAttempted || got.pickupResult != 7 {
		t.Fatalf("result = %+v, want live-class pickup", got)
	}
	want := []string{"flags:target", "class:target", "pickup:target:crown:1:1"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestCrownCollide4EBB50PassesNilCrownAfterTargetGates(t *testing.T) {
	target := &crownCollideTestObject4EBB50{name: "player", class: crownCollidePlayerClass4EBB50}
	state := &crownCollideTestState4EBB50{pickupResult: 1}

	got := crownCollide4EBB50(
		(*crownCollideTestObject4EBB50)(nil),
		target,
		[1]uintptr{0xfeedface},
		state.hooks(),
	)
	if !got.pickupAttempted || got.pickupResult != 1 {
		t.Fatalf("result = %+v", got)
	}
	want := []string{"flags:player", "class:player", "pickup:player:nil:1:1"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestCrownCollide4EBB50ClassFaultOccursAfterFlags(t *testing.T) {
	target := &crownCollideTestObject4EBB50{name: "target"}
	state := &crownCollideTestState4EBB50{}
	state.onClass = func() { panic("class fault") }

	defer func() {
		if got := recover(); got != "class fault" {
			t.Fatalf("panic = %v, want class fault", got)
		}
		want := []string{"flags:target", "class:target"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %v, want %v", state.events, want)
		}
	}()
	crownCollide4EBB50(
		(*crownCollideTestObject4EBB50)(nil),
		target,
		(*struct{})(nil),
		state.hooks(),
	)
}
