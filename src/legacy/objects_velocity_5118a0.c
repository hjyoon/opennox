#include <stdint.h>
#include <string.h>

#include "GAME4_1.h"

extern int32_t nox_update_objects_velocity_bits_5118a0(uint32_t);

// Keep the public float ABI while preserving its raw DWORD across cgo's
// Windows/386 float argument copies, which otherwise quiet signaling NaNs.
int nox_xxx_updateObjectsVelocity_5118A0(float step) {
	uint32_t bits;
	memcpy(&bits, &step, sizeof(bits));
	return nox_update_objects_velocity_bits_5118a0(bits);
}
