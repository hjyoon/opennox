package server

import (
	"testing"

	"github.com/opennox/libs/object"
)

func TestOpenUpdate53DBB0Transitions(t *testing.T) {
	s := new(Server)
	s.SetFrame(100)
	s.SetTickRate(30)

	opening := &Object{ObjClass: object.ClassImmobile, Field5: 0x102}
	s.OpenUpdate53DBB0(opening)
	if opening.Field5 != 0x102 || opening.Field34 != 0 {
		t.Fatalf("opening without flag = (%#x, %d), want (0x102, 0)", opening.Field5, opening.Field34)
	}
	opening.ObjFlags = 0x8000
	s.OpenUpdate53DBB0(opening)
	if opening.Field5 != 0x104 || opening.Field34 != 160 {
		t.Fatalf("opening = (%#x, %d), want (0x104, 160)", opening.Field5, opening.Field34)
	}

	// The closing phase uses a strict greater-than deadline.
	s.SetFrame(160)
	s.OpenUpdate53DBB0(opening)
	if opening.Field5 != 0x104 {
		t.Fatalf("at deadline status = %#x, want 0x104", opening.Field5)
	}
	s.SetFrame(161)
	s.OpenUpdate53DBB0(opening)
	if opening.Field5 != 0x108 {
		t.Fatalf("after deadline status = %#x, want 0x108", opening.Field5)
	}
}

func TestOpenUpdate53DBB0PriorityAndRemoval(t *testing.T) {
	s := new(Server)
	s.SetFrame(12)
	s.SetTickRate(30)

	// An object with both bits set must still take the opening branch first.
	obj := &Object{ObjClass: object.ClassImmobile, ObjFlags: 0x8000, Field5: 0x106}
	s.Objs.AddToUpdatable(obj)
	s.OpenUpdate53DBB0(obj)
	if obj.Field5 != 0x104 || obj.IsUpdatable != 1 {
		t.Fatalf("priority state = (%#x, %d), want (0x104, 1)", obj.Field5, obj.IsUpdatable)
	}

	obj.Field5 = 0x108
	s.OpenUpdate53DBB0(obj)
	if obj.IsUpdatable != 0 || s.Objs.UpdatableList != nil {
		t.Fatalf("finished object remained updatable: %d/%p", obj.IsUpdatable, s.Objs.UpdatableList)
	}
	s.OpenUpdate53DBB0(nil)
}
