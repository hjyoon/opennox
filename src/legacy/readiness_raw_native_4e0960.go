package legacy

/*
#include "readiness_raw_4e0960.h"
*/
import "C"

import "unsafe"

func init() {
	readinessRawEngine4E0960 = readinessRawNative4E0960
}

//export nox_xxx_itemCheckReadinessEffect_4E0960
func nox_xxx_itemCheckReadinessEffect_4E0960(item *C.nox_object_t) C.int32_t {
	return C.int32_t(readinessRawNative4E0960(unsafe.Pointer(item)))
}

func readinessRawNative4E0960(item unsafe.Pointer) int32 {
	// 004E0965 precedes the original item-nil/class admission. The table is
	// cached even when an after-read stimulus changes item.InitData.
	table := readinessRawRead4E0960(item, 0, 0).Pointer
	if item == nil || readinessRawRead4E0960(item, 1, 0).Word&0x13001000 == 0 {
		return 0
	}
	for slot := 2; slot < 4; slot++ {
		effect := readinessRawRead4E0960(table, 2, slot).Pointer
		if effect == nil {
			continue
		}
		callback := readinessRawRead4E0960(effect, 3, 0).Pointer
		if callback != ReadinessEffectPointer4E0960() {
			continue
		}
		// 004E099C reloads the matched slot from the cached table. Its new
		// descriptor is used without another callback or nil check.
		effect = readinessRawRead4E0960(table, 2, slot).Pointer
		return int32(readinessRawRead4E0960(effect, 4, 0).Word)
	}
	return 0
}
