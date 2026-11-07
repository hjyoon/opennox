package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import "unsafe"

// Bound to the same implementation as the public C entry. The raw entry
// preserves native pointers and original lazy reads; product admission is
// handled by its existing callers.
var rawDefaultDamageNativeEngine4E0B30 = rawDefaultDamageNative4E0B30

//export nox_xxx_damageDefaultProc_4E0B30
func nox_xxx_damageDefaultProc_4E0B30(target, source, weapon *C.nox_object_t, damage, typ C.int32_t) C.int32_t {
	return C.int32_t(rawDefaultDamageNative4E0B30(unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), int32(damage), int32(typ)))
}
