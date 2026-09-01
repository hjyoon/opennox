package legacy

/*
#include <stdint.h>

#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func wandUseExportCall53F290(owner, wand *server.Object) int32 {
	return int32(C.nox_xxx_useLesserFireballStaff_53F290(
		asObjectC(owner),
		asObjectC(wand),
	))
}

func wandCastUseExportCall53F4F0(owner, wand *server.Object) int32 {
	return int32(C.nox_xxx_useWandCastSpell_53F4F0(
		asObjectC(owner),
		asObjectC(wand),
	))
}

func fireWandUseExportCall53F670(owner, wand *server.Object) int32 {
	return int32(C.nox_xxx_useFireWand_53F670(
		asObjectC(owner),
		asObjectC(wand),
	))
}

//export nox_xxx_useLesserFireballStaff_53F290
func nox_xxx_useLesserFireballStaff_53F290(
	owner, wand *C.nox_object_t,
) C.int32_t {
	return C.int32_t(bool2int(Nox_xxx_useWand_53F290(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(wand)),
	)))
}

//export nox_xxx_useWandCastSpell_53F4F0
func nox_xxx_useWandCastSpell_53F4F0(
	owner, wand *C.nox_object_t,
) C.int32_t {
	return C.int32_t(bool2int(Nox_xxx_useWandCast_53F4F0(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(wand)),
	)))
}

//export nox_xxx_useFireWand_53F670
func nox_xxx_useFireWand_53F670(
	owner, wand *C.nox_object_t,
) C.int32_t {
	return C.int32_t(bool2int(Nox_xxx_useFireWand_53F670(
		asObjectS((*nox_object_t)(owner)),
		asObjectS((*nox_object_t)(wand)),
	)))
}

func Get_nox_xxx_useLesserFireballStaff_53F290() unsafe.Pointer {
	return C.nox_xxx_useLesserFireballStaff_53F290
}

func Get_nox_xxx_useWandCastSpell_53F4F0() unsafe.Pointer {
	return C.nox_xxx_useWandCastSpell_53F4F0
}

func Get_nox_xxx_useFireWand_53F670() unsafe.Pointer {
	return C.nox_xxx_useFireWand_53F670
}
