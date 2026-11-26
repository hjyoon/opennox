//go:build !amd64 && !arm64

#include <stdint.h>
#include <string.h>

#include "GAME3_3.h"

extern void nox_xxx_mapDamageUnitsAroundBits_4E25B0(float*, uint32_t, uint32_t, int32_t, int32_t,
                                                 nox_object_t*, nox_object_t*);

// Preserve the incoming DWORDs before crossing cgo. Generated float
// argument copies can use x87 on 386 and quiet signaling NaNs before
// the radial damage routine reads the original radius words.
void nox_xxx_mapDamageUnitsAround_4E25B0(float* center, float outer, float inner, int damage, int typ,
                                      nox_object_t* source, nox_object_t* excluded) {
	uint32_t outer_bits, inner_bits;
	memcpy(&outer_bits, &outer, sizeof(outer_bits));
	memcpy(&inner_bits, &inner, sizeof(inner_bits));
	nox_xxx_mapDamageUnitsAroundBits_4E25B0(center, outer_bits, inner_bits, damage, typ, source, excluded);
}
