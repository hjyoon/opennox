package server

import (
	"testing"
	"unsafe"
)

func TestAbilityRuntimeInit4FB990OracleGeometry(t *testing.T) {
	if got, want := abilityRuntimePlayerSlots4FB990, 32; got != want {
		t.Fatalf("player slots = %d, want %d", got, want)
	}
	if got, want := abilityRuntimeAbilitySlots4FB990, 6; got != want {
		t.Fatalf("ability slots = %d, want %d", got, want)
	}
	if got, want := abilityRuntimeCooldownBytes4FB990, 0x300; got != want {
		t.Fatalf("cooldown bytes = %#x, want %#x", got, want)
	}
	if got, want := executingAbilityClassName4FB990, "executingAbilityClass"; got != want {
		t.Fatalf("allocator name = %q, want %q", got, want)
	}
	if got, want := executingAbilityClassRecordBytes4FB990, 24; got != want {
		t.Fatalf("PE32 record bytes = %d, want %d", got, want)
	}
	if got, want := executingAbilityClassPoolCapacity4FB990, 64; got != want {
		t.Fatalf("pool capacity = %d, want %d", got, want)
	}
}

func TestExecAbilityClass4FB990UsesNativePointerWidth(t *testing.T) {
	var record ExecAbilityClass
	ptrBytes := unsafe.Sizeof(uintptr(0))
	unitOffset := uintptr(4)
	if ptrBytes > unitOffset {
		unitOffset = ptrBytes
	}
	frameOffset := unitOffset + ptrBytes
	nextOffset := frameOffset + 8
	if rem := nextOffset % ptrBytes; rem != 0 {
		nextOffset += ptrBytes - rem
	}
	prevOffset := nextOffset + ptrBytes
	wantSize := prevOffset + ptrBytes

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Abil", unsafe.Offsetof(record.Abil), 0},
		{"Unit", unsafe.Offsetof(record.Unit), unitOffset},
		{"Frame", unsafe.Offsetof(record.Frame), frameOffset},
		{"Active", unsafe.Offsetof(record.Active), frameOffset + 4},
		{"Next", unsafe.Offsetof(record.Next), nextOffset},
		{"Prev", unsafe.Offsetof(record.Prev), prevOffset},
		{"Size", unsafe.Sizeof(record), wantSize},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d for %d-byte pointers", check.name, check.got, check.want, ptrBytes)
		}
	}
	if ptrBytes == 4 && unsafe.Sizeof(record) != executingAbilityClassRecordBytes4FB990 {
		t.Fatalf("PE32 record bytes = %d, oracle %d", unsafe.Sizeof(record), executingAbilityClassRecordBytes4FB990)
	}
}

func TestAbilityRuntimeInit4FB990ReplacesSessionState(t *testing.T) {
	s := new(Server)
	unit := new(Object)
	node := &ExecAbilityClass{Abil: AbilityHarpoon, Frame: 123, Active: 1}
	old := map[*Object]*unitAbilities{
		unit: {
			Cooldowns: [AbilityMax]int{AbilityHarpoon: 77},
			ExecList:  node,
		},
	}
	a := serverAbilities{s: s, ByUnit: old}

	a.Init4FB990()

	if a.s != s {
		t.Fatal("server owner changed during ability-runtime initialization")
	}
	if a.ByUnit == nil {
		t.Fatal("ability-runtime map is nil after initialization")
	}
	if got := len(a.ByUnit); got != 0 {
		t.Fatalf("active ability-runtime entries = %d, want 0", got)
	}
	if got := a.GetFor(unit); got.Cooldowns[AbilityHarpoon] != 0 || got.ExecList != nil {
		t.Fatalf("new unit state = %+v, want zero cooldowns and no execution list", got)
	}
	old[new(Object)] = new(unitAbilities)
	if got := len(a.ByUnit); got != 1 {
		t.Fatalf("new runtime aliases old map: active entries = %d, want 1", got)
	}
}

func TestAbilityRuntimeResetUsesInit4FB990Semantics(t *testing.T) {
	unit := new(Object)
	a := serverAbilities{ByUnit: map[*Object]*unitAbilities{unit: new(unitAbilities)}}

	a.Reset()

	if a.ByUnit == nil || len(a.ByUnit) != 0 {
		t.Fatalf("Reset runtime state = %#v, want a fresh empty map", a.ByUnit)
	}
}
