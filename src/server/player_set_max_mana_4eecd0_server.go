package server

import "unsafe"

func playerSetMaxManaNative4EECD0(unit *Object, maximum uint16) uintptr {
	return playerSetMaxMana4EECD0(playerSetMaxManaHooks4EECD0[
		*Object,
		*PlayerUpdateData,
		uintptr,
	]{
		loadUnitArg: func() (*Object, uintptr) {
			return unit, uintptr(unsafe.Pointer(unit))
		},
		loadClassLow: func(unit *Object) uint8 {
			return *(*uint8)(unsafe.Pointer(&unit.ObjClass))
		},
		loadUpdateData: func(unit *Object) (*PlayerUpdateData, uintptr) {
			update := (*PlayerUpdateData)(unit.UpdateData)
			return update, uintptr(unsafe.Pointer(update))
		},
		loadMaximumArg: func() uint16 {
			return maximum
		},
		storeMaximum: func(update *PlayerUpdateData, maximum uint16) {
			update.ManaMax = maximum
		},
	})
}

// PlayerSetMaxMana4EECD0 stores the exact low maximum-mana word through
// native-width Object and PlayerUpdateData pointers. The uintptr result keeps
// the original unit-or-UpdateData return register as an integer identity and
// must not be converted back into a Go pointer.
func PlayerSetMaxMana4EECD0(unit *Object, maximum uint16) uintptr {
	return playerSetMaxManaNative4EECD0(unit, maximum)
}
