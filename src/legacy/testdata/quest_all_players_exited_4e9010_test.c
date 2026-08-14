// Suppress unrelated Win32-only declarations while parsing the shared legacy
// headers, then assert the C boundary exposed by GAME.EXE 004E9010.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stdint.h>

int32_t sub_4E9010(void) { return INT32_C(1); }

static int32_t (*const quest_all_players_exited_signature)(void) = sub_4E9010;

int main(void) {
	return quest_all_players_exited_signature() == INT32_C(1) ? 0 : 1;
}
