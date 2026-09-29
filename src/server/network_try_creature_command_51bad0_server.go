package server

import "encoding/binary"

// NetworkTryCreatureCommandPacketSize51BAD0 is the exact
// MSG_TRY_CREATURE_COMMAND packet width.
const NetworkTryCreatureCommandPacketSize51BAD0 = networkTryCreatureCommandPacketSize51BAD0

// NetworkTryCreatureCommandRuntime51BAD0 owns the debug observation and the
// already restored native-width 00533900 order dispatch.
type NetworkTryCreatureCommandRuntime51BAD0 struct {
	NetDebug    func() bool
	TestHighBit func(uint16)
	OrderUnit   func(*Object, *Object, uint32)
}

// NetworkTryCreatureCommand51BAD0 binds the packet decoder to native Object,
// PlayerUpdateData, and Player layouts on every pointer width.
func (s *Server) NetworkTryCreatureCommand51BAD0(
	unit *Object,
	update *PlayerUpdateData,
	packet *[NetworkTryCreatureCommandPacketSize51BAD0]byte,
	runtime NetworkTryCreatureCommandRuntime51BAD0,
) int32 {
	return networkTryCreatureCommand51BAD0(unit, update, networkTryCreatureCommandHooks51BAD0[
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
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerStatus: func(player *Player) uint32 {
			return player.Field3680
		},
		objectFromNetCode: s.ObjectFromNetCode4ECCB0,
		loadOrder: func() uint8 {
			return packet[3]
		},
		orderUnit: runtime.OrderUnit,
	})
}
