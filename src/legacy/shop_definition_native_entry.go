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

//export sub_510320
func sub_510320(item *nox_object_t, session unsafe.Pointer) {
	game4TradeRuntime().world.DecrementShopDefinitionNative510320((*server.TradeSession)(session), asObjectS(item))
}

//export nox_xxx_getSomeShopData_5103A0
func nox_xxx_getSomeShopData_5103A0(session unsafe.Pointer, item *nox_object_t) C.int {
	return C.int(game4TradeRuntime().world.IndexShopDefinitionNative5103A0((*server.TradeSession)(session), asObjectS(item)))
}

//export sub_5103F0
func sub_5103F0(item *nox_object_t, definition unsafe.Pointer) C.int {
	world := game4TradeRuntime().world
	return C.int(bool2int(world.ShopDefinitionMatchesItemNative5103F0(asObjectS(item), (*server.ShopkeeperItemDefinition)(definition))))
}
