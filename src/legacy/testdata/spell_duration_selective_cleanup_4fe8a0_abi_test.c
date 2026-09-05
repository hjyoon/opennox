#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../spell_duration_selective_cleanup_4fe8a0.h"

typedef void (*spell_duration_selective_cleanup_fn)(int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "004FE8A0 mode must remain exact int32");
_Static_assert(
	_Generic(&sub_4FE8A0, spell_duration_selective_cleanup_fn: 1, default: 0),
	"004FE8A0 must receive one exact int32 mode");

static int32_t observed_mode;
static unsigned int observed_calls;

void sub_4FE8A0(int32_t mode) {
	observed_mode = mode;
	++observed_calls;
}

static void check_call(spell_duration_selective_cleanup_fn cleanup, int32_t mode) {
	cleanup(mode);
	assert(observed_mode == mode);
}

int main(void) {
	spell_duration_selective_cleanup_fn const cleanup = sub_4FE8A0;

	check_call(cleanup, INT32_MIN);
	check_call(cleanup, INT32_C(-1));
	check_call(cleanup, INT32_C(0));
	check_call(cleanup, INT32_C(1));
	check_call(cleanup, INT32_MAX);
	assert(observed_calls == 5);
	return 0;
}
