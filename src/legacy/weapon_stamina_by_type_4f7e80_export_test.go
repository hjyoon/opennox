package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestWeaponStaminaByTypeExport4F7E80PreservesDwordAndPriority(t *testing.T) {
	representatives := [...]uint32{
		0x00000200,
		0x00004000,
		0x00000800,
		0x00000100,
		0x00001000,
		0x00002000,
		0x00008000,
		0x00000400,
	}
	for subset := uint32(0); subset < 1<<len(representatives); subset++ {
		var flags uint32
		for bit, mask := range representatives {
			if subset&(1<<bit) != 0 {
				flags |= mask
			}
		}
		if subset&1 != 0 {
			flags |= 0xf80000ff
		}
		want := server.WeaponStaminaByType4F7E80(flags)
		if got := weaponStaminaByTypeExportCall4F7E80(flags); got != want {
			t.Fatalf("flags %#08x export cost = %d, want %d", flags, got, want)
		}
	}
}

func TestWeaponStaminaByTypeExport4F7E80EverySpecialBitAndHighBits(t *testing.T) {
	for bit := uint32(0x00008000); bit <= 0x04000000; bit <<= 1 {
		if got := weaponStaminaByTypeExportCall4F7E80(bit); got != 45 {
			t.Fatalf("special bit %#08x export cost = %d, want 45", bit, got)
		}
	}
	if got := weaponStaminaByTypeExportCall4F7E80(0xf80000ff); got != 10 {
		t.Fatalf("irrelevant high/low bits export cost = %d, want 10", got)
	}
	if got := weaponStaminaByTypeExportCall4F7E80(0xffffffff); got != 70 {
		t.Fatalf("all bits export cost = %d, want 70", got)
	}
}
