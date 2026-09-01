#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../fixed_rng_seed_4fb940.h"

typedef void (*fixed_rng_seed_fn)(void);

_Static_assert(CHAR_BIT == 8, "seed bytes must remain eight bits");
_Static_assert(sizeof(unsigned int) == 4, "platform RNG seed must remain uint32");
_Static_assert(
	_Generic(&sub_4FB940, fixed_rng_seed_fn: 1, default: 0),
	"first fixed RNG seed wrapper must take no arguments and return void");
_Static_assert(
	_Generic(&sub_4FB950, fixed_rng_seed_fn: 1, default: 0),
	"second fixed RNG seed wrapper must take no arguments and return void");
_Static_assert(
	_Generic(&sub_4FC560, fixed_rng_seed_fn: 1, default: 0),
	"third fixed RNG seed wrapper must take no arguments and return void");

static unsigned int observed_seeds[6];
static unsigned int observed_calls;

void nox_platform_srand(unsigned int seed) {
	assert(observed_calls < sizeof(observed_seeds) / sizeof(observed_seeds[0]));
	observed_seeds[observed_calls++] = seed;
}

int main(void) {
	fixed_rng_seed_fn const first = sub_4FB940;
	fixed_rng_seed_fn const second = sub_4FB950;
	fixed_rng_seed_fn const third = sub_4FC560;

	first();
	second();
	third();
	first();
	second();
	third();

	assert(observed_calls == 6);
	assert(observed_seeds[0] == UINT32_C(0x143D));
	assert(observed_seeds[1] == UINT32_C(0x22EA));
	assert(observed_seeds[2] == UINT32_C(0x22EB));
	assert(observed_seeds[3] == UINT32_C(0x143D));
	assert(observed_seeds[4] == UINT32_C(0x22EA));
	assert(observed_seeds[5] == UINT32_C(0x22EB));
	return 0;
}
