package legacy

/*
#include <stdint.h>

#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func useReadLegacy53F7C0(owner, readable *server.Object) int32 {
	return GetServer().S().ReadUse53F7C0(owner, readable)
}

func readUseExportCall53F7C0(owner, readable *server.Object) int32 {
	return int32(C.nox_xxx_useRead_53F7C0(
		asObjectC(owner),
		asObjectC(readable),
	))
}

//export nox_xxx_useRead_53F7C0
func nox_xxx_useRead_53F7C0(
	owner, readable *C.nox_object_t,
) C.int32_t {
	return C.int32_t(useReadLegacy53F7C0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(readable)),
	))
}
