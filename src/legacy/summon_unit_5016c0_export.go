package legacy

/*
#include "summon_unit_5016c0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func summonUnitExportCall5016C0(
	typeID int32,
	position *types.Pointf,
	owner *server.Object,
	direction uint8,
) *server.Object {
	result := C.nox_xxx_unitDoSummonAt_5016C0(
		C.int32_t(typeID),
		(*C.float)(unsafe.Pointer(position)),
		asObjectC(owner),
		C.uint8_t(direction),
	)
	return asObjectS((*nox_object_t)(result))
}

//export nox_xxx_unitDoSummonAt_5016C0
func nox_xxx_unitDoSummonAt_5016C0(
	typeID C.int32_t,
	position *C.float,
	owner *C.nox_object_t,
	direction C.uint8_t,
) *C.nox_object_t {
	created := Nox_xxx_unitDoSummonAt_5016C0(
		int32(typeID),
		(*types.Pointf)(unsafe.Pointer(position)),
		asObjectS((*nox_object_t)(owner)),
		uint8(direction),
	)
	return (*C.nox_object_t)(asObjectC(created))
}
