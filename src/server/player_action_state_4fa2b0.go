package server

const playerRangedWeaponMask4FA2B0 = uint32(0x047f0000)

// playerActionStateHooks4FA2B0 exposes every observable load and call in
// GAME.EXE 004FA2B0 without assuming the width or layout of an Object pointer.
type playerActionStateHooks4FA2B0[O, U, P, W, D any] struct {
	loadUpdateData       func(O) U
	loadState            func(U) uint8
	isAbilityActive      func(O, Ability) bool
	isWarcryActive       func(O, Ability) bool
	loadPlayer           func(U) P
	loadWeaponEquip      func(P) uint32
	loadEquippedWeapon   func(U) W
	loadWeaponUseData    func(W) D
	loadWeaponFlags      func(D) uint8
	loadAnimationVariant func(P) uint8
	weaponAnimation      func(uint32) int32
}

// playerWeaponAnimation4FA280 preserves GAME.EXE 004FA280: it scans weapon
// bits 2 through 26 in ascending order and returns the full table entry for
// the first set bit.
func playerWeaponAnimation4FA280(equip uint32, loadAnimation func(bit int) uint32) int32 {
	for bit := 2; bit < PlayerWeaponCnt; bit++ {
		if equip&(uint32(1)<<bit) != 0 {
			return int32(loadAnimation(bit))
		}
	}
	return 0
}

// playerActionState4FA2B0 preserves GAME.EXE 004FA2B0, including ability
// priority, short-circuit order, and conditional pointer dereferences. The
// original function does not validate unit/update/player/weapon/use-data
// pointers, so those fault contracts deliberately remain with the hooks.
func playerActionState4FA2B0[O, U, P, W, D any](
	unit O,
	hooks playerActionStateHooks4FA2B0[O, U, P, W, D],
) int32 {
	update := hooks.loadUpdateData(unit)
	switch hooks.loadState(update) {
	case 0:
		return 4
	case 1, 14, 22:
		if hooks.isAbilityActive(unit, AbilityWarcry) && hooks.isWarcryActive(unit, AbilityWarcry) {
			return 46
		}
		if hooks.isAbilityActive(unit, AbilityBerserk) {
			return 45
		}
		player := hooks.loadPlayer(update)
		equip := hooks.loadWeaponEquip(player)
		if equip&playerRangedWeaponMask4FA2B0 != 0 {
			weapon := hooks.loadEquippedWeapon(update)
			useData := hooks.loadWeaponUseData(weapon)
			flags := hooks.loadWeaponFlags(useData)
			return int32((^flags & 2) | 0x1d)
		}
		if equip == 0 || equip == 1 {
			if variant := hooks.loadAnimationVariant(player); variant != 0 {
				return int32(variant)
			}
		}
		return hooks.weaponAnimation(equip)
	case 2, 10:
		return 21
	case 3:
		return 1
	case 4:
		return 2
	case 5:
		return 6
	case 12:
		return 3
	case 13:
		player := hooks.loadPlayer(update)
		if hooks.loadWeaponEquip(player)&0x400 != 0 {
			return 38
		}
		return 0
	case 15, 16, 17:
		return 40
	case 18:
		return 48
	case 19:
		return 49
	case 20:
		return 47
	case 21:
		return 30
	case 23:
		return 50
	case 24:
		return 19
	case 25:
		return 20
	case 26:
		return 15
	case 27, 28, 29:
		return 16
	case 30:
		return 52
	case 32:
		return 54
	default:
		return 0
	}
}
