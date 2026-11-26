//go:build !amd64 && !arm64

#include <stdint.h>
#include <string.h>

#include "GAME4_1.h"

extern int32_t nox_update_objects_velocity_bits_5118a0(uint32_t);

// Preserve the raw DWORD when generated cgo argument copies can use x87
// and quiet signaling NaNs.
int nox_xxx_updateObjectsVelocity_5118A0(float step) {
	uint32_t bits;
	memcpy(&bits, &step, sizeof(bits));
	return nox_update_objects_velocity_bits_5118a0(bits);
}
