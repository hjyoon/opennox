package server

import (
	"math"
	"testing"
)

func TestPlayerAbilityCooldownIndexedSetNative4FC4A0AllSlotsAndReturn(t *testing.T) {
	a := new(serverAbilities)
	a.Init4FB990()
	for playerIndex := int32(0); playerIndex < abilityRuntimePlayerSlots4FB990; playerIndex++ {
		for ability := AbilityInvalid; ability < AbilityMax; ability++ {
			want := int32(0x40000000) + playerIndex*int32(AbilityMax) + int32(ability)
			if got := a.PlayerAbilityCooldownIndexedSet4FC4A0(playerIndex, ability, want); got != want {
				t.Fatalf("player %d ability %d return = %d, want %d", playerIndex, ability, got, want)
			}
			if got := a.PlayerAbilityCooldownAt(uint8(playerIndex), ability); got != want {
				t.Fatalf("player %d ability %d stored = %d, want %d", playerIndex, ability, got, want)
			}
		}
	}
}

func TestPlayerAbilityCooldownIndexedSetNative4FC4A0SharesAccessorStorage(t *testing.T) {
	a := new(serverAbilities)
	a.Init4FB990()

	want := int32(math.MinInt32 + 0x4a0)
	if got := a.PlayerAbilityCooldownIndexedSet4FC4A0(17, AbilityTreadLightly, want); got != want {
		t.Fatalf("return = %#08x, want %#08x", uint32(got), uint32(want))
	}
	if got := a.PlayerAbilityCooldownAt(17, AbilityTreadLightly); got != want {
		t.Fatalf("accessor value = %#08x, want %#08x", uint32(got), uint32(want))
	}

	a.SetPlayerAbilityCooldownAt(31, AbilityInfravis, math.MinInt32)
	if got := a.PlayerAbilityCooldownAt(31, AbilityInfravis); got != math.MinInt32 {
		t.Fatalf("routed setter value = %#08x, want 0x80000000", uint32(got))
	}
}

func TestPlayerAbilityCooldownIndexedSetNative4FC4A0UsesWrappedFlatIndex(t *testing.T) {
	a := new(serverAbilities)
	a.Init4FB990()

	// Six times INT32_MIN wraps to zero in the PE32 calculation.
	if got := a.PlayerAbilityCooldownIndexedSet4FC4A0(math.MinInt32, AbilityInfravis, 57); got != 57 {
		t.Fatalf("return = %d, want 57", got)
	}
	if got := a.PlayerAbilityCooldownAt(0, AbilityInfravis); got != 57 {
		t.Fatalf("wrapped flat slot 5 = %d, want 57", got)
	}

	// The original validates neither component: -1*6 + 6 addresses flat zero.
	if got := a.PlayerAbilityCooldownIndexedSet4FC4A0(-1, AbilityMax, 91); got != 91 {
		t.Fatalf("alias return = %d, want 91", got)
	}
	if got := a.PlayerAbilityCooldownAt(0, AbilityInvalid); got != 91 {
		t.Fatalf("aliased flat slot 0 = %d, want 91", got)
	}
}

func TestPlayerAbilityCooldownIndexedSetNative4FC4A0OutOfMatrixTrapsBeforeStore(t *testing.T) {
	a := new(serverAbilities)
	a.Init4FB990()
	a.cooldowns[0][AbilityInvalid] = 73

	defer func() {
		if recover() == nil {
			t.Fatal("out-of-matrix flat index did not trap")
		}
		if got := a.PlayerAbilityCooldownAt(0, AbilityInvalid); got != 73 {
			t.Fatalf("sentinel = %d, want 73", got)
		}
	}()
	a.PlayerAbilityCooldownIndexedSet4FC4A0(-1, AbilityInvalid, 99)
}
