#include <stdint.h>

#include "common/platform/platform.h"
#include "fixed_rng_seed_4ef570.h"

void sub_4EF570(void) { nox_platform_srand(UINT32_C(0x7DA)); }
