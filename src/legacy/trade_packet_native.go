package legacy

import (
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

func tradePacketBalance50FA00(recipient *server.Object, session *server.TradeSession) int32 {
	packet := server.BuildTradeP2PBalancePacket50FA00(session, recipient)
	return tradePacketSendNative(recipient, packet[:])
}

func tradePacketRemove50FF90(recipient *server.Object, _ unsafe.Pointer, item *server.Object) int32 {
	packet := server.BuildTradeP2PRemovePacket50FF90(item)
	return tradePacketSendNative(recipient, packet[:])
}
