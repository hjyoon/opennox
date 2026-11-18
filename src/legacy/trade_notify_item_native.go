package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func tradeNotifyEnd50F560(recipient *server.Object) int32 {
	packet := [2]byte{0xc9, 0x02}
	return tradePacketSendNative(recipient, packet[:])
}

func tradeNotifyContinue50F6E0(recipient *server.Object) int32 {
	packet := server.BuildTradeContinuePacket50F6E0()
	return tradePacketSendNative(recipient, packet[:])
}

func tradeNotifyItem50FAE0(recipient, offering *server.Object, _ unsafe.Pointer, item *server.Object, cost uint32) int32 {
	packet := server.BuildTradeP2PItemPacket50FAE0(recipient, offering, item, cost)
	return tradePacketSendNative(recipient, packet[:])
}
