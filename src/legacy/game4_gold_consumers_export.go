package legacy

/*
#include "game4_gold_consumers.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_50F3A0
func sub_50F3A0(ptr unsafe.Pointer) {
	rt := game4TradeRuntime()
	rt.world.CancelP2PTradeNative50F3A0((*server.TradeSession)(ptr), rt.p2p())
}

//export sub_50F7F0
func sub_50F7F0(owner, gold *nox_object_t) C.uintptr_t {
	return C.uintptr_t(game4GoldReturn50F7F0(asObjectS(owner), asObjectS(gold)))
}

//export sub_50FB90
func sub_50FB90(ptr unsafe.Pointer) C.uintptr_t {
	return C.uintptr_t(game4TradeReprice50FB90((*server.TradeSession)(ptr)))
}

//export sub_5100C0_trade
func sub_5100C0_trade(owner *nox_object_t, ptr unsafe.Pointer, code C.uint16_t) {
	rt := game4TradeRuntime()
	rt.world.BuyShopItemNative5100C0(asObjectS(owner), (*server.TradeSession)(ptr), uint16(code), rt.buy())
}

//export sub_510640_trade
func sub_510640_trade(owner *nox_object_t, ptr unsafe.Pointer, typ C.uint16_t, count C.uint8_t) C.uintptr_t {
	rt := game4TradeRuntime()
	_, result := rt.world.BuyShopItemsByType510640(asObjectS(owner), (*server.TradeSession)(ptr), uint16(typ), uint8(count), rt.buy())
	return C.uintptr_t(result)
}

//export sub_510BE0_trade
func sub_510BE0_trade(owner *nox_object_t, ptr unsafe.Pointer, code C.uint16_t) C.uintptr_t {
	rt := game4TradeRuntime()
	_, result := rt.world.SellShopItem510BE0(asObjectS(owner), (*server.TradeSession)(ptr), uint16(code), rt.sell())
	return C.uintptr_t(result)
}

//export sub_510D10
func sub_510D10(owner *nox_object_t, ptr unsafe.Pointer, typ C.uint16_t, count C.uint8_t) {
	rt := game4TradeRuntime()
	rt.world.SellShopItemsByTypeNative510D10(asObjectS(owner), (*server.TradeSession)(ptr), uint16(typ), uint8(count), rt.sell())
}

//export sub_510AE0
func sub_510AE0(owner *nox_object_t, ptr unsafe.Pointer, code C.uint16_t) C.uintptr_t {
	rt := game4TradeRuntime()
	_, result := rt.world.RepairShopItem510AE0(asObjectS(owner), (*server.TradeSession)(ptr), uint16(code), rt.repair())
	return C.uintptr_t(result)
}

//export nox_server_netTradeNative_GAME4_51BAD0
func nox_server_netTradeNative_GAME4_51BAD0(data *C.uchar, size C.int, owner *nox_object_t) C.int {
	if size < 2 || data == nil {
		return -1
	}
	packet := unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size))
	return C.int(game4TradeRuntime().packet(asObjectS(owner), packet))
}
