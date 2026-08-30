package server

import "encoding/binary"

// NetworkInfoBookPacketSize51BAD0 is the exact MSG_INFO_BOOK_DATA request and
// response width.
const NetworkInfoBookPacketSize51BAD0 = networkInfoBookPacketSize51BAD0

type NetworkInfoBookRuntime51BAD0 struct {
	NetDebug    func() bool
	TestHighBit func(uint16)
	Send        func(uint8, [NetworkInfoBookPacketSize51BAD0]byte)
}

func tradeItemByCode510DE0(update *PlayerUpdateData, code uint32) *Object {
	if update.Trade70 == nil {
		return nil
	}
	for node := update.Trade70.Field20; node != nil; node = node.Field8 {
		if node.Item0.NetCode == code {
			return node.Item0
		}
	}
	return nil
}

// NetworkInfoBook51BAD0 binds MSG_INFO_BOOK_DATA to native Object,
// PlayerUpdateData, TradeSession, and TradeItem layouts on every pointer width.
func (s *Server) NetworkInfoBook51BAD0(
	unit *Object,
	update *PlayerUpdateData,
	packet *[NetworkInfoBookPacketSize51BAD0]byte,
	runtime NetworkInfoBookRuntime51BAD0,
) int32 {
	return networkInfoBook51BAD0(unit, update, networkInfoBookHooks51BAD0[
		*Object,
		*PlayerUpdateData,
	]{
		loadWireCode: func() uint16 {
			return binary.LittleEndian.Uint16(packet[1:3])
		},
		dynamicUnitCode: s.packetDynamicUnitCode578B40,
		netDebug:        runtime.NetDebug,
		testHighBit:     runtime.TestHighBit,
		findInventory:   EquippedItemByCode4F7920,
		findTrade:       tradeItemByCode510DE0,
		findWorld:       s.ObjectFromNetCode4ECCB0,
		unitCode: func(item *Object) uint16 {
			return uint16(s.GetUnitNetCode(item))
		},
		loadKind: func() uint8 {
			return packet[3]
		},
		loadDefaultInfo: func(item *Object) uint8 {
			return *(*uint8)(item.UseData.Ptr)
		},
		loadGuideName: func(item *Object) string {
			return item.UseDataFieldGuide().Creature()
		},
		guideID: func(name string) uint8 {
			return uint8(RewardFieldGuideID4F0D20(name))
		},
		loadRecipient: func(update *PlayerUpdateData) uint8 {
			return update.Player.PlayerInd
		},
		send: runtime.Send,
	})
}
