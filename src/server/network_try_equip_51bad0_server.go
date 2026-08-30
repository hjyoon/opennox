package server

import "encoding/binary"

// NetworkTryEquipPacketSize51BAD0 is the exact MSG_TRY_EQUIP packet width.
const NetworkTryEquipPacketSize51BAD0 = networkTryEquipPacketSize51BAD0

// NetworkTryEquipRuntime51BAD0 supplies engine observations and the restored
// native-width 004F2F70 equip path.
type NetworkTryEquipRuntime51BAD0 struct {
	NetDebug    func() bool
	TestHighBit func(uint16)
	GameBlocked func() bool
	TryEquip    func(*Object, *Object)
}

// NetworkTryEquip51BAD0 binds the original packet contract to native Object,
// PlayerUpdateData, and Player layouts on every pointer width.
func (s *Server) NetworkTryEquip51BAD0(
	unit *Object,
	update *PlayerUpdateData,
	packet *[NetworkTryEquipPacketSize51BAD0]byte,
	runtime NetworkTryEquipRuntime51BAD0,
) int32 {
	return networkTryEquip51BAD0(unit, update, networkTryEquipHooks51BAD0[
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
		findItemByCode: EquippedItemByCode4F7920,
		tryEquip:       runtime.TryEquip,
	})
}
