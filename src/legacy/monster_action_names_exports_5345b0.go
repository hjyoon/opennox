package legacy

/*
#include "GAME4_3.h"
#include "cgo_const.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//export sub_5345B0
func sub_5345B0(action C.int) *C.char {
	return (*C.char)(unsafe.Pointer(alloc.InternCString(monsterSelectedActionNameNative5345B0(uint32(action)))))
}

//export sub_534650
func sub_534650(action C.int) *C.char {
	name, ok := monsterActionNameNative534650(uint32(action))
	if !ok {
		return nil
	}
	return (*C.char)(unsafe.Pointer(alloc.InternCString(name)))
}

//export nox_xxx_actionNByNameMB_5345F0
func nox_xxx_actionNByNameMB_5345F0(name *C.nox_cgo_const_char) C.int {
	return C.int(monsterActionByName528DB0(GoStringP(unsafe.Pointer(name)), 39, 38))
}

//export nox_xxx_actionByName_534670
func nox_xxx_actionByName_534670(name *C.nox_cgo_const_char) C.int {
	return C.int(monsterActionByName528DB0(GoStringP(unsafe.Pointer(name)), 72, 0))
}
