package server

import "encoding/binary"

// NetworkTryUsePacketSize51BAD0 is the exact MSG_TRY_USE packet width.
const NetworkTryUsePacketSize51BAD0 = networkTryUsePacketSize51BAD0

type NetworkTryUseRuntime51BAD0 struct {
	NetDebug    func() bool
	TestHighBit func(uint16)
	GameBlocked func() bool
}

// NetworkTryUse51BAD0 binds MSG_TRY_USE to native Object, PlayerUpdateData,
// and Use callback layouts on every pointer width.
func (s *Server) NetworkTryUse51BAD0(
	unit *Object,
	update *PlayerUpdateData,
	packet *[NetworkTryUsePacketSize51BAD0]byte,
	runtime NetworkTryUseRuntime51BAD0,
) int32 {
	return networkTryUse51BAD0(unit, update, networkTryUseHooks51BAD0[
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
		useByNetCode: func(owner, item *Object) {
			_ = s.UseByNetCode53F8E0(owner, item)
		},
	})
}
