#include <stdint.h>

#include "common/platform/platform.h"
#include "fixed_rng_seed_4ef560.h"

void sub_4EF560(void) { nox_platform_srand(UINT32_C(0x22D7)); }
