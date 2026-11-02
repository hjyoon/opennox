package legacy

/*
#include "GAME3_2.h"
_Static_assert(_Generic(&nox_xxx_wallSendDestroyed_4DF0A0,int (*)(nox_map_wall_504290*,int):1,default:0), "native wall-destroyed pointer argument and scalar C-int result ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_wallSendDestroyed_4DF0A0
func nox_xxx_wallSendDestroyed_4DF0A0(wall *C.nox_map_wall_504290, recipient C.int) C.int {
	return C.int((*server.Wall)(unsafe.Pointer(wall)).WallSendDestroyedNative4DF0A0(int32(recipient), secretWallPacketRuntime4DF120()))
}
