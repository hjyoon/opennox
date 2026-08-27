package server

const playerTryDequipFlag4F2FB0 = int32(1)

// PlayerTryDequipFunc4F2FB0 is the fixed-width callback contract used by the
// two lower dequip routines called from GAME.EXE 004F2FB0.
type PlayerTryDequipFunc4F2FB0 func(owner, item *Object, flag1, flag2 int32) int32

// PlayerTryDequip4F2FB0 preserves GAME.EXE 004F2FB0. It always attempts the
// weapon path first with flags 1 and 1. Any nonzero weapon result returns one
// without calling the armor path. Armor is attempted with the same flags only
// after a zero weapon result, and its nonzero result is likewise normalized to
// one. The wrapper does not dereference or validate either object pointer.
//
//go:noinline
func PlayerTryDequip4F2FB0(
	owner, item *Object,
	tryWeapon, tryArmor PlayerTryDequipFunc4F2FB0,
) int32 {
	if tryWeapon(owner, item, playerTryDequipFlag4F2FB0, playerTryDequipFlag4F2FB0) != 0 {
		return 1
	}
	if tryArmor(owner, item, playerTryDequipFlag4F2FB0, playerTryDequipFlag4F2FB0) != 0 {
		return 1
	}
	return 0
}
