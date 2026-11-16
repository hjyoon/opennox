package legacy

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Tests replace only the declared network service. Packet construction and
// native object/session traversal always use the production implementation.
var tradePacketSendForTest func(int, []byte, *server.Object, int) int

func tradePacketSendNative(recipient *server.Object, packet []byte) int32 {
	update := (*server.PlayerUpdateData)(recipient.UpdateData)
	index := int(update.Player.PlayerInd)
	if send := tradePacketSendForTest; send != nil {
		return int32(send(index, packet, nil, 1))
	}
	return int32(GetServer().S().NetSendPacketXxx1(index, packet, nil, 1))
}

func tradePacketStatus50F720(recipient *server.Object, session *server.TradeSession) int32 {
	packet := server.BuildTradeP2PStatusPacket50F720(session, recipient)
	return tradePacketSendNative(recipient, packet[:])
}

// The existing P2P builder has a zero delta for two players. This entry also
// preserves the original merchant branches and uint32 subtraction wrapping.
// Its original contract requires a participant and initialized Gold records.
func tradePacketBalance50FA00(recipient *server.Object, session *server.TradeSession) int32 {
	packet := [14]byte{0xc9, 0x06}
	if session.Field8 == recipient {
		binary.LittleEndian.PutUint32(packet[2:6], session.Field48.InitDataGold().Amount)
		if uint32(session.Field12.ObjClass)&4 == 0 {
			binary.LittleEndian.PutUint32(packet[6:10], session.Field44-session.Field40)
		}
		binary.LittleEndian.PutUint32(packet[10:14], session.Field52.InitDataGold().Amount)
	} else if session.Field12 == recipient {
		binary.LittleEndian.PutUint32(packet[2:6], session.Field52.InitDataGold().Amount)
		if uint32(session.Field8.ObjClass)&4 == 0 {
			binary.LittleEndian.PutUint32(packet[6:10], session.Field40-session.Field44)
		}
		binary.LittleEndian.PutUint32(packet[10:14], session.Field48.InitDataGold().Amount)
	}
	return tradePacketSendNative(recipient, packet[:])
}

func tradePacketRemove50FF90(recipient *server.Object, _ unsafe.Pointer, item *server.Object) int32 {
	packet := server.BuildTradeP2PRemovePacket50FF90(item)
	return tradePacketSendNative(recipient, packet[:])
}
