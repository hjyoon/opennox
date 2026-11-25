#include <stdint.h>

#include "common/platform/platform.h"
#include "fixed_rng_seed_4f0630.h"

void sub_4F0630(void) { nox_platform_srand(UINT32_C(0x7DB)); }
