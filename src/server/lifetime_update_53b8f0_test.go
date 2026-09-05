package server

import (
	"math"
	"reflect"
	"testing"
)

type lifetimeUpdateTestData53B8F0 struct {
	duration uint32
}

type lifetimeUpdateTestObject53B8F0 struct {
	creation uint32
	update   *lifetimeUpdateTestData53B8F0
	flags    uint32
	death    int
}

type lifetimeUpdateTestState53B8F0 struct {
	frame   uint32
	events  []string
	called  int
	deleted *lifetimeUpdateTestObject53B8F0
	onData  func(*lifetimeUpdateTestObject53B8F0)
	onDur   func(*lifetimeUpdateTestData53B8F0)
	onDeath func(*lifetimeUpdateTestObject53B8F0)
	onStore func(*lifetimeUpdateTestObject53B8F0)
}

func (s *lifetimeUpdateTestState53B8F0) hooks() lifetimeUpdateHooks53B8F0[
	*lifetimeUpdateTestObject53B8F0,
	*lifetimeUpdateTestData53B8F0,
	int,
] {
	return lifetimeUpdateHooks53B8F0[
		*lifetimeUpdateTestObject53B8F0,
		*lifetimeUpdateTestData53B8F0,
		int,
	]{
		frame: func() uint32 {
			s.events = append(s.events, "frame")
			return s.frame
		},
		loadCreationFrame: func(obj *lifetimeUpdateTestObject53B8F0) uint32 {
			s.events = append(s.events, "creation")
			return obj.creation
		},
		loadUpdateData: func(obj *lifetimeUpdateTestObject53B8F0) *lifetimeUpdateTestData53B8F0 {
			s.events = append(s.events, "update")
			data := obj.update
			if s.onData != nil {
				s.onData(obj)
			}
			return data
		},
		loadDuration: func(data *lifetimeUpdateTestData53B8F0) uint32 {
			s.events = append(s.events, "duration")
			if s.onDur != nil {
				s.onDur(data)
			}
			return data.duration
		},
		loadFlags: func(obj *lifetimeUpdateTestObject53B8F0) uint32 {
			s.events = append(s.events, "flags")
			return obj.flags
		},
		loadDeath: func(obj *lifetimeUpdateTestObject53B8F0) int {
			s.events = append(s.events, "death")
			death := obj.death
			if s.onDeath != nil {
				s.onDeath(obj)
			}
			return death
		},
		storeFlags: func(obj *lifetimeUpdateTestObject53B8F0, flags uint32) {
			s.events = append(s.events, "store")
			obj.flags = flags
			if s.onStore != nil {
				s.onStore(obj)
			}
		},
		callDeath: func(death int, obj *lifetimeUpdateTestObject53B8F0) {
			s.events = append(s.events, "call")
			s.called = death
			if obj.flags&lifetimeUpdateDeadFlag53B8F0 == 0 {
				panic("death callback observed flags before Dead was stored")
			}
		},
		delayedDelete: func(obj *lifetimeUpdateTestObject53B8F0) {
			s.events = append(s.events, "delete")
			s.deleted = obj
			if obj.flags&lifetimeUpdateDeadFlag53B8F0 == 0 {
				panic("delayed delete observed flags before Dead was stored")
			}
		},
	}
}

func TestLifetimeUpdate53B8F0ExactBoundaryReadsOnlyAgeInputs(t *testing.T) {
	obj := &lifetimeUpdateTestObject53B8F0{
		creation: 100,
		update:   &lifetimeUpdateTestData53B8F0{duration: 30},
		flags:    0x41,
		death:    17,
	}
	state := &lifetimeUpdateTestState53B8F0{frame: 130}
	lifetimeUpdate53B8F0(obj, state.hooks())
	want := []string{"frame", "creation", "update", "duration"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %#v, want %#v", state.events, want)
	}
	if obj.flags != 0x41 || state.called != 0 || state.deleted != nil {
		t.Fatalf("flags/called/deleted = (%#x, %d, %p), want (0x41, 0, nil)", obj.flags, state.called, state.deleted)
	}
}

func TestLifetimeUpdate53B8F0DeathPathOrderAndArguments(t *testing.T) {
	obj := &lifetimeUpdateTestObject53B8F0{
		creation: 100,
		update:   &lifetimeUpdateTestData53B8F0{duration: 30},
		flags:    0xa5001234,
		death:    23,
	}
	state := &lifetimeUpdateTestState53B8F0{frame: 131}
	lifetimeUpdate53B8F0(obj, state.hooks())
	want := []string{"frame", "creation", "update", "duration", "flags", "death", "store", "call"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %#v, want %#v", state.events, want)
	}
	if obj.flags != 0xa5009234 || state.called != 23 || state.deleted != nil {
		t.Fatalf("flags/called/deleted = (%#x, %d, %p), want (0xa5009234, 23, nil)", obj.flags, state.called, state.deleted)
	}
}

