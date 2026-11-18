package legacy

/*
#include "GAME4_1.h"
#include "server__system__trade.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func sendShopTradePacket(world *server.Server, recipient *server.Object, packet []byte, sequence int) C.int {
	update := (*server.PlayerUpdateData)(recipient.UpdateData)
	return C.int(world.NetSendPacketXxx(int(update.Player.PlayerInd), packet, nil, 1, sequence))
}

func sendShopItemRemove50E820(world *server.Server, recipient, item *server.Object) C.int {
	packet := server.BuildShopItemRemovePacket50E820(item)
	return sendShopTradePacket(world, recipient, packet[:], 1)
}

//export sub_50E820
func sub_50E820(recipient, item *nox_object_t) C.int {
	return sendShopItemRemove50E820(game4TradeRuntime().world, asObjectS(recipient), asObjectS(item))
}

//export sub_50F450
func sub_50F450(recipient *nox_object_t) C.int {
	packet := server.BuildTradeP2PClosePacket50F450()
	return sendShopTradePacket(game4TradeRuntime().world, asObjectS(recipient), packet[:], 0)
}

//export sub_5104F0
func sub_5104F0(recipient *nox_object_t, amount C.short) C.int {
	packet := server.BuildShopMissingGoldPacket5104F0(uint16(amount))
	return sendShopTradePacket(game4TradeRuntime().world, asObjectS(recipient), packet[:], 0)
}

//export sub_50E7A0
func sub_50E7A0(ptr unsafe.Pointer, item *nox_object_t) C.int {
	world := game4TradeRuntime().world
	if world.DetachShopItemNative50E7A0((*server.TradeSession)(ptr), asObjectS(item), func(recipient, removed *server.Object) {
		sendShopItemRemove50E820(world, recipient, removed)
	}) {
		return 1
	}
	return 0
}
