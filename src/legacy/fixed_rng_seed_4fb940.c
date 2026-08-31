#include <stdint.h>

#include "common/platform/platform.h"
#include "fixed_rng_seed_4fb940.h"

void sub_4FB940(void) { nox_platform_srand(UINT32_C(0x143D)); }

void sub_4FB950(void) { nox_platform_srand(UINT32_C(0x22EA)); }
