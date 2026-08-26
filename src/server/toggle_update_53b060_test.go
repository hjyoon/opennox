package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func toggleUpdateTestObject53B060(update *ToggleUpdateData) *Object {
	return &Object{
		ObjClass:   object.ClassImmobile,
		ObjFlags:   object.FlagActive | object.FlagEnabled,
		UpdateData: unsafe.Pointer(update),
	}
}

func toggleUpdateTestRuntime53B060(events *[]string, caller *Object) ToggleUpdateRuntime53B060 {
	return ToggleUpdateRuntime53B060{
		CollideTarget: func(*Object) *Object { return caller },
		AudioEvent: func(id uint32, _ *Object) {
			*events = append(*events, "audio")
		},
		ScriptCallback: func(*ScriptCallback, *Object, *Object, ScriptEventType) {
			*events = append(*events, "script")
		},
	}
}

func TestToggleUpdate53B060ActivationCycle(t *testing.T) {
	update := &ToggleUpdateData{Flags: 0x1, SoundActivate: 17}
	unit := toggleUpdateTestObject53B060(update)
	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(100)
	events := []string{}
	runtime := toggleUpdateTestRuntime53B060(&events, new(Object))

	if got := s.ToggleUpdate53B060(unit, runtime); got != 0x5 {
		t.Fatalf("initial return = %#x, want 0x5", got)
	}
	if update.State != 0 || unit.Frame134 != 100 || len(events) != 0 {
		t.Fatalf("initial state/frame/events = %d/%d/%v", update.State, unit.Frame134, events)
	}

	update.Flags |= 0x1
	s.SetFrame(101)
	if got := s.ToggleUpdate53B060(unit, runtime); got != 0xd {
		t.Fatalf("activation return = %#x, want 0xd", got)
	}
	if update.Flags != 0xc || update.State != 3 || unit.Frame134 != 131 {
		t.Fatalf("activation flags/state/frame = %#x/%d/%d, want 0xc/3/131", update.Flags, update.State, unit.Frame134)
	}
	if len(events) != 2 || events[0] != "audio" || events[1] != "script" {
		t.Fatalf("activation events = %v", events)
	}

	s.SetFrame(132)
	s.ToggleUpdate53B060(unit, runtime)
	if update.State != 1 {
		t.Fatalf("released state = %d, want 1", update.State)
	}

	update.Flags |= 0x1
	s.SetFrame(133)
	s.ToggleUpdate53B060(unit, runtime)
	if update.State != 0 || unit.Frame134 != 163 || len(events) != 4 {
		t.Fatalf("deactivation state/frame/events = %d/%d/%v", update.State, unit.Frame134, events)
	}
}

func TestToggleUpdate53B060OneShotAndDisabled(t *testing.T) {
	t.Run("one shot completes", func(t *testing.T) {
		update := &ToggleUpdateData{Flags: 0xb, State: 1}
		unit := toggleUpdateTestObject53B060(update)
		s := new(Server)
		s.SetFrame(2)
		events := []string{}
		s.ToggleUpdate53B060(unit, toggleUpdateTestRuntime53B060(&events, nil))
		if update.State != 5 || update.Flags != 0xe {
			t.Fatalf("state/flags = %d/%#x, want 5/0xe", update.State, update.Flags)
		}
	})
	t.Run("disabled clears request and initialized bits", func(t *testing.T) {
		update := &ToggleUpdateData{Flags: 0xffffffff, State: 7}
		unit := toggleUpdateTestObject53B060(update)
		unit.ObjFlags &^= object.FlagEnabled
		got := new(Server).ToggleUpdate53B060(unit, ToggleUpdateRuntime53B060{})
		if update.Flags != 0xfffffff6 || got != 0xf6 || update.State != 7 {
			t.Fatalf("flags/return/state = %#x/%#x/%d", update.Flags, got, update.State)
		}
	})
}
