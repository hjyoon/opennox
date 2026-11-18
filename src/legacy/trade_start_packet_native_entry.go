package legacy

/*
#include "server__system__trade.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_servShopStart_50EF10_trade
func nox_xxx_servShopStart_50EF10_trade(starter, target *nox_object_t) unsafe.Pointer {
	return unsafe.Pointer(game4TradeRuntime().start(asObjectS(starter), asObjectS(target)))
}

//export sub_50F0F0
func sub_50F0F0(_ *nox_object_t, ptr unsafe.Pointer) C.int {
	return C.int(game4TradeRuntime().world.SendShopStartNative50F0F0((*server.TradeSession)(ptr)))
}

//export sub_50F1A0
func sub_50F1A0(recipient *nox_object_t, ptr unsafe.Pointer) C.int {
	return C.int(game4TradeRuntime().world.SendTradeStartNative50F1A0(asObjectS(recipient), (*server.TradeSession)(ptr)))
}
