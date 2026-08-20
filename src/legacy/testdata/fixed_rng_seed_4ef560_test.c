#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../fixed_rng_seed_4ef560.h"

typedef void (*fixed_rng_seed_fn)(void);

_Static_assert(CHAR_BIT == 8, "seed bytes must remain eight bits");
_Static_assert(sizeof(unsigned int) == 4, "platform RNG seed must remain uint32");
_Static_assert(
	_Generic(&sub_4EF560, fixed_rng_seed_fn: 1, default: 0),
	"fixed RNG seed wrapper must take no arguments and return void");

static unsigned int observed_seed;
static unsigned int observed_calls;

void nox_platform_srand(unsigned int seed) {
	observed_seed = seed;
	++observed_calls;
}

int main(void) {
	fixed_rng_seed_fn const seed = sub_4EF560;

	seed();
	assert(observed_calls == 1);
	assert(observed_seed == UINT32_C(0x22D7));

	seed();
	assert(observed_calls == 2);
	assert(observed_seed == UINT32_C(0x22D7));
	return 0;
}
