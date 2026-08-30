package server

const (
	weaponStaminaBowFlag4F7E80          = uint32(0x00000200)
	weaponStaminaCrossbowFlag4F7E80     = uint32(0x00004000)
	weaponStaminaStaffFlag4F7E80        = uint32(0x00000800)
	weaponStaminaSmallMeleeFlag4F7E80   = uint32(0x00000100)
	weaponStaminaMediumMeleeFlag4F7E80  = uint32(0x00001000)
	weaponStaminaHeavyMeleeFlag4F7E80   = uint32(0x00002000)
	weaponStaminaSpecialMeleeMask4F7E80 = uint32(0x07ff8000)
	weaponStaminaThrownFlag4F7E80       = uint32(0x00000400)
)

// weaponStaminaByType4F7E80 preserves GAME.EXE 004F7E80. The input is the
// complete PE32 weapon-flag dword. Multiple matching groups return the first
// cost in the original branch order; bits outside the tested groups are
// ignored.
func weaponStaminaByType4F7E80(flags uint32) int32 {
	if flags&weaponStaminaBowFlag4F7E80 != 0 {
		return 70
	}
	if flags&weaponStaminaCrossbowFlag4F7E80 != 0 {
		return 100
	}
	if flags&weaponStaminaStaffFlag4F7E80 != 0 {
		return 50
	}
	if flags&weaponStaminaSmallMeleeFlag4F7E80 != 0 {
		return 45
	}
	if flags&weaponStaminaMediumMeleeFlag4F7E80 != 0 {
		return 75
	}
	if flags&weaponStaminaHeavyMeleeFlag4F7E80 != 0 {
		return 100
	}
	if flags&weaponStaminaSpecialMeleeMask4F7E80 != 0 {
		return 45
	}
	if flags&weaponStaminaThrownFlag4F7E80 != 0 {
		return 75
	}
	return 10
}
