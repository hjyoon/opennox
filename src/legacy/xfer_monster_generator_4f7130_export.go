package legacy

/*
#include "xfer_monster_generator_4f7130.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

var monsterGeneratorXferCall4F7130 = func(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return Nox_xxx_XFerMonsterGenNative4F7130(cf, object)
}

func monsterGeneratorXferExportCall4F7130(object *server.Object) int32 {
	return int32(C.nox_xxx_XFerMonsterGen_4F7130(asObjectC(object)))
}

//export nox_xxx_XFerMonsterGen_4F7130
func nox_xxx_XFerMonsterGen_4F7130(object *C.nox_object_t) C.int32_t {
	return C.int32_t(monsterGeneratorXferCall4F7130(
		cryptfile.Global(),
		asObjectS((*nox_object_t)(unsafe.Pointer(object))),
	))
}
