package server

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

type playerManaSubNativeDeps4EEBF0 struct {
	loadEngineGodMode func() bool
	protectMana       func(uint32, int16) uintptr
}

func playerManaSubNative4EEBF0(
	unit *Object,
	amount int32,
	deps playerManaSubNativeDeps4EEBF0,
) uintptr {
	return playerManaSub4EEBF0(playerManaSubHooks4EEBF0[
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
		loadEngineGodMode: deps.loadEngineGodMode,
		loadUpdateData: func(unit *Object) (*PlayerUpdateData, uintptr) {
			update := (*PlayerUpdateData)(unit.UpdateData)
			return update, uintptr(unsafe.Pointer(update))
		},
		loadCurrent: func(update *PlayerUpdateData) uint16 {
			return update.ManaCur
		},
		loadAmountArg: func() int32 {
			return amount
		},
		storePrevious: func(update *PlayerUpdateData, value uint16) {
			update.ManaPrev = value
		},
		storeCurrent: func(update *PlayerUpdateData, value uint16) {
			update.ManaCur = value
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadProtection: func(player *Player) uint32 {
			return player.ProtUnitManaCur
		},
		protectMana: deps.protectMana,
	})
}

// PlayerManaSub4EEBF0 binds GAME.EXE 004EEBF0 to native pointer-width server
// layouts. Its uintptr result preserves the original return register without
// turning an accidental pointer-like value into a Go pointer. It must not be
// converted back to a pointer. The legacy protection service crosses this
// boundary only through a fixed-width token and delta.
func (*Server) PlayerManaSub4EEBF0(
	unit *Object,
	amount int32,
	protectMana func(uint32, int16) uintptr,
) uintptr {
	return playerManaSubNative4EEBF0(unit, amount, playerManaSubNativeDeps4EEBF0{
		loadEngineGodMode: func() bool {
			return noxflags.HasEngine(noxflags.EngineGodMode)
		},
		protectMana: protectMana,
	})
}
