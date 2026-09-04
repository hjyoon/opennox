package server

import "unsafe"

type playerManaRechargeNativeDeps4FD030 struct {
	addMana func(*Object, int16) uint16
}

func playerManaRechargeNative4FD030(
	unit *Object,
	amount int16,
	deps playerManaRechargeNativeDeps4FD030,
) uint16 {
	return playerManaRecharge4FD030(playerManaRechargeHooks4FD030[*Object]{
		loadUnitArg: func() (*Object, uint16) {
			// GAME.EXE initializes AX from the object argument. Only that
			// numeric low word is returned by the non-Player path.
			return unit, uint16(uintptr(unsafe.Pointer(unit)))
		},
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadAmountArg: func() int16 {
			return amount
		},
		addMana: deps.addMana,
	})
}

// PlayerManaRecharge4FD030 binds GAME.EXE 004FD030 to a native-width Object
// pointer and a fixed-width signed mana amount. addMana owns the already
// restored 004EEB80 service and its exact uint16 result is returned.
//
//go:noinline
func (*Server) PlayerManaRecharge4FD030(
	unit *Object,
	amount int16,
	addMana func(*Object, int16) uint16,
) uint16 {
	return playerManaRechargeNative4FD030(unit, amount, playerManaRechargeNativeDeps4FD030{
		addMana: addMana,
	})
}
