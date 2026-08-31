package server

import (
	"testing"

	"github.com/opennox/libs/object"
)

func TestAbilityRuntimeGlobalList4FBEE0FiltersAndUnlinksByUnit(t *testing.T) {
	unitA := &Object{ObjClass: object.ClassPlayer}
	unitB := &Object{ObjClass: object.ClassPlayer}
	a1 := &ExecAbilityClass{Unit: unitA, Abil: AbilityHarpoon, Active: 1}
	b := &ExecAbilityClass{Unit: unitB, Abil: AbilityHarpoon, Active: 0, Prev: a1}
	a2 := &ExecAbilityClass{Unit: unitA, Abil: AbilityHarpoon, Active: 1, Prev: b}
	a1.Next = b
	b.Next = a2
	a := serverAbilities{execList: a1}

	if !a.IsActive(unitA, AbilityHarpoon) || !a.IsActive(unitB, AbilityHarpoon) {
		t.Fatal("global list did not find both units")
	}
	if !a.IsActiveVal(unitA, AbilityHarpoon) {
		t.Fatal("active value for unit A = false, want true")
	}
	if a.IsActiveVal(unitB, AbilityHarpoon) {
		t.Fatal("active value for unit B = true, want false")
	}
	if !a.IsAnyActive(unitA) || !a.IsAnyActive(unitB) || a.IsAnyActive(new(Object)) {
		t.Fatal("global any-active filtering mismatch")
	}

	a.DisableAbilityAaa(unitA, AbilityHarpoon)

	if a.ExecHead() != b || b.Prev != nil || b.Next != nil {
		t.Fatalf("remaining list = head %p, prev %p, next %p; want B/nil/nil", a.ExecHead(), b.Prev, b.Next)
	}
	if *a1 != (ExecAbilityClass{}) || *a2 != (ExecAbilityClass{}) {
		t.Fatalf("removed records were not released: A1=%+v A2=%+v", *a1, *a2)
	}
	if a.IsActive(unitA, AbilityHarpoon) || !a.IsActive(unitB, AbilityHarpoon) {
		t.Fatal("post-unlink unit filtering mismatch")
	}
}
