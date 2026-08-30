package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const playerWeaponUseFlagsOffset4FA2B0 = uintptr(96)

func playerWeaponAnimationNative4FA280(equip uint32) int32 {
	return playerWeaponAnimation4FA280(equip, func(bit int) uint32 {
		return memmap.Uint32(0x587000, uintptr(215824+4*bit))
	})
}

func playerActionStateNative4FA2B0(
	s *Server,
	unit *Object,
	weaponAnimation func(uint32) int32,
) int32 {
	return playerActionState4FA2B0(unit, playerActionStateHooks4FA2B0[
		*Object, *PlayerUpdateData, *Player, *Object, unsafe.Pointer,
	]{
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadState: func(update *PlayerUpdateData) uint8 {
			return uint8(update.State)
		},
		isAbilityActive: func(unit *Object, ability Ability) bool {
			return s.Abils.IsActive(unit, ability)
		},
		isWarcryActive: func(unit *Object, ability Ability) bool {
			return s.Abils.IsActiveVal(unit, ability)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadWeaponEquip: func(player *Player) uint32 {
			return player.WeaponEquip
		},
		loadEquippedWeapon: func(update *PlayerUpdateData) *Object {
			return update.EquippedWeapon
		},
		loadWeaponUseData: func(weapon *Object) unsafe.Pointer {
			return weapon.UseData.Ptr
		},
		loadWeaponFlags: func(useData unsafe.Pointer) uint8 {
			return *(*uint8)(unsafe.Add(useData, playerWeaponUseFlagsOffset4FA2B0))
		},
		loadAnimationVariant: func(player *Player) uint8 {
			return uint8(player.Field8)
		},
		weaponAnimation: weaponAnimation,
	})
}

// PlayerActionState4FA2B0 maps the player's current action to its network
// animation state using native-width Object, update-data, Player, weapon, and
// use-data pointers throughout.
//
//go:noinline
func (s *Server) PlayerActionState4FA2B0(unit *Object) int32 {
	return playerActionStateNative4FA2B0(s, unit, playerWeaponAnimationNative4FA280)
}
