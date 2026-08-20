package server

import "unsafe"

type playerManaAddNativeDeps4EEB80 struct {
	protectMana         func(uint32, int16)
	protectPlayerHPMana func(uint32, uint16) uint16
}

func playerManaAddNative4EEB80(
	unit *Object,
	amount int32,
	deps playerManaAddNativeDeps4EEB80,
) uint16 {
	return playerManaAdd4EEB80(playerManaAddHooks4EEB80[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadUnitArg: func() (*Object, uint16) {
			// GAME.EXE returns AX from the object argument on both entry-gate
			// exits. Only the numeric low word is observed; it is never turned
			// back into a pointer.
			return unit, uint16(uintptr(unsafe.Pointer(unit)))
		},
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadAmountArg: func() int32 {
			return amount
		},
		loadCurrent: func(update *PlayerUpdateData) uint16 {
			return update.ManaCur
		},
		loadMaximum: func(update *PlayerUpdateData) uint16 {
			return update.ManaMax
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
		protectMana:         deps.protectMana,
		protectPlayerHPMana: deps.protectPlayerHPMana,
	})
}

// PlayerManaAdd4EEB80 binds GAME.EXE 004EEB80 to native pointer-width server
// layouts. The legacy protection services cross this boundary only through
// fixed-width tokens, deltas, values, and results.
func (*Server) PlayerManaAdd4EEB80(
	unit *Object,
	amount int32,
	protectMana func(uint32, int16),
	protectPlayerHPMana func(uint32, uint16) uint16,
) uint16 {
	return playerManaAddNative4EEB80(unit, amount, playerManaAddNativeDeps4EEB80{
		protectMana:         protectMana,
		protectPlayerHPMana: protectPlayerHPMana,
	})
}