func TestLifetimeUpdate53B8F0NilDeathFallsBackAfterFlagStore(t *testing.T) {
	obj := &lifetimeUpdateTestObject53B8F0{
		creation: 8,
		update:   &lifetimeUpdateTestData53B8F0{duration: 0},
		flags:    0x40000001,
	}
	state := &lifetimeUpdateTestState53B8F0{frame: 9}
	lifetimeUpdate53B8F0(obj, state.hooks())
	want := []string{"frame", "creation", "update", "duration", "flags", "death", "store", "delete"}
	if !reflect.DeepEqual(state.events, want) || obj.flags != 0x40008001 || state.deleted != obj {
		t.Fatalf("events/flags/deleted = (%#v, %#x, %p), want (%#v, 0x40008001, %p)", state.events, obj.flags, state.deleted, want, obj)
	}
}

func TestLifetimeUpdate53B8F0StrictUnsignedExpiry(t *testing.T) {
	for _, tc := range []struct {
		name     string
		frame    uint32
		creation uint32
		duration uint32
		expired  bool
	}{
		{name: "zero age zero duration", frame: 7, creation: 7, duration: 0},
		{name: "zero duration one tick old", frame: 8, creation: 7, duration: 0, expired: true},
		{name: "one tick early", frame: 129, creation: 100, duration: 30},
		{name: "exact duration", frame: 130, creation: 100, duration: 30},
		{name: "one tick expired", frame: 131, creation: 100, duration: 30, expired: true},
		{name: "frame wrap exact duration", frame: 4, creation: math.MaxUint32 - 20, duration: 25},
		{name: "frame wrap expired", frame: 5, creation: math.MaxUint32 - 20, duration: 25, expired: true},
		{name: "maximum duration", frame: math.MaxUint32, creation: 0, duration: math.MaxUint32},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := &lifetimeUpdateTestObject53B8F0{
				creation: tc.creation,
				update:   &lifetimeUpdateTestData53B8F0{duration: tc.duration},
			}
			state := &lifetimeUpdateTestState53B8F0{frame: tc.frame}
			lifetimeUpdate53B8F0(obj, state.hooks())
			if (state.deleted != nil) != tc.expired {
				t.Fatalf("deleted = %p, want expired=%t", state.deleted, tc.expired)
			}
		})
	}
}

func TestLifetimeUpdate53B8F0CachesDataFlagsAndDeath(t *testing.T) {
	first := &lifetimeUpdateTestData53B8F0{duration: 1}
	obj := &lifetimeUpdateTestObject53B8F0{update: first, flags: 0x25, death: 31}
	state := &lifetimeUpdateTestState53B8F0{frame: 2}
	state.onData = func(obj *lifetimeUpdateTestObject53B8F0) {
		obj.update = &lifetimeUpdateTestData53B8F0{duration: math.MaxUint32}
	}
	state.onDeath = func(obj *lifetimeUpdateTestObject53B8F0) {
		obj.flags = 0xdead0000
	}
	state.onStore = func(obj *lifetimeUpdateTestObject53B8F0) {
		obj.death = 0
	}
	lifetimeUpdate53B8F0(obj, state.hooks())
	if obj.flags != 0x8025 || state.called != 31 {
		t.Fatalf("cached flags/death = (%#x, %d), want (0x8025, 31)", obj.flags, state.called)
	}
}

func TestLifetimeUpdate53B8F0LoadsPostDurationFieldsLive(t *testing.T) {
	obj := &lifetimeUpdateTestObject53B8F0{
		update: &lifetimeUpdateTestData53B8F0{duration: 0},
		flags:  1,
		death:  3,
	}
	state := &lifetimeUpdateTestState53B8F0{frame: 1}
	state.onDur = func(*lifetimeUpdateTestData53B8F0) {
		obj.flags = 0x40
		obj.death = 7
	}
	lifetimeUpdate53B8F0(obj, state.hooks())
	if obj.flags != 0x8040 || state.called != 7 {
		t.Fatalf("live post-duration flags/death = (%#x, %d), want (0x8040, 7)", obj.flags, state.called)
	}
}

func TestLifetimeUpdate53B8F0NilSourceFaultOrder(t *testing.T) {
	state := &lifetimeUpdateTestState53B8F0{frame: 1}
	defer func() {
		if recover() == nil {
			t.Fatal("nil source returned without a panic")
		}
		want := []string{"frame", "creation"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %#v, want %#v", state.events, want)
		}
	}()
	lifetimeUpdate53B8F0[*lifetimeUpdateTestObject53B8F0, *lifetimeUpdateTestData53B8F0, int](nil, state.hooks())
}
