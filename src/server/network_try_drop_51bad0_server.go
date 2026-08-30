package server

import (
	"encoding/binary"

	"github.com/opennox/libs/types"
)

// NetworkTryDropPacketSize51BAD0 is the exact MSG_TRY_DROP packet width.
const NetworkTryDropPacketSize51BAD0 = networkTryDropPacketSize51BAD0

// NetworkTryDropRuntime51BAD0 supplies the engine-debug observation and the
// already restored native-width 004ED810 drop path.
type NetworkTryDropRuntime51BAD0 struct {
	NetDebug    func() bool
	TestHighBit func(uint16)
	Drop        func(*Object, *Object, *types.Pointf)
}

func (s *Server) packetDynamicUnitCode578B40(code uint16) uint32 {
	if code&0x8000 == 0 {
		return uint32(code)
	}
	obj := s.ObjectByExtent4ED020(uint32(code & 0x7fff))
	if obj == nil {
		return 0
	}
	return obj.NetCode
}

// NetworkTryDrop51BAD0 binds the original packet contract to native Object,
// PlayerUpdateData, and Player layouts on every pointer width.
func (s *Server) NetworkTryDrop51BAD0(
	unit *Object,
	update *PlayerUpdateData,
	packet *[NetworkTryDropPacketSize51BAD0]byte,
	runtime NetworkTryDropRuntime51BAD0,
) int32 {
	return networkTryDrop51BAD0(unit, update, networkTryDropHooks51BAD0[
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
		loadTradeActive: func(update *PlayerUpdateData) bool {
			return update.Trade70 != nil
		},
		loadDialogActive: func(update *PlayerUpdateData) bool {
			return update.DialogWith != nil
		},
		loadUnitFlagsLow: func(unit *Object) uint8 {
			return uint8(unit.ObjFlags)
		},
		findItemByCode: EquippedItemByCode4F7920,
		loadDestinationX: func() uint16 {
			return binary.LittleEndian.Uint16(packet[3:5])
		},
		loadDestinationY: func() uint16 {
			return binary.LittleEndian.Uint16(packet[5:7])
		},
		drop: runtime.Drop,
	})
}
