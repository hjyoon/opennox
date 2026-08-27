package server

import "encoding/binary"

// NetworkTryGetPacketSize51BAD0 is the exact MSG_TRY_GET packet width.
const NetworkTryGetPacketSize51BAD0 = networkTryGetPacketSize51BAD0

// NetworkTryGetRuntime51BAD0 supplies state and effects owned by the root and
// legacy layers while the packet decision remains native Go.
type NetworkTryGetRuntime51BAD0 struct {
	NetDebug        func() bool
	TestHighBit     func(uint16)
	GameBlocked     func() bool
	Pickup          func(*Object, *Object)
	CarryingTooMuch func(*Object)
}

// NetworkTryGet51BAD0 binds MSG_TRY_GET to native Object, PlayerUpdateData,
// and Player layouts on every pointer width.
func (s *Server) NetworkTryGet51BAD0(
	unit *Object,
	update *PlayerUpdateData,
	packet *[NetworkTryGetPacketSize51BAD0]byte,
	runtime NetworkTryGetRuntime51BAD0,
) int32 {
	return networkTryGet51BAD0(unit, update, networkTryGetHooks51BAD0[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadWireCode: func() uint16 {
			return binary.LittleEndian.Uint16(packet[1:3])
		},
		dynamicUnitCode: s.packetDynamicUnitCode578B40,
		netDebug:        runtime.NetDebug,
		testHighBit:     runtime.TestHighBit,
		gameBlocked:     runtime.GameBlocked,
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerStatus: func(player *Player) uint32 {
			return player.Field3680
		},
		loadTradeActive: func(update *PlayerUpdateData) bool {
			return update.Trade70 != nil
		},
		loadDialogActive: func(update *PlayerUpdateData) bool {
			return update.DialogWith != nil
		},
		loadUnitFlagsLow: func(unit *Object) uint8 {
			return uint8(unit.ObjFlags)
		},
		objectFromNetCode: s.ObjectFromNetCode4ECCB0,
		loadInventoryFirst: func(unit *Object) *Object {
			return unit.InvFirstItem
		},
		loadInventoryNext: func(item *Object) *Object {
			return item.InvNextItem
		},
		loadWeight: func(item *Object) uint8 {
			return item.Weight
		},
		loadCarryCapacity: func(unit *Object) uint16 {
			return unit.CarryCapacity
		},
		pickup:          runtime.Pickup,
		carryingTooMuch: runtime.CarryingTooMuch,
	})
}
