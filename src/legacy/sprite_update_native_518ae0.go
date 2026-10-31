package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
#include "GAME5_2.h"
#include "GAME3_3.h"
#include "common__net_list.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

var spriteUpdateRuntimeNative518AE0 = func() server.SpriteUpdateNativeRuntime518AE0 {
	return server.SpriteUpdateNativeRuntime518AE0{
		UnitCode: func(unit *server.Object) uint32 {
			return uint32(C.nox_xxx_netGetUnitCodeServ_578AC0(asObjectC(unit)))
		},
		ReadType: func(index int) uint32 {
			return *memmap.PtrUint32(0x5d4594, uintptr(2386972+4*index))
		},
		ElevatorLink: (*server.Object).ElevatorLinkFor,
		Important: func(player int32, packet [4]byte, _ int32, reliable int32) int32 {
			return int32(C.nox_xxx_netSendPacket1_4E5390(C.int(player), unsafe.Pointer(&packet[0]), C.int(len(packet)), nil, C.int(reliable)))
		},
		Queue: func(player, kind int32, packet [4]byte) int32 {
			return int32(C.nox_netlist_addToMsgListCli_40EBC0(C.int(player), C.int(kind), (*C.uchar)(unsafe.Pointer(&packet[0])), C.int(len(packet))))
		},
		Special: func(recipient, unit *server.Object) int32 {
			return int32(bool2int(GetServer().NetUpdateObjectSpecialNative527E50(recipient, unit)))
		},
	}
}

func NetSpriteUpdate518AE0(recipient *server.Object, player int32, unit *server.Object) int32 {
	return server.SpriteUpdateNative518AE0(recipient, player, unit, spriteUpdateRuntimeNative518AE0())
}

//export nox_xxx_netSpriteUpdate_518AE0
func nox_xxx_netSpriteUpdate_518AE0(recipient *C.nox_object_t, player C.int, unit *C.uint32_t) C.int {
	return C.int(NetSpriteUpdate518AE0(asObjectS(recipient), int32(player), (*server.Object)(unsafe.Pointer(unit))))
}
