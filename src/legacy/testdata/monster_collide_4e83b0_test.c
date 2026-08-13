// Suppress unrelated Win32-only declarations while the shared legacy headers
// are parsed, then restore and assert every exposed field used by 004E83B0.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_script_callback_t) == 8, "script callback size");
_Static_assert(offsetof(nox_script_callback_t, flags) == 0, "script callback flags offset");
_Static_assert(offsetof(nox_script_callback_t, func) == 4, "script callback function offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");

static nox_object_t* seen_monster;
static nox_object_t* seen_other;
static uint32_t result_word;

void* nox_xxx_collideMonsterEventProc_4E83B0(nox_object_t* monster, nox_object_t* other) {
	seen_monster = monster;
	seen_other = other;
	return &result_word;
}

static void* (*const collide_signature)(nox_object_t*, nox_object_t*) =
	nox_xxx_collideMonsterEventProc_4E83B0;

int main(void) {
	nox_object_t monster = {0};
	nox_object_t other = {0};
	monster.data_update = &result_word;
	if (collide_signature(&monster, &other) != &result_word) {
		return 1;
	}
	if (seen_monster != &monster || seen_other != &other) {
		return 2;
	}
	if (collide_signature(&monster, 0) != &result_word || seen_other != 0) {
		return 3;
	}
	return 0;
}
