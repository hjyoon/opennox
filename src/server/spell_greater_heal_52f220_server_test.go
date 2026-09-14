package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestGreaterHeal52F2E0NativeBindingReadsTarget48(t *testing.T) {
	s := &Server{}
	r := &DurSpell{}
	r.Pos.X = math.Float32frombits(0x3fdccccc) // PE32's false target pointer.
	if got := s.SpellGreaterHealUpdate52F2E0(r, SpellGreaterHealRuntime52F2E0{}); got != 1 {
		t.Fatalf("nil target update = %d", got)
	}
	target := &Object{ObjFlags: object.Flags(0x20)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= 0xffffffff {
		t.Fatal("expected target above 4 GiB")
	}
	r.Target48 = target
	if got := s.SpellGreaterHealUpdate52F2E0(r, SpellGreaterHealRuntime52F2E0{}); got != 1 {
		t.Fatalf("destroyed target update = %d", got)
	}
}

func TestGreaterHeal52F220NativeCreateNoCasterNoMode(t *testing.T) {
	s := &Server{}
	r := &DurSpell{}
	r.Pos.X = math.Float32frombits(0x3fdccccc)
	if got := s.SpellGreaterHealCreate52F220(r, SpellGreaterHealRuntime52F220{}); got != 1 || r.Target48 != nil {
		t.Fatalf("nil caster create = %d, target = %p", got, r.Target48)
	}
}
