package legacy

/*
// The shared header still contains unrelated Win32-only layout assertions.
#define _Static_assert(...)
#include "GAME4_1.h"
#undef _Static_assert
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var minimapMonsterIterator50AAE0 server.MinimapMonsterIterator50AAE0

var minimapMonsterFirstImpl50AAE0 = func() *server.Object {
	return minimapMonsterIterator50AAE0.First(GetServer().S().Objs.First())
}

var minimapMonsterNextImpl50AB10 = func() *server.Object {
	return minimapMonsterIterator50AAE0.Next()
}

func minimapMonsterFirstExportCall50AAE0() unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_minimapFirstMonster_50AAE0())
}

func minimapMonsterNextExportCall50AB10() unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_minimapNextMonster_50AB10())
}

//export nox_minimapFirstMonsterObject_50AAE0
func nox_minimapFirstMonsterObject_50AAE0() *nox_object_t {
	return asObjectC(minimapMonsterFirstImpl50AAE0())
}

//export nox_minimapNextMonsterObject_50AB10
func nox_minimapNextMonsterObject_50AB10() *nox_object_t {
	return asObjectC(minimapMonsterNextImpl50AB10())
}
