package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
)

// ScriptRaiseZombieRaw516D00 retains the C entry's mixed return: the original
// object pointer on a no-op, or the raise helper's zero-extended DWORD result.
// The safe bool script wrapper remains separate from this raw C contract.
func ScriptRaiseZombieRaw516D00(unit *Object, raise func(*Object) uint32) uintptr {
	result := uintptr(unsafe.Pointer(unit))
	if unit != nil && uint8(unit.ObjClass)&uint8(object.ClassMonster) != 0 && unit.ObjFlags.Has(object.FlagDead) {
		unit.UpdateDataMonster().StatusFlags &^= object.MonStatusStayDead
		result = uintptr(raise(unit))
	}
	return result
}
