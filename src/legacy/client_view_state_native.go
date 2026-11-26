package legacy

/*
#include "GAME1_2.h"
#include "GAME2_2.h"

extern uint32_t dword_5d4594_805820, nox_xxx_useAudio_587000_80772;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_430AA0
func sub_430AA0(mode C.int) C.int {
	switch mode {
	case 1:
		C.dword_5d4594_805820, C.nox_xxx_useAudio_587000_80772 = 1, 9
		return 0
	case 2:
		C.dword_5d4594_805820, C.nox_xxx_useAudio_587000_80772 = 2, 13
		return 0
	default:
		C.dword_5d4594_805820, C.nox_xxx_useAudio_587000_80772 = 0, 5
		return mode - 2
	}
}

//export nox_client_setMousePos_430B10
func nox_client_setMousePos_430B10(x, y C.int) {
	nox_client_changeMousePos_430A00(int32(x), int32(y), true)
}

//export sub_430B50
func sub_430B50(left, top, right, bottom C.int) C.int {
	legacyWords.dword_5d4594_3807140, legacyWords.dword_5d4594_3807136 = C.uint32_t(left), C.uint32_t(top)
	legacyWords.dword_5d4594_3807116, legacyWords.dword_5d4594_3807152 = C.uint32_t(right), C.uint32_t(bottom)
	return bottom
}

//export sub_431370
func sub_431370() C.int {
	return C.int(bool2int(sub_488B60() != 0))
}

//export sub_431380
func sub_431380() {
	sub_488BA0()
	sub_4896E0()
}

//export nox_xxx_getHostInfoPtr_431770
func nox_xxx_getHostInfoPtr_431770() *C.char {
	return (*C.char)(memmap.PtrOff(0x5D4594, 807172))
}

//export nox_xxx_initTime_435570
func nox_xxx_initTime_435570() C.longlong {
	ticks := uint64(nox_platform_get_ticks())
	*memmap.PtrUint64(0x5D4594, 811908) = ticks
	return C.longlong(ticks)
}

//export sub_435690
func sub_435690(dst *C.uint32_t) *C.uint32_t {
	if dst != nil {
		point := unsafe.Slice(dst, 2)
		point[0] = C.uint32_t(memmap.Uint32(0x5D4594, 811364))
		point[1] = C.uint32_t(memmap.Uint32(0x5D4594, 811368))
	}
	return dst
}

//export sub_435700
func sub_435700(text *wchar2_t, style C.int) *wchar2_t {
	buffer := memmap.PtrT[[264]uint16](0x5D4594, 811376)[:]
	n := 0
	if text != nil {
		for n < len(buffer)-1 {
			unit := *(*uint16)(unsafe.Add(unsafe.Pointer(text), 2*n))
			if unit == 0 {
				break
			}
			buffer[n] = unit
			n++
		}
	}
	buffer[n] = 0
	*memmap.PtrInt32(0x5D4594, 811060) = int32(style)
	return (*wchar2_t)(unsafe.Pointer(&buffer[0]))
}
