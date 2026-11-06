package legacy

/*
#include "regeneration_raw_4e01d0.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func init() {
	regenerationRawEngine4E01D0 = regenerationRawNative4E01D0
}

//export nox_xxx_effectRegeneration_4E01D0
func nox_xxx_effectRegeneration_4E01D0(effect unsafe.Pointer, item *C.nox_object_t) {
	regenerationRawNative4E01D0(effect, (*server.Object)(unsafe.Pointer(item)))
}

func regenerationRawNative4E01D0(effect unsafe.Pointer, item *server.Object) {
	if item == nil {
		return
	}
	owner := (*server.Object)(regenerationRawRead4E01D0(unsafe.Pointer(item), 0).Pointer)
	if owner == nil || regenerationRawRead4E01D0(unsafe.Pointer(owner), 2).Pointer == nil {
		return
	}
	frame := regenerationRawRead4E01D0(nil, 8).Word
	last := regenerationRawRead4E01D0(unsafe.Pointer(owner), 3).Word
	fps := regenerationRawRead4E01D0(nil, 9).Word
	if frame-last < fps || regenerationRawRead4E01D0(unsafe.Pointer(owner), 4).Word&0x8020 != 0 {
		return
	}
	maximum := regenerationRawMaximum4E01D0(owner)
	if regenerationRawCurrent4E01D0(owner) >= maximum {
		return
	}
	class := regenerationRawRead4E01D0(unsafe.Pointer(item), 1).Word
	rate := regenerationRawRead4E01D0(effect, 5).Word
	if class&0x02000000 != 0 && regenerationRawArmor4E01D0(item)&0x4000 != 0 {
		rate /= 3
	}
	maximum = regenerationRawMaximum4E01D0(owner)
	fps = regenerationRawRead4E01D0(nil, 9).Word
	interval := regenerationRawDivide4E01D0(0, rate*fps, uint32(maximum))
	frame = regenerationRawRead4E01D0(nil, 8).Word
	if regenerationRawDivide4E01D0(1, frame, interval) == 0 {
		regenerationRawAdjust4E01D0(owner, 1)
	}
}
