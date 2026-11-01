package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

var basicObjectSendRuntimeNative5187E0 = func() server.BasicObjectSendRuntime5187E0 { return GetServer().BasicObjectSendRuntimeNative5187E0() }

func NetSendPhantomObject5187E0(index int32, unit *server.Object) int32 {
	return server.PhantomObjectSendNative5187E0(index, unit, basicObjectSendRuntimeNative5187E0())
}
func NetSendSimpleObject5188A0(index int32, unit *server.Object) int32 {
	return server.SimpleObjectSendNative5188A0(index, unit, basicObjectSendRuntimeNative5187E0())
}

//export nox_xxx_netSendPhantomPlrMb_5187E0
func nox_xxx_netSendPhantomPlrMb_5187E0(index C.int, unit *C.nox_object_t) C.int {
	return C.int(NetSendPhantomObject5187E0(int32(index), asObjectS(unit)))
}

//export nox_xxx_netSendSimpleObj_5188A0
func nox_xxx_netSendSimpleObj_5188A0(index C.int, unit *C.nox_object_t) C.int {
	return C.int(NetSendSimpleObject5188A0(int32(index), asObjectS(unit)))
}

func Nox_xxx_netSendPhantomPlrMb_5187E0(index int, unit *server.Object) int {
	return int(C.nox_xxx_netSendPhantomPlrMb_5187E0(C.int(index), asObjectC(unit)))
}
func Nox_xxx_netSendSimpleObj_5188A0(index int, unit *server.Object) int {
	return int(C.nox_xxx_netSendSimpleObj_5188A0(C.int(index), asObjectC(unit)))
}
