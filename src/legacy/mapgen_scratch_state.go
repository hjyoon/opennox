package legacy

/*
#include "GAME3_2.h"
#include "cgo_const.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const randomMapNameOffset = 197860

func mapgenTileInitState() *[72]byte {
	return memmap.PtrT[[72]byte](mapgenTileStateBase, mapgenTileIDOffset)
}

//export nox_xxx_tileInitdataClear_4D3C50
func nox_xxx_tileInitdataClear_4D3C50(state *C.nox_cgo_const_void) {
	*mapgenTileInitState() = *(*[72]byte)(unsafe.Pointer(state))
}

//export sub_4D3C70
func sub_4D3C70() *C.uchar {
	return (*C.uchar)(unsafe.Pointer(mapgenTileInitState()))
}

func Sub_4D42E0(name string) {
	sub_4D42E0((*C.nox_cgo_const_char)(internCStr(name)))
}

//export sub_4D42E0
func sub_4D42E0(name *C.nox_cgo_const_char) C.uint {
	src := (*byte)(unsafe.Pointer(name))
	size := C.uint(alloc.StrLen(src) + 1)
	copy(memmap.Slice(0x587000, randomMapNameOffset)[:int(size)], unsafe.Slice(src, int(size)))
	return size
}

//export nox_xxx_getRandMapName_4D4310
func nox_xxx_getRandMapName_4D4310() *C.char {
	return (*C.char)(memmap.PtrOff(0x587000, randomMapNameOffset))
}
