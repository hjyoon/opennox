package server

import (
	"math"
	"testing"
	"unsafe"
)

func TestActiveAbilityDurationNative4FC030GlobalListAndWidths(t *testing.T) {
	s := new(Server)
	s.SetFrame(1)
	unit := new(Object)
	other := new(Object)
	match := &ExecAbilityClass{
		Unit: unit, Abil: Ability(-7), Frame: 0x80000001, Active: 0,
	}
	head := &ExecAbilityClass{
		Unit: other, Abil: Ability(-7), Frame: 77, Active: 1, Next: match,
	}
	s.Abils = serverAbilities{s: s, execList: head}

	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}
	if got, want := s.Abils.Sub4FC030(unit, Ability(-7)), int32(math.MinInt32); got != want {
		t.Fatalf("duration = %d (%08x), want %d (%08x)", got, uint32(got), want, uint32(want))
	}
	if s.Abils.ExecHead() != head || head.Next != match || match.Active != 0 {
		t.Fatal("duration lookup mutated the global execution list")
	}
	if got := s.Abils.Sub4FC030(unit, AbilityWarcry); got != -1 {
		t.Fatalf("ability miss = %d, want -1", got)
	}
}

func TestActiveAbilityDurationNative4FC030MissDoesNotReadFrameOwner(t *testing.T) {
	unit := new(Object)
	a := serverAbilities{execList: &ExecAbilityClass{Unit: new(Object), Abil: AbilityBerserk}}
	if got := a.Sub4FC030(unit, AbilityBerserk); got != -1 {
		t.Fatalf("miss = %d, want -1", got)
	}
}
