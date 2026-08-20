package server

import "unsafe"

type playerManaRefreshNativeDeps4EECF0 struct {
	protectMana func(uint32, int16) uintptr
}

func playerManaRefreshNative4EECF0(
	unit *Object,
	deps playerManaRefreshNativeDeps4EECF0,
) uintptr {
	return playerManaRefresh4EECF0(playerManaRefreshHooks4EECF0[
		*Object,
		*PlayerUpdateData,
		*Player,
		uintptr,
	]{
		loadUnitArg: func() (*Object, uintptr) {
			return unit, uintptr(unsafe.Pointer(unit))
		},
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadCurrent: func(update *PlayerUpdateData) uint16 {
			return update.ManaCur
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		storePrevious: func(update *PlayerUpdateData, current uint16) {
			update.ManaPrev = current
		},
		loadMaximum: func(update *PlayerUpdateData) uint16 {
			return update.ManaMax
		},
		storeCurrent: func(update *PlayerUpdateData, maximum uint16) {
			update.ManaCur = maximum
		},
		loadProtection: func(player *Player) uint32 {
			return player.ProtUnitManaCur
		},
		protectMana: deps.protectMana,
	})
}

// PlayerManaRefresh4EECF0 binds GAME.EXE 004EECF0 to native pointer-width
// server layouts. Its uintptr result preserves EAX without converting a
// protection-service return value into a Go pointer. The legacy boundary uses
// only the fixed-width protection token and signed mana word.
func (*Server) PlayerManaRefresh4EECF0(
	unit *Object,
	protectMana func(uint32, int16) uintptr,
) uintptr {
	return playerManaRefreshNative4EECF0(unit, playerManaRefreshNativeDeps4EECF0{
		protectMana: protectMana,
	})
}
