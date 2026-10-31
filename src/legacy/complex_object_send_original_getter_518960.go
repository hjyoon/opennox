package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
#include "GAME5_2.h"
#include "GAME1_1.h"
#include "common__random.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include <stdlib.h>
_Static_assert(_Generic(&nox_xxx_netSendComplexObject_518960, int (*)(int,uint32_t*,int):1,default:0), "complex sender int/native pointer/int ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func complexObjectSendPublicC518960(player int32, unit *server.Object, report int32) int32 {
	return int32(C.nox_xxx_netSendComplexObject_518960(C.int(player), (*C.uint32_t)(unsafe.Pointer(unit)), C.int(report)))
}

func complexObjectUnitCodeC518960(unit *server.Object) uint32 {
	return uint32(C.nox_xxx_netGetUnitCodeServ_578AC0(asObjectC(unit)))
}

func complexObjectFloatToIntC518960(value float32) int32 {
	return int32(C.nox_float2int(C.float(value)))
}

func complexObjectRandomC518960(min, max int32) int32 {
	return int32(C.nox_common_randomInt_415FA0(C.int(min), C.int(max)))
}

func complexObjectReportHealthC518960(player int32, code uint16, delta int16) {
	C.nox_xxx_netReportHealthDelta_4D8760(C.int(player), C.short(code), C.short(delta))
}

func complexObjectDirectionC518960(direction int32, out *server.IndexedDirectionVector509E20) {
	var v C.int2
	C.nox_xxx_xferIndexedDirection_509E20(C.int(direction), &v)
	out.X, out.Y = int32(v.field_0), int32(v.field_4)
}

func complexObjectTakeImportantPacketC518960() (int32, []byte) {
	packet := C.nox_server_getImportantFirst_4E4F80()
	if packet == nil {
		return 0, nil
	}
	player := int32(packet.recipient)
	data := C.GoBytes(unsafe.Pointer(&packet.payload[0]), C.int(packet.payload_size))
	C.sub_4E4FC0(packet)
	return player, data
}

func complexObjectWordIDC518960(name string) uint32 {
	p := C.CString(name)
	defer C.free(unsafe.Pointer(p))
	return uint32(C.nox_xxx_getNameId_4E3AA0(p))
}
