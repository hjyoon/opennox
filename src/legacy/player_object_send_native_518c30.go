package legacy

/*
#include "defs.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_1.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var playerObjectSendRuntimeNative518C30 = func() server.PlayerObjectSendRuntime518C30 {
	return GetServer().PlayerObjectSendRuntimeNative518C30()
}

func NetPlayerObjectSend518C30(recipient, unit *server.Object, stream, health int32) int32 {
	return server.PlayerObjectSendNative518C30(recipient, unit, stream, health, playerObjectSendRuntimeNative518C30())
}

//export nox_xxx_netPlayerObjSend_518C30
func nox_xxx_netPlayerObjSend_518C30(recipient, unit *C.nox_object_t, stream, health C.int) C.int {
	return C.int(NetPlayerObjectSend518C30(asObjectS(recipient), asObjectS(unit), int32(stream), int32(health)))
}

func PlayerSendHealthDeltaNative4D8760(index int32, code uint16, delta int16) {
	C.nox_xxx_netReportHealthDelta_4D8760(C.int(index), C.short(code), C.short(delta))
}

func PlayerSendFloatToIntNative419A70(value float32) int32 {
	return int32(C.nox_float2int(C.float(value)))
}

func PlayerSendDirectionNative509E20(index int32) server.IndexedDirectionVector509E20 {
	var value C.int2
	C.nox_xxx_xferIndexedDirection_509E20(C.int(index), &value)
	return server.IndexedDirectionVector509E20{X: int32(value.field_0), Y: int32(value.field_4)}
}
