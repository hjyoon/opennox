package server

import (
	"encoding/binary"

	"github.com/opennox/libs/types"
)

// NetworkInventoryFailPacketSize51BAD0 is the exact MSG_INVENTORY_FAIL packet
// width.
const NetworkInventoryFailPacketSize51BAD0 = networkInventoryFailPacketSize51BAD0

type NetworkInventoryFailRuntime51BAD0 struct {
	Drop            func(*Object, *Object, *types.Pointf)
	CarryingTooMuch func(*Object)
}

// NetworkInventoryFail51BAD0 binds MSG_INVENTORY_FAIL to the native Object
// layout on every pointer width.
func (s *Server) NetworkInventoryFail51BAD0(
	unit *Object,
	packet *[NetworkInventoryFailPacketSize51BAD0]byte,
	runtime NetworkInventoryFailRuntime51BAD0,
) int32 {
	return networkInventoryFail51BAD0(unit, networkInventoryFailHooks51BAD0[*Object]{
		loadCode: func() uint16 {
			return binary.LittleEndian.Uint16(packet[1:3])
		},
		findItem: EquippedItemByCode4F7920,
		loadPosition: func(unit *Object) *types.Pointf {
			return &unit.PosVec
		},
		drop:          runtime.Drop,
		carryingHeavy: runtime.CarryingTooMuch,
	})
}
