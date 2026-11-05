package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME3_2.h"
*/
import "C"

import "unsafe"

// The original setge/setl predicates read signed strength before writing the
// caller's DWORD. Keep the result cached when output aliases the effect.
func modifierPredicateNative4E03D0(effect unsafe.Pointer, output *int32, inversion bool) int32 {
	strength := int32(modifierPredicateRead4E03D0(effect, 0, 0).Word)
	var result int32
	if (inversion && strength >= 1) || (!inversion && strength < 1) {
		result = 1
	}
	modifierPredicateStore4E03D0(output, uint32(result))
	return result
}

//export nox_xxx_inversionEffect_4E03D0
func nox_xxx_inversionEffect_4E03D0(effect unsafe.Pointer, item *C.nox_object_t, context C.uintptr_t, target *C.nox_object_t, other C.uintptr_t, output *C.int32_t) C.int32_t {
	return C.int32_t(modifierPredicateNative4E03D0(effect, (*int32)(unsafe.Pointer(output)), true))
}

//export nox_xxx_gripEffect_4E0480
func nox_xxx_gripEffect_4E0480(effect unsafe.Pointer, item *C.nox_object_t, context C.uintptr_t, target *C.nox_object_t, other C.uintptr_t, output *C.int32_t) C.int32_t {
	return C.int32_t(modifierPredicateNative4E03D0(effect, (*int32)(unsafe.Pointer(output)), false))
}

//export nox_xxx_unusedCheckGripEffect_4E03F0
func nox_xxx_unusedCheckGripEffect_4E03F0(first, second C.uintptr_t, item *C.nox_object_t, context C.uintptr_t) C.int32_t {
	base := unsafe.Pointer(item)
	if base == nil || modifierPredicateRead4E03D0(base, 1, 0).Word&0x03001000 == 0 {
		return 0
	}
	data := modifierPredicateRead4E03D0(base, 2, 0).Pointer
	output := int32(context)
	for slot := 2; slot < 4; slot++ {
		effect := modifierPredicateRead4E03D0(data, 3, slot).Pointer
		if effect == nil {
			continue
		}
		function := modifierPredicateRead4E03D0(effect, 4, 0).Pointer
		if function != modifierPredicatesCallback4E03D0(1) {
			continue
		}
		// The fourth argument is cached across calls, while its caller slot is
		// reused as output and can change after each actual Grip callback.
		modifierPredicateInvoke4E03D0(function, effect, [4]uintptr{uintptr(base), uintptr(second), uintptr(context), uintptr(first)}, &output)
		if output == 0 {
			return 1
		}
	}
	return 0
}
