package server

import "unsafe"

func playerGetMaxManaNative4EECB0(unit *Object) uint16 {
	return playerGetMaxMana4EECB0(playerGetMaxManaHooks4EECB0[*Object, *PlayerUpdateData]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadClassLow: func(unit *Object) uint8 {
			return *(*uint8)(unsafe.Pointer(&unit.ObjClass))
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadMaximum: func(update *PlayerUpdateData) uint16 {
			return update.ManaMax
		},
	})
}

// PlayerGetMaxMana4EECB0 returns the exact AX-width maximum-mana value from
// GAME.EXE 004EECB0 using native-width object and update-data pointers.
func PlayerGetMaxMana4EECB0(unit *Object) uint16 {
	return playerGetMaxManaNative4EECB0(unit)
}
