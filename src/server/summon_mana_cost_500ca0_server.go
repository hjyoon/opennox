package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const summonManaCostBlobBase500CA0 = uint32(0x00587000)

func summonManaCostBlobOffset500CA0(address uint32) uintptr {
	return uintptr(address - summonManaCostBlobBase500CA0)
}

func summonManaCostLoad500CA0(address uint32) int32 {
	return memmap.Int32(
		uintptr(summonManaCostBlobBase500CA0),
		summonManaCostBlobOffset500CA0(address),
	)
}

func summonManaCostNative500CA0(spellID int32, unit *Object, loadCost func(uint32) int32) int32 {
	return summonManaCost500CA0(spellID, unit, summonManaCostHooks500CA0[*Object]{
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadCost: loadCost,
	})
}

// SummonManaCost500CA0 binds GAME.EXE 00500CA0 to a native-width Object
// pointer and the packed PE32 data blob. Invalid, unchecked spell IDs retain
// the oracle's uint32 address calculation; an address outside the registered
// blob faults instead of being silently remapped by native-width arithmetic.
func SummonManaCost500CA0(spellID int32, unit *Object) int32 {
	return summonManaCostNative500CA0(spellID, unit, summonManaCostLoad500CA0)
}

var _ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
