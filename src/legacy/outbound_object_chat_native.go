package legacy

/*
#include "GAME4_2.h"
#include "cgo_const.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/opennox/v1/server"
)

var outboundObjectChatRuntime = func() server.OutboundObjectChatRuntime {
	return GetServer().S().OutboundObjectChatRuntime()
}

//export nox_xxx_netObjectOutOfSight_528A60
func nox_xxx_netObjectOutOfSight_528A60(recipient C.int, unit *nox_object_t) C.int {
	return C.int(server.NetObjectVisibility528A60(int(recipient), asObjectS(unit), netmsg.MSG_OBJECT_OUT_OF_SIGHT, outboundObjectChatRuntime()))
}

//export nox_xxx_netObjectInShadows_528A90
func nox_xxx_netObjectInShadows_528A90(recipient C.int, unit *nox_object_t) C.int {
	return C.int(server.NetObjectVisibility528A60(int(recipient), asObjectS(unit), netmsg.MSG_OBJECT_IN_SHADOWS, outboundObjectChatRuntime()))
}

//export nox_xxx_monsterCmdSend_528BD0
func nox_xxx_monsterCmdSend_528BD0(unit, source *nox_object_t, command *C.nox_cgo_const_char, duration C.short) C.uintptr_t {
	return C.uintptr_t(server.MonsterCommandSendRaw528BD0(asObjectS(unit), asObjectS(source), GoStringP(unsafe.Pointer(command)), int16(duration), outboundObjectChatRuntime()))
}

//export nox_xxx_destroyEveryChatMB_528D60
func nox_xxx_destroyEveryChatMB_528D60() C.int {
	return C.int(server.DestroyEveryChatRaw528D60(outboundObjectChatRuntime()))
}
