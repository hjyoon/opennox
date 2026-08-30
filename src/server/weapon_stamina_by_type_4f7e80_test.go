package server

import "testing"

var weaponStaminaPriority4F7E80 = []struct {
	mask uint32
	cost int32
}{
	{0x00000200, 70},
	{0x00004000, 100},
	{0x00000800, 50},
	{0x00000100, 45},
	{0x00001000, 75},
	{0x00002000, 100},
	{0x07ff8000, 45},
	{0x00000400, 75},
}

func weaponStaminaOracle4F7E80(flags uint32) int32 {
	for _, branch := range weaponStaminaPriority4F7E80 {
		if flags&branch.mask != 0 {
			return branch.cost
		}
	}
	return 10
}

func TestWeaponStaminaByType4F7E80BranchValuesAndPriority(t *testing.T) {
	if got := weaponStaminaByType4F7E80(0); got != 10 {
		t.Fatalf("zero flags cost = %d, want 10", got)
	}
	for i, branch := range weaponStaminaPriority4F7E80 {
		if got := weaponStaminaByType4F7E80(branch.mask & -branch.mask); got != branch.cost {
			t.Fatalf("branch %d single-bit cost = %d, want %d", i, got, branch.cost)
		}
		lower := uint32(0)
		for _, candidate := range weaponStaminaPriority4F7E80[i+1:] {
			lower |= candidate.mask
		}
		if got := weaponStaminaByType4F7E80(branch.mask | lower); got != branch.cost {
			t.Fatalf("branch %d priority cost = %d, want %d", i, got, branch.cost)
		}
	}
}

func TestWeaponStaminaByType4F7E80EveryRelevantBitCombination(t *testing.T) {
	relevantBits := [...]uint32{
		0x00000100, 0x00000200, 0x00000400, 0x00000800,
		0x00001000, 0x00002000, 0x00004000, 0x00008000,
		0x00010000, 0x00020000, 0x00040000, 0x00080000,
		0x00100000, 0x00200000, 0x00400000, 0x00800000,
		0x01000000, 0x02000000, 0x04000000,
	}
	const irrelevantBits = uint32(0xf80000ff)

	for subset := uint32(0); subset < 1<<len(relevantBits); subset++ {
		var flags uint32
		for bit, mask := range relevantBits {
			if subset&(1<<bit) != 0 {
				flags |= mask
			}
		}
		want := weaponStaminaOracle4F7E80(flags)
		if got := weaponStaminaByType4F7E80(flags); got != want {
			t.Fatalf("flags %#08x cost = %d, want %d", flags, got, want)
		}
		if got := weaponStaminaByType4F7E80(flags | irrelevantBits); got != want {
			t.Fatalf("flags %#08x with irrelevant bits cost = %d, want %d", flags, got, want)
		}
	}
}

func TestWeaponStaminaByType4F7E80EverySpecialGroupBit(t *testing.T) {
	for bit := uint32(0x00008000); bit <= 0x04000000; bit <<= 1 {
		if got := weaponStaminaByType4F7E80(bit); got != 45 {
			t.Fatalf("special-group bit %#08x cost = %d, want 45", bit, got)
		}
	}
}
