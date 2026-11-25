package cfixture

/*
#include "GAME3_2.h"
*/
import "C"

import "unsafe"

func WriteMapgenTileStateC(state *[72]byte) {
	C.nox_xxx_tileInitdataClear_4D3C50(unsafe.Pointer(state))
}

func MapgenTileStateC() *[72]byte {
	return (*[72]byte)(unsafe.Pointer(C.sub_4D3C70()))
}

func WriteRandomMapNameC(name *byte) uint32 {
	return uint32(C.sub_4D42E0((*C.char)(unsafe.Pointer(name))))
}

func RandomMapNameC() *byte {
	return (*byte)(unsafe.Pointer(C.nox_xxx_getRandMapName_4D4310()))
}

func OrderMapgenCornersC(corners *[8]uint32) *[8]uint32 {
	return (*[8]uint32)(unsafe.Pointer(C.sub_4D3C80((*C.uint32_t)(unsafe.Pointer(corners)))))
}

func MapgenOrientationDirectionC(direction int32) int32 {
	return int32(C.sub_4D3FF0(C.int(direction)))
}
