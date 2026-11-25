#include <stdint.h>

#include "common/platform/platform.h"
#include "fixed_rng_seed_4eed30.h"

void sub_4EED30(void) { nox_platform_srand(UINT32_C(0x22D6)); }
