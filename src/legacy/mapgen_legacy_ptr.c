#include "mapgen_legacy_ptr.h"

uint32_t nox_mapgen_register_address_go(uintptr_t address);
uintptr_t nox_mapgen_resolve_address_go(uintptr_t value);
void nox_mapgen_forget_address_go(uintptr_t address);

uint32_t nox_mapgenLegacyPtrRegister(void* ptr) {
	return nox_mapgen_register_address_go((uintptr_t)ptr);
}

void* nox_mapgenLegacyPtrResolve(uintptr_t value) {
	return (void*)nox_mapgen_resolve_address_go(value);
}

void nox_mapgenLegacyPtrForget(void* ptr) {
	nox_mapgen_forget_address_go((uintptr_t)ptr);
}
