package server

import "testing"

func TestWeaponStaminaByType4F7E80NativeBinding(t *testing.T) {
	for _, flags := range []uint32{
		0,
		0x00000400,
		0x04000000,
		0xf80000ff,
		0xffffffff,
	} {
		want := weaponStaminaOracle4F7E80(flags)
		if got := WeaponStaminaByType4F7E80(flags); got != want {
			t.Fatalf("flags %#08x cost = %d, want %d", flags, got, want)
		}
	}
}
