#ifndef OPENNOX_GAME4_MAP_ARCHIVE_H
#define OPENNOX_GAME4_MAP_ARCHIVE_H

#include "GAME4.h"

// Private Go entries use writable C pointer spellings because cgo exports
// cannot spell const. The public ABI below retains GAME4.h's const contract.
void game4MapBuildPlaceBounds(void*, void*, void*, void*, void*);
int game4MapPlaceGeometry(float, float, float, float, void*, int32_t, int32_t, int32_t*, int32_t*);
int game4MapOpen(char*);

#endif
