package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME3_2.h"
*/
import "C"

import "unsafe"

func modifierPredicateNativeScanCall4E03D0(first, second uintptr, item unsafe.Pointer, context uintptr) int32 {
	return int32(C.nox_xxx_unusedCheckGripEffect_4E03F0(C.uintptr_t(first), C.uintptr_t(second), (*C.nox_object_t)(item), C.uintptr_t(context)))
}
