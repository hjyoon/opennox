#ifndef NOX_MAPGEN_LEGACY_PTR_H
#define NOX_MAPGEN_LEGACY_PTR_H

#include <stdint.h>

// Map generation records retain GAME.EXE's fixed PE32 layout. Pointer fields
// in those records therefore hold a 32-bit token while the backing allocation
// remains a native-width host pointer.
uint32_t nox_mapgenLegacyPtrRegister(void* ptr);
void* nox_mapgenLegacyPtrResolve(uintptr_t value);
void nox_mapgenLegacyPtrForget(void* ptr);

#endif // NOX_MAPGEN_LEGACY_PTR_H
