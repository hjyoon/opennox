package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func makeScorchNative537AF0(pos types.Pointf, kind int) {
	srv := GetServer()
	srv.S().MakeScorch537AF0(pos, kind, server.MakeScorchRuntime537AF0{
		CreateObjectAt: func(obj, owner *server.Object, pos types.Pointf) {
			srv.CreateObjectAt(obj, owner, pos)
		},
	})
}

// Keep this indirection dynamic so tests cover both Go and legacy C entries
// without constructing the global game server.
var makeScorchCall537AF0 = makeScorchNative537AF0

func Nox_xxx_sMakeScorch_537AF0(pos types.Pointf, kind int) {
	makeScorchCall537AF0(pos, kind)
}

//export nox_xxx_sMakeScorch_native_537AF0
func nox_xxx_sMakeScorch_native_537AF0(pos unsafe.Pointer, kind C.int) {
	if pos == nil {
		return
	}
	makeScorchCall537AF0(*(*types.Pointf)(pos), int(kind))
}

func makeScorchLegacyEntry537AF0(pos types.Pointf, kind int) {
	C.nox_xxx_sMakeScorch_537AF0((*C.float)(unsafe.Pointer(&pos)), C.int(kind))
}
