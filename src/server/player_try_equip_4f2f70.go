package server

const playerTryEquipFlag4F2F70 = int32(1)

// PlayerTryEquipFunc4F2F70 is the fixed-width callback contract used by the
// two lower equip routines called from GAME.EXE 004F2F70.
type PlayerTryEquipFunc4F2F70 func(owner, item *Object, flag1, flag2 int32) int32

// PlayerTryEquip4F2F70 preserves GAME.EXE 004F2F70. It always attempts the
// weapon path first with flags 1 and 1. Any nonzero weapon result returns one
// without calling the armor path. Armor is attempted with the same flags only
// after a zero weapon result, and its nonzero result is likewise normalized to
// one. The wrapper does not dereference or validate either object pointer.
//
//go:noinline
func PlayerTryEquip4F2F70(
	owner, item *Object,
	tryWeapon, tryArmor PlayerTryEquipFunc4F2F70,
) int32 {
	if tryWeapon(owner, item, playerTryEquipFlag4F2F70, playerTryEquipFlag4F2F70) != 0 {
		return 1
	}
	if tryArmor(owner, item, playerTryEquipFlag4F2F70, playerTryEquipFlag4F2F70) != 0 {
		return 1
	}
	return 0
}
