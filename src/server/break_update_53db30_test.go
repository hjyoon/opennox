package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestBreakUpdate53DB30TransitionsAndNativePointer(t *testing.T) {
	s := new(Server)
	s.SetFrame(math.MaxUint32 - 5)
	s.SetTickRate(4)

	source := &Object{
		ObjClass: object.ClassImmobile,
		ObjFlags: 0x8000,
		Field5:   0x102,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
	}
	s.BreakUpdate53DB30(source)
	if source.Field5 != 0x104 || source.ObjFlags != 0x8040 || source.Field34 != 2 {
		t.Fatalf("opening state = (%#x, %#x, %d), want (0x104, 0x8040, 2)", source.Field5, source.ObjFlags, source.Field34)
	}

	// The middle phase uses a strict greater-than comparison.
	s.SetFrame(2)
	s.BreakUpdate53DB30(source)
	if source.Field5 != 0x104 {
		t.Fatalf("at deadline status = %#x, want 0x104", source.Field5)
	}
	s.SetFrame(3)
	s.BreakUpdate53DB30(source)
	if source.Field5 != 0x108 {
		t.Fatalf("after deadline status = %#x, want 0x108", source.Field5)
	}

	s.Objs.AddToUpdatable(source)
	s.BreakUpdate53DB30(source)
	if source.IsUpdatable != 0 || s.Objs.UpdatableList != nil {
		t.Fatalf("terminal source remained updatable: %d/%p", source.IsUpdatable, s.Objs.UpdatableList)
	}
}

func TestBreakUpdate53DB30BitPriorityAndFlagGate(t *testing.T) {
	s := new(Server)
	s.SetFrame(100)
	s.SetTickRate(30)

	withoutFlag := &Object{ObjClass: object.ClassImmobile, Field5: 0x2}
	s.BreakUpdate53DB30(withoutFlag)
	if withoutFlag.Field5 != 0x2 || withoutFlag.Field34 != 0 {
		t.Fatalf("opening without flag = (%#x, %d), want (0x2, 0)", withoutFlag.Field5, withoutFlag.Field34)
	}

	// Bit 2 wins over bit 4, matching the original if/else-if chain.
	priority := &Object{ObjClass: object.ClassImmobile, ObjFlags: 0x8000, Field5: 0x6}
	s.BreakUpdate53DB30(priority)
	if priority.Field5 != 0x4 || priority.Field34 != 160 {
		t.Fatalf("priority state = (%#x, %d), want (0x4, 160)", priority.Field5, priority.Field34)
	}
	s.BreakUpdate53DB30(nil)
}

func TestBreakAndRemoveUpdate53DC30ExactStates(t *testing.T) {
	s := new(Server)
	s.SetFrame(100)
	s.SetTickRate(30)

	// This updater switches on the whole word: extra bits suppress every case.
	extraBits := &Object{ObjClass: object.ClassImmobile, ObjFlags: 0x8000, Field5: 0x102}
	s.BreakAndRemoveUpdate53DC30(extraBits, BreakAndRemoveUpdateRuntime53DC30{})
	if extraBits.Field5 != 0x102 || extraBits.Field34 != 0 || extraBits.ObjFlags != 0x8000 {
		t.Fatalf("extra-bit state changed: (%#x, %d, %#x)", extraBits.Field5, extraBits.Field34, extraBits.ObjFlags)
	}

	source := &Object{ObjClass: object.ClassImmobile, ObjFlags: 0x8000, Field5: 0x2}
	s.BreakAndRemoveUpdate53DC30(source, BreakAndRemoveUpdateRuntime53DC30{})
	if source.Field5 != 0x4 || source.Field34 != 160 || source.ObjFlags != 0x8040 {
		t.Fatalf("opening state = (%#x, %d, %#x), want (0x4, 160, 0x8040)", source.Field5, source.Field34, source.ObjFlags)
	}
	s.SetFrame(160)
	s.BreakAndRemoveUpdate53DC30(source, BreakAndRemoveUpdateRuntime53DC30{})
	if source.Field5 != 0x4 {
		t.Fatalf("at deadline status = %#x, want 0x4", source.Field5)
	}
	s.SetFrame(161)
	s.BreakAndRemoveUpdate53DC30(source, BreakAndRemoveUpdateRuntime53DC30{})
	if source.Field5 != 0x8 {
		t.Fatalf("after deadline status = %#x, want 0x8", source.Field5)
	}

	s.Objs.AddToUpdatable(source)
	var deleted *Object
	s.BreakAndRemoveUpdate53DC30(source, BreakAndRemoveUpdateRuntime53DC30{
		DelayedDelete: func(got *Object) { deleted = got },
	})
	if source.IsUpdatable != 0 || s.Objs.UpdatableList != nil || deleted != source {
		t.Fatalf("terminal state = updatable %d/%p, deleted %p; want removed and %p", source.IsUpdatable, s.Objs.UpdatableList, deleted, source)
	}
}

func TestBreakAndRemoveUpdate53DC30MissingDeleteStillRemoves(t *testing.T) {
	s := new(Server)
	source := &Object{Field5: 0x8}
	s.Objs.AddToUpdatable(source)
	s.BreakAndRemoveUpdate53DC30(source, BreakAndRemoveUpdateRuntime53DC30{})
	if source.IsUpdatable != 0 || s.Objs.UpdatableList != nil {
		t.Fatalf("source remained updatable: %d/%p", source.IsUpdatable, s.Objs.UpdatableList)
	}
	s.BreakAndRemoveUpdate53DC30(nil, BreakAndRemoveUpdateRuntime53DC30{
		DelayedDelete: func(*Object) { t.Fatal("nil source reached deletion") },
	})
}
