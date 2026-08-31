package server

import (
	"math"
	"testing"
	"unsafe"
)

func TestActiveAbilityDeadlineNative4FC070GlobalListAndWidths(t *testing.T) {
	s := new(Server)
	s.SetFrame(1)
	unit := new(Object)
	other := new(Object)
	match := &ExecAbilityClass{
		Unit: unit, Abil: Ability(-7), Frame: 99, Active: 0,
	}
	head := &ExecAbilityClass{
		Unit: other, Abil: Ability(-7), Frame: 77, Active: 1, Next: match,
	}
	s.Abils = serverAbilities{s: s, execList: head}

	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}
	s.Abils.Sub4FC070(unit, Ability(-7), math.MinInt32)
	if got, want := match.Frame, uint32(0x80000001); got != want {
		t.Fatalf("deadline = %08x, want %08x", got, want)
	}
	if s.Abils.ExecHead() != head || head.Next != match || match.Active != 0 || head.Frame != 77 {
		t.Fatal("deadline adjustment mutated list topology, Active, or a skipped record")
	}
}

func TestActiveAbilityDeadlineNative4FC070MissDoesNotReadFrameOwner(t *testing.T) {
	unit := new(Object)
	a := serverAbilities{execList: &ExecAbilityClass{Unit: new(Object), Abil: AbilityBerserk, Frame: 44}}
	a.Sub4FC070(unit, AbilityBerserk, 123)
	if got := a.execList.Frame; got != 44 {
		t.Fatalf("miss deadline = %d, want unchanged 44", got)
	}

	empty := serverAbilities{}
	empty.Sub4FC070(unit, AbilityBerserk, 123)
}
