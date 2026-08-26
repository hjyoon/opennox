package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func trapDoorUpdateTestObject53DE80(data *TrapDoorCollideData) *Object {
	return &Object{
		ObjClass:    object.ClassImmobile,
		ObjFlags:    object.FlagActive | object.FlagEnabled,
		CollideData: unsafe.Pointer(data),
	}
}

func TestTrapDoorUpdate53DE80EnabledStates(t *testing.T) {
	t.Run("opening becomes open", func(t *testing.T) {
		unit := trapDoorUpdateTestObject53DE80(&TrapDoorCollideData{})
		unit.Field5 = 0x102
		new(Server).TrapDoorUpdate53DE80(unit, TrapDoorUpdateRuntime53DE80{})
		if unit.Field5 != 0x108 {
			t.Fatalf("xstatus = %#x, want 0x108", unit.Field5)
		}
	})
	t.Run("closing waits for deadline", func(t *testing.T) {
		data := &TrapDoorCollideData{NextFrame: 20, Activated: 7}
		unit := trapDoorUpdateTestObject53DE80(data)
		unit.Field5 = 0x104
		s := new(Server)
		s.SetFrame(19)
		s.TrapDoorUpdate53DE80(unit, TrapDoorUpdateRuntime53DE80{})
		if unit.Field5 != 0x104 || data.Activated != 7 {
			t.Fatalf("pre-deadline state = %#x/%d, want 0x104/7", unit.Field5, data.Activated)
		}
		s.SetFrame(20)
		s.TrapDoorUpdate53DE80(unit, TrapDoorUpdateRuntime53DE80{})
		if unit.Field5 != 0x108 {
			t.Fatalf("deadline xstatus = %#x, want 0x108", unit.Field5)
		}
	})
	t.Run("idle clears activation", func(t *testing.T) {
		data := &TrapDoorCollideData{Activated: 9}
		unit := trapDoorUpdateTestObject53DE80(data)
		new(Server).TrapDoorUpdate53DE80(unit, TrapDoorUpdateRuntime53DE80{})
		if data.Activated != 0 {
			t.Fatalf("activation = %d, want 0", data.Activated)
		}
	})
}

func TestTrapDoorUpdate53DE80DisabledDeadline(t *testing.T) {
	data := &TrapDoorCollideData{NextFrame: 100}
	unit := trapDoorUpdateTestObject53DE80(data)
	unit.ObjFlags &^= object.FlagEnabled
	unit.Field5 = 0x102
	s := new(Server)
	s.SetFrame(100)
	s.SetTickRate(30)
	var sound uint32
	s.TrapDoorUpdate53DE80(unit, TrapDoorUpdateRuntime53DE80{
		AudioEvent: func(id uint32, got *Object) {
			if got != unit {
				t.Fatal("audio object differs")
			}
			sound = id
		},
	})
	if !unit.ObjFlags.Has(object.FlagEnabled) || unit.Field5 != 0x104 {
		t.Fatalf("object state = enabled:%t xstatus:%#x, want true/0x104", unit.ObjFlags.Has(object.FlagEnabled), unit.Field5)
	}
	if data.NextFrame != 250 || sound != 874 {
		t.Fatalf("deadline/sound = %d/%d, want 250/874", data.NextFrame, sound)
	}
}
