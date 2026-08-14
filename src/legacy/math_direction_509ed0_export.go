package legacy

/*
// The shared header still contains unrelated Win32-only layout assertions.
// This boundary has its own multi-architecture fixture for float2 and the
// exported signature, so suppress those unrelated assertions while cgo parses
// the canonical declaration.
#define _Static_assert(...)
#include "GAME4_1.h"
#undef _Static_assert
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_math_509ED0
func nox_xxx_math_509ED0(vector *C.float2) C.int {
	point := *(*types.Pointf)(unsafe.Pointer(vector))
	return C.int(server.DirFromVec(point))
}
