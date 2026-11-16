package legacy

/*
#include "defs.h"
*/
import "C"

import "unsafe"

//export nox_xxx_sendEndTradeMsg_50F560
func nox_xxx_sendEndTradeMsg_50F560(recipient *nox_object_t) C.int {
	return C.int(tradeNotifyEnd50F560(asObjectS(recipient)))
}

//export sub_50F6E0
func sub_50F6E0(recipient *nox_object_t) C.int {
	return C.int(tradeNotifyContinue50F6E0(asObjectS(recipient)))
}

//export sub_50FAE0
func sub_50FAE0(recipient, offering *nox_object_t, unusedSession unsafe.Pointer, item *nox_object_t, cost C.int) C.int {
	return C.int(tradeNotifyItem50FAE0(asObjectS(recipient), asObjectS(offering), unusedSession, asObjectS(item), uint32(cost)))
}
