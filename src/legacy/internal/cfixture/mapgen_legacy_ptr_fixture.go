package cfixture

/*
#include "mapgen_legacy_ptr.h"
#include <stddef.h>

_Static_assert(_Generic(&nox_mapgenLegacyPtrRegister, uint32_t (*)(void*): 1, default: 0), "mapgen token registration ABI");
_Static_assert(_Generic(&nox_mapgenLegacyPtrForget, void (*)(void*): 1, default: 0), "mapgen token release ABI");
_Static_assert(_Generic(&nox_mapgenLegacyPtrResolve, uintptr_t (*)(uintptr_t): 1, default: 0), "mapgen token resolution ABI");

static int mapgen_legacy_ptr_lifecycle(void* owner) {
	uint32_t token = nox_mapgenLegacyPtrRegister(owner);
	int registered = token != 0 && nox_mapgenLegacyPtrRegister(owner) == token &&
		nox_mapgenLegacyPtrResolve(token) == (uintptr_t)owner &&
		nox_mapgenLegacyPtrResolve((uintptr_t)owner) == (uintptr_t)owner;
	nox_mapgenLegacyPtrForget(owner);
	nox_mapgenLegacyPtrForget(NULL);
	return registered && nox_mapgenLegacyPtrResolve(token) == token &&
		nox_mapgenLegacyPtrRegister(NULL) == 0 && nox_mapgenLegacyPtrResolve(0) == 0 &&
		nox_mapgenLegacyPtrResolve(1) == 1;
}
*/
import "C"

import "unsafe"

func MapgenLegacyPointerLifecycle(owner unsafe.Pointer) bool {
	return C.mapgen_legacy_ptr_lifecycle(owner) != 0
}
