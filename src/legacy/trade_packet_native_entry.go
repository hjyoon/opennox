package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export sub_50F720
func sub_50F720(recipient *nox_object_t, session unsafe.Pointer) C.int {
	return C.int(tradePacketStatus50F720(asObjectS(recipient), (*server.TradeSession)(session)))
}

//export nox_xxx_tradeP2PUpdStuff_50FA00
func nox_xxx_tradeP2PUpdStuff_50FA00(recipient *nox_object_t, session unsafe.Pointer) C.int {
	return C.int(tradePacketBalance50FA00(asObjectS(recipient), (*server.TradeSession)(session)))
}

//export sub_50FF90
func sub_50FF90(recipient *nox_object_t, unusedSession unsafe.Pointer, item *nox_object_t) C.int {
	return C.int(tradePacketRemove50FF90(asObjectS(recipient), unusedSession, asObjectS(item)))
}
