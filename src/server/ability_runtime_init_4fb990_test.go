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
	node := &ExecAbilityClass{Unit: unit, Abil: AbilityHarpoon, Frame: 123, Active: 1}
	a := serverAbilities{s: s, execList: node}
	a.cooldowns[7][AbilityHarpoon] = 77

	a.Init4FB990()

	if a.s != s {
		t.Fatal("server owner changed during ability-runtime initialization")
	}
	if got := a.PlayerAbilityCooldownAt(7, AbilityHarpoon); got != 0 {
		t.Fatalf("new cooldown = %d, want 0", got)
	}
	if got := a.ExecHead(); got != nil {
		t.Fatalf("new execution-list head = %p, want nil", got)
	}
}

func TestAbilityRuntimeResetUsesInit4FB990Semantics(t *testing.T) {
	unit := new(Object)
	a := serverAbilities{execList: &ExecAbilityClass{Unit: unit}}
	a.cooldowns[3][AbilityWarcry] = -9

	a.Reset()

	if got := a.PlayerAbilityCooldownAt(3, AbilityWarcry); got != 0 {
		t.Fatalf("Reset cooldown = %d, want 0", got)
	}
	if got := a.ExecHead(); got != nil {
		t.Fatalf("Reset execution-list head = %p, want nil", got)
	}
}
