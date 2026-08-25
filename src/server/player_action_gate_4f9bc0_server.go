package server

import (
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func playerCanMoveNative4F9BC0(unit *Object, quest bool) int32 {
	if unit.HasEnchant(ENCHANT_FREEZE) || unit.HasEnchant(ENCHANT_HELD) {
		return 0
	}
	update := unit.UpdateDataPlayer()
	if quest && update.Trade70 != nil {
		return 0
	}
	if update.State == PlayerState1 {
		weapon := update.EquippedWeapon
		if weapon != nil && weapon.Class().Has(object.ClassWeapon) && weapon.WeaponClass().Has(object.WeaponCrossbow) {
			return 0
		}
	}
	return 1
}

// PlayerCanMove4F9BC0 preserves GAME.EXE 004F9BC0 while following the
// native-width Object and PlayerUpdateData pointers on 64-bit targets.
func PlayerCanMove4F9BC0(unit *Object) int32 {
	return playerCanMoveNative4F9BC0(unit, noxflags.HasGame(noxflags.GameModeQuest))
}

func playerCanAttackNative4F9C40(unit *Object) int32 {
	if unit.HasEnchant(ENCHANT_FREEZE) || unit.UpdateDataPlayer().State == PlayerState23 {
		return 0
	}
	return 1
}

// PlayerCanAttack4F9C40 preserves GAME.EXE 004F9C40 without reading the
// original Win32 update-data pointer slot directly.
func PlayerCanAttack4F9C40(unit *Object) int32 {
	return playerCanAttackNative4F9C40(unit)
}
