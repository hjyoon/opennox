package legacy

/*
#include <stdint.h>
#include "defs.h"
#include "GAME3_2.h"
// Transport only: exercise the public scan with native-width opaque arguments.
static int32_t modifier_predicate_native_scan_call_4e03d0(uintptr_t first,uintptr_t second,nox_object_t* item,uintptr_t context) {
 return nox_xxx_unusedCheckGripEffect_4E03F0(first,second,item,context);
}
*/
import "C"

import "unsafe"

func modifierPredicateNativeScanCall4E03D0(first, second uintptr, item unsafe.Pointer, context uintptr) int32 {
	return int32(C.modifier_predicate_native_scan_call_4e03d0(C.uintptr_t(first), C.uintptr_t(second), (*C.nox_object_t)(item), C.uintptr_t(context)))
}
