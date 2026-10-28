#include <stdint.h>
#include <string.h>

#include "GAME3_3.h"

extern void nox_xxx_mapDamageUnitsAroundBits_4E25B0(float*, uint32_t, uint32_t, int32_t, int32_t,
                                                 nox_object_t*, nox_object_t*);

// Preserve the incoming DWORDs before crossing cgo. Its generated float
// argument copies use x87 on Windows/386 and quiet signaling NaNs, although
// GAME.EXE copies these two radii without evaluating them.
void nox_xxx_mapDamageUnitsAround_4E25B0(float* center, float outer, float inner, int damage, int typ,
                                      nox_object_t* source, nox_object_t* excluded) {
	uint32_t outer_bits, inner_bits;
	memcpy(&outer_bits, &outer, sizeof(outer_bits));
	memcpy(&inner_bits, &inner, sizeof(inner_bits));
	nox_xxx_mapDamageUnitsAroundBits_4E25B0(center, outer_bits, inner_bits, damage, typ, source, excluded);
}
