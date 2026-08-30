package server

import "encoding/binary"

// NetworkTryDequipPacketSize51BAD0 is the exact MSG_TRY_DEQUIP packet width.
const NetworkTryDequipPacketSize51BAD0 = networkTryDequipPacketSize51BAD0

// NetworkTryDequipRuntime51BAD0 supplies engine observations and the restored
// native-width 004F2FB0 dequip path.
type NetworkTryDequipRuntime51BAD0 struct {
	NetDebug    func() bool
	TestHighBit func(uint16)
	TryDequip   func(*Object, *Object)
}

// NetworkTryDequip51BAD0 binds the original packet contract to native Object,
// PlayerUpdateData, and Player layouts on every pointer width.
func (s *Server) NetworkTryDequip51BAD0(
	unit *Object,
	update *PlayerUpdateData,
	packet *[NetworkTryDequipPacketSize51BAD0]byte,
	runtime NetworkTryDequipRuntime51BAD0,
) int32 {
	return networkTryDequip51BAD0(unit, update, networkTryDequipHooks51BAD0[
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
		findItemByCode: EquippedItemByCode4F7920,
		loadState: func(update *PlayerUpdateData) uint8 {
			return uint8(update.State)
		},
		loadItemClass: func(item *Object) uint32 {
			return uint32(item.ObjClass)
		},
		loadItemSubclass: func(item *Object) uint32 {
			return uint32(item.ObjSubClass)
		},
		tryDequip: runtime.TryDequip,
	})
}
