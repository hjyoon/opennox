package legacy

/*
#include <stdint.h>
#include "defs.h"
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_50F3A0
func sub_50F3A0(ptr unsafe.Pointer) {
	rt := game4TradeRuntime()
	rt.world.CancelP2PTradeNative50F3A0((*server.TradeSession)(ptr), rt.p2p())
}

//export nox_game4_goldReturn_50F7F0
func nox_game4_goldReturn_50F7F0(owner, gold *nox_object_t) C.uintptr_t {
	return C.uintptr_t(game4GoldReturn50F7F0(asObjectS(owner), asObjectS(gold)))
}

//export nox_game4_tradeReprice_50FB90
func nox_game4_tradeReprice_50FB90(ptr unsafe.Pointer) C.uintptr_t {
	return C.uintptr_t(game4TradeReprice50FB90((*server.TradeSession)(ptr)))
}

//export sub_5100C0_trade
func sub_5100C0_trade(owner *nox_object_t, ptr unsafe.Pointer, code C.uint16_t) {
	rt := game4TradeRuntime()
	rt.world.BuyShopItemNative5100C0(asObjectS(owner), (*server.TradeSession)(ptr), uint16(code), rt.buy())
}

//export nox_game4_shopBuyByType_510640
func nox_game4_shopBuyByType_510640(owner *nox_object_t, ptr unsafe.Pointer, typ C.uint16_t, count C.uint8_t) C.uintptr_t {
	rt := game4TradeRuntime()
	return C.uintptr_t(rt.world.BuyShopItemsByTypeNative510640(asObjectS(owner), (*server.TradeSession)(ptr), uint16(typ), uint8(count), rt.buy()))
}

//export nox_game4_shopSell_510BE0
func nox_game4_shopSell_510BE0(owner *nox_object_t, ptr unsafe.Pointer, code C.uint16_t) C.uintptr_t {
	rt := game4TradeRuntime()
	unit := asObjectS(owner)
	// The decoded callers discard EAX. Keep its meaningful original
	// inventory-present residue without interpreting a wire code as a pointer.
	result := uintptr(0)
	if unit.InvFirstItem != nil {
		result = uintptr(uint16(code))
	} else if glyph := rt.world.Types.ByID("Glyph"); glyph != nil {
		result = uintptr(uint32(glyph.Ind()))
	}
	rt.world.SellShopItemNative510BE0(unit, (*server.TradeSession)(ptr), uint16(code), rt.sell())
	runtime.KeepAlive(unit)
	return C.uintptr_t(result)
}

//export sub_510D10
func sub_510D10(owner *nox_object_t, ptr unsafe.Pointer, typ C.uint16_t, count C.uint8_t) {
	rt := game4TradeRuntime()
	rt.world.SellShopItemsByTypeNative510D10(asObjectS(owner), (*server.TradeSession)(ptr), uint16(typ), uint8(count), rt.sell())
}

//export nox_game4_shopRepair_510AE0
func nox_game4_shopRepair_510AE0(owner *nox_object_t, ptr unsafe.Pointer, code C.uint16_t) C.uintptr_t {
	rt := game4TradeRuntime()
	unit := asObjectS(owner)
	result := uintptr(game4GoldGet(unit))
	if unit.InvFirstItem != nil {
		result = uintptr(uint16(code))
	}
	rt.world.RepairShopItemNative510AE0(unit, (*server.TradeSession)(ptr), uint16(code), rt.repair())
	runtime.KeepAlive(unit)
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
