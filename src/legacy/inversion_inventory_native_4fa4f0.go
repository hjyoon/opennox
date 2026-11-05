package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME4.h"
*/
import "C"

import "unsafe"

//export nox_xxx_checkInversionEffect_4FA4F0
func nox_xxx_checkInversionEffect_4FA4F0(root, target *C.nox_object_t) C.int32_t {
	base, other := unsafe.Pointer(root), unsafe.Pointer(target)
	for item := inversionInventoryRead4FA4F0(base, 6, 0).Pointer; item != nil; item = inversionInventoryRead4FA4F0(item, 7, 0).Pointer {
		if inversionInventoryRead4FA4F0(item, 9, 0).Word&0x100 == 0 {
			continue
		}
		if inversionInventoryRead4FA4F0(item, 1, 0).Word&0x13001000 == 0 {
			continue
		}
		data := inversionInventoryRead4FA4F0(item, 2, 0).Pointer
		for slot := 2; slot < 4; slot++ {
			effect := inversionInventoryRead4FA4F0(data, 3, slot).Pointer
			var output int32
			if effect == nil {
				continue
			}
			function := inversionInventoryRead4FA4F0(effect, 4, 0).Pointer
			if function != modifierPredicatesCallback4E03D0(0) {
				continue
			}
			// Original 4FA54C calls the real parent service even for weak or
			// negative effects. 4FA561 then reloads the actual callback.
			parent := inversionInventoryParentRuntime4FA4F0(other)
			function = inversionInventoryRead4FA4F0(effect, 4, 0).Pointer
			inversionInventoryInvoke4FA4F0(function, effect, [4]uintptr{uintptr(item), uintptr(base), uintptr(other), uintptr(parent)}, &output)
			if output == 1 {
				return 1
			}
		}
	}
	return 0
}
