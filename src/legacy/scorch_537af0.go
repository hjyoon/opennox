package legacy

/*
#include "GAME4_3.h"

typedef void (*scorch_entry)(float*, int);
typedef int (*scorch_init_entry)(void);
_Static_assert(_Generic(&nox_xxx_sMakeScorch_537AF0, scorch_entry: 1, default: 0), "scorch ABI");
_Static_assert(_Generic(&nox_xxx_scorchInit_537BD0, scorch_init_entry: 1, default: 0), "scorch init ABI");

static void call_scorch_indirect(float* point, int kind) {
	scorch_entry volatile entry = nox_xxx_sMakeScorch_537AF0;
	entry(point, kind);
}
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

//export nox_xxx_sMakeScorch_537AF0
func nox_xxx_sMakeScorch_537AF0(pos *C.float, kind C.int) {
	if pos == nil {
		return
	}
	makeScorchCall537AF0(*(*types.Pointf)(unsafe.Pointer(pos)), int(kind))
}

//export nox_xxx_scorchInit_537BD0
func nox_xxx_scorchInit_537BD0() C.int {
	return C.int(GetServer().S().ScorchInit537BD0())
}

func makeScorchLegacyEntry537AF0(pos *types.Pointf, kind int, indirect bool) {
	p := (*C.float)(unsafe.Pointer(pos))
	if indirect {
		C.call_scorch_indirect(p, C.int(kind))
	} else {
		C.nox_xxx_sMakeScorch_537AF0(p, C.int(kind))
	}
}

func scorchInitLegacyEntry537BD0() int {
	return int(C.nox_xxx_scorchInit_537BD0())
}
