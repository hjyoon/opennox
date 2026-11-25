#include <stdint.h>

#include "common/platform/platform.h"
#include "fixed_rng_seed_4f3e20.h"

void sub_4F3E20(void) { nox_platform_srand(UINT32_C(0x4E2A)); }
