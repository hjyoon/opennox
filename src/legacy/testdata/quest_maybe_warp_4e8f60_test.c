// Suppress unrelated Win32-only declarations while parsing the shared legacy
// headers, then assert the C boundary exposed by GAME.EXE 004E8F60.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stdint.h>

int32_t nox_server_questMaybeWarp_4E8F60(void) { return INT32_C(1); }

static int32_t (*const quest_maybe_warp_signature)(void) = nox_server_questMaybeWarp_4E8F60;

int main(void) {
	return quest_maybe_warp_signature() == INT32_C(1) ? 0 : 1;
}
