//go:build 386 || arm

package server

import "unsafe"

func storeSecondaryWeaponABI32(update *PlayerUpdateData, item *Object) {
	update.Field27 = uint32(uintptr(unsafe.Pointer(item)))
}
