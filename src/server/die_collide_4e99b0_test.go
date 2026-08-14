package server

import (
	"reflect"
	"testing"
)

type dieCollideTestObject4E99B0 struct {
	name  string
	class uint8
	flags uint32
	death int
}

type dieCollideTestState4E99B0 struct {
	events   []string
	sameTeam int32
	onSame   func()
	onDeath  func(*dieCollideTestObject4E99B0) int
	onStore  func(*dieCollideTestObject4E99B0)
	called   int
	deleted  *dieCollideTestObject4E99B0
}

func (s *dieCollideTestState4E99B0) hooks() dieCollideHooks4E99B0[*dieCollideTestObject4E99B0, int] {
	return dieCollideHooks4E99B0[*dieCollideTestObject4E99B0, int]{
		unitsOnSameTeam: func(source, target *dieCollideTestObject4E99B0) int32 {
			s.events = append(s.events, "same")
			if s.onSame != nil {
				s.onSame()
			}
			return s.sameTeam
		},
		classLow: func(obj *dieCollideTestObject4E99B0) uint8 {
			s.events = append(s.events, "class:"+obj.name)
			return obj.class
		},
		loadFlags: func(obj *dieCollideTestObject4E99B0) uint32 {
			if obj == nil {
				s.events = append(s.events, "flags:nil")
				panic("nil flags")
			}
			s.events = append(s.events, "flags:"+obj.name)
			return obj.flags
		},
		loadDeath: func(obj *dieCollideTestObject4E99B0) int {
			s.events = append(s.events, "death:"+obj.name)
			if s.onDeath != nil {
				return s.onDeath(obj)
			}
			return obj.death
		},
		storeFlags: func(obj *dieCollideTestObject4E99B0, flags uint32) {
			s.events = append(s.events, "store:"+obj.name)
			obj.flags = flags
			if s.onStore != nil {
				s.onStore(obj)
			}
		},
		callDeath: func(death int, obj *dieCollideTestObject4E99B0) {
			s.events = append(s.events, "call:"+obj.name)
			s.called = death
		},
		delayedDelete: func(obj *dieCollideTestObject4E99B0) {
			s.events = append(s.events, "delete:"+obj.name)
			s.deleted = obj
		},
	}
}

func TestDieCollide4E99B0NilTargetReadsNothing(t *testing.T) {
	state := &dieCollideTestState4E99B0{}
	dieCollide4E99B0[*dieCollideTestObject4E99B0, int](nil, nil, struct{ unread int }{9}, state.hooks())
	if len(state.events) != 0 {
		t.Fatalf("events = %#v, want none", state.events)
	}
}

func TestDieCollide4E99B0TargetGatesAndAnyNonzeroSameTeam(t *testing.T) {
	for _, tc := range []struct {
		name     string
		class    uint8
		sameTeam int32
		want     []string
	}{
		{name: "same-positive", class: 6, sameTeam: 1, want: []string{"same"}},
		{name: "same-negative", class: 6, sameTeam: -1, want: []string{"same"}},
		{name: "non-unit", class: 0xf1, want: []string{"same", "class:target"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := &dieCollideTestObject4E99B0{name: "source"}
			target := &dieCollideTestObject4E99B0{name: "target", class: tc.class}
			state := &dieCollideTestState4E99B0{sameTeam: tc.sameTeam}
			dieCollide4E99B0(source, target, nil, state.hooks())
			if !reflect.DeepEqual(state.events, tc.want) {
				t.Fatalf("events = %#v, want %#v", state.events, tc.want)
			}
		})
	}
}

func TestDieCollide4E99B0DeathCallbackOrderAndArguments(t *testing.T) {
	source := &dieCollideTestObject4E99B0{name: "source", flags: 0xa5001234, death: 17}
	target := &dieCollideTestObject4E99B0{name: "target", class: 0x82}
	state := &dieCollideTestState4E99B0{}
	dieCollide4E99B0(source, target, nil, state.hooks())
	want := []string{"same", "class:target", "flags:source", "death:source", "store:source", "call:source"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %#v, want %#v", state.events, want)
	}
	if source.flags != 0xa5009234 || state.called != 17 || state.deleted != nil {
		t.Fatalf("flags/called/deleted = (%#x, %d, %p), want (%#x, 17, nil)", source.flags, state.called, state.deleted, uint32(0xa5009234))
	}
}

func TestDieCollide4E99B0NilDeathStoresBeforeDelayedDelete(t *testing.T) {
	source := &dieCollideTestObject4E99B0{name: "source", flags: 0x40000001}
	target := &dieCollideTestObject4E99B0{name: "target", class: 4}
	state := &dieCollideTestState4E99B0{}
	dieCollide4E99B0(source, target, nil, state.hooks())
	want := []string{"same", "class:target", "flags:source", "death:source", "store:source", "delete:source"}
	if !reflect.DeepEqual(state.events, want) || source.flags != 0x40008001 || state.deleted != source {
		t.Fatalf("events/flags/deleted = (%#v, %#x, %p), want (%#v, %#x, %p)", state.events, source.flags, state.deleted, want, uint32(0x40008001), source)
	}
}

func TestDieCollide4E99B0ReadsLiveFieldsAfterSameTeam(t *testing.T) {
	source := &dieCollideTestObject4E99B0{name: "source"}
	target := &dieCollideTestObject4E99B0{name: "target"}
	state := &dieCollideTestState4E99B0{}
	state.onSame = func() {
		target.class = 2
		source.flags = 0x40
		source.death = 23
	}
	dieCollide4E99B0(source, target, nil, state.hooks())
	if source.flags != 0x8040 || state.called != 23 {
		t.Fatalf("flags/called = (%#x, %d), want (0x8040, 23)", source.flags, state.called)
	}
}

func TestDieCollide4E99B0CachesFlagsAndDeathBeforeStore(t *testing.T) {
	source := &dieCollideTestObject4E99B0{name: "source", flags: 0x25, death: 31}
	target := &dieCollideTestObject4E99B0{name: "target", class: 2}
	state := &dieCollideTestState4E99B0{}
	state.onDeath = func(obj *dieCollideTestObject4E99B0) int {
		obj.flags = 0xdead0000
		return obj.death
	}
	state.onStore = func(obj *dieCollideTestObject4E99B0) { obj.death = 0 }
	dieCollide4E99B0(source, target, nil, state.hooks())
	if source.flags != 0x8025 || state.called != 31 {
		t.Fatalf("cached flags/death = (%#x, %d), want (0x8025, 31)", source.flags, state.called)
	}
}

func TestDieCollide4E99B0NilSourceFaultOrder(t *testing.T) {
	target := &dieCollideTestObject4E99B0{name: "target", class: 2}
	state := &dieCollideTestState4E99B0{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil source returned without a panic")
		}
		want := []string{"same", "class:target", "flags:nil"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %#v, want %#v", state.events, want)
		}
	}()
	dieCollide4E99B0[*dieCollideTestObject4E99B0, int](nil, target, nil, state.hooks())
}
