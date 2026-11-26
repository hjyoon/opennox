#include "mapgen_legacy_ptr.h"

uintptr_t nox_mapgen_resolve_address_go(uintptr_t value);

// Unregistered words pass through to C in pointer form.
void* nox_mapgenLegacyPtrResolve(uintptr_t value) {
	return (void*)nox_mapgen_resolve_address_go(value);
}
