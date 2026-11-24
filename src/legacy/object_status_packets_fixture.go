package legacy

/*
#include "GAME3_2.h"
#include "GAME4_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func objectStatusPacketNativeCalls(player uint8, object *server.Object) []func() int32 {
	recipient := C.int(player)
	unit := asObjectC(object)
	words := (*C.uint32_t)(unsafe.Pointer(unit))
	spell := uint32(0x89abcdef)
	return []func() int32{
		func() int32 { return int32(C.nox_xxx_netSendFlagStatus_4D95A0(recipient, -127, 127, 0x35, -21555)) },
		func() int32 { return int32(C.nox_xxx_netSendBallStatus_4D95F0(recipient, -2, 0x7012)) },
		func() int32 { return int32(C.nox_xxx_netReportSpellStat_4D9630(recipient, C.int(int32(spell)), 127)) },
		func() int32 { return int32(C.nox_xxx_netSendSecondaryWeapon_4D9670(recipient, words, -128)) },
		func() int32 { return int32(C.nox_xxx_netMsgLastQuiver_4D96B0(recipient, words)) },
		func() int32 { return int32(C.nox_xxx_netMsgInventoryLoaded_4D96E0(recipient)) },
		func() int32 { return int32(C.nox_xxx_netFriendAddRemove_4D97A0(recipient, unit, 1)) },
		func() int32 { return int32(C.nox_xxx_netFriendAddRemove_4D97A0(recipient, unit, -1)) },
		func() int32 { return int32(C.sub_4D97E0(recipient)) },
		func() int32 { return int32(C.nox_xxx_netMsgFadeBeginPlayer(recipient, -2, 0)) },
	}
}

func objectStatusDestroyNativeCall(object *server.Object) {
	C.nox_xxx_netReportDestroyObject_5289D0(asObjectC(object))
}
