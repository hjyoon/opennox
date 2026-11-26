package legacy

/*
#include "GAME4_1.h"
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

//export nox_xxx_minimapFirstMonster_50AAE0
func nox_xxx_minimapFirstMonster_50AAE0() *C.float {
	if obj := nox_minimapFirstMonsterObject_50AAE0(); obj != nil {
		return &obj.x
	}
	return nil
}

//export nox_xxx_minimapNextMonster_50AB10
func nox_xxx_minimapNextMonster_50AB10() *C.float {
	if obj := nox_minimapNextMonsterObject_50AB10(); obj != nil {
		return &obj.x
	}
	return nil
}

//export nox_minimapFirstMonsterObject_50AAE0
func nox_minimapFirstMonsterObject_50AAE0() *nox_object_t {
	return asObjectC(minimapMonsterFirstImpl50AAE0())
}

//export nox_minimapNextMonsterObject_50AB10
func nox_minimapNextMonsterObject_50AB10() *nox_object_t {
	return asObjectC(minimapMonsterNextImpl50AB10())
}
