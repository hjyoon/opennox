package legacy

/*
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "map_temp_lists_503f40.h"
_Static_assert(_Generic(&sub_4DF120,int (*)(nox_map_wall_504290*):1,default:0), "native close-wall public pointer ABI");
_Static_assert(_Generic(&sub_4DF180,int (*)(nox_map_wall_504290*):1,default:0), "native open-wall public pointer ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func secretWallPacketRuntime4DF120() server.SecretWallPacketRuntime4DF120 {
	if secretWallNativeRuntimeOverride4DF120 != nil {
		return *secretWallNativeRuntimeOverride4DF120
	}
	return server.SecretWallPacketRuntime4DF120{
		FirstUnit: func() *server.Object {
			return (*server.Object)(unsafe.Pointer(C.nox_xxx_getFirstPlayerUnit_4DA7C0()))
		},
		NextUnit: func(unit *server.Object) *server.Object {
			return (*server.Object)(unsafe.Pointer(C.nox_xxx_getNextPlayerUnit_4DA7F0((*C.nox_object_t)(unsafe.Pointer(unit)))))
		},
		Send: func(recipient int, packet []byte, related *server.Object, remove int) int {
			return int(C.nox_xxx_netSendPacket0_4E5420(C.int(recipient), unsafe.Pointer(unsafe.SliceData(packet)), C.int(len(packet)), (*C.nox_object_t)(unsafe.Pointer(related)), C.int(remove)))
		},
	}
}

//export sub_4DF120
func sub_4DF120(wall *C.nox_map_wall_504290) C.int {
	return C.int((*server.Wall)(unsafe.Pointer(wall)).SecretWallStateNative4DF120(false, secretWallPacketRuntime4DF120()))
}

//export sub_4DF180
func sub_4DF180(wall *C.nox_map_wall_504290) C.int {
	return C.int((*server.Wall)(unsafe.Pointer(wall)).SecretWallStateNative4DF120(true, secretWallPacketRuntime4DF120()))
}
