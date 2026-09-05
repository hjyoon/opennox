#include <assert.h>
#include <limits.h>

#include "../spell_duration_cleanup_4fe880.h"

typedef void (*spell_duration_cleanup_fn)(void);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(
	_Generic(&sub_4FE880, spell_duration_cleanup_fn: 1, default: 0),
	"004FE880 must preserve its no-argument void ABI");

static unsigned int observed_calls;

void sub_4FE880(void) {
	++observed_calls;
}

int main(void) {
	spell_duration_cleanup_fn const cleanup = sub_4FE880;
	cleanup();
	cleanup();
	assert(observed_calls == 2);
	return 0;
}
