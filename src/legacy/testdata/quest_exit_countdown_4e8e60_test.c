// Suppress unrelated Win32-only declarations while parsing the shared legacy
// headers, then assert every C field exposed by GAME.EXE 004E8E60.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(((nox_player_update_data_t*)0)->quest_exit) == sizeof(void*),
	"QuestExit pointer width");
_Static_assert(sizeof(((nox_player_update_data_t*)0)->quest_warp_gate) == sizeof(void*),
	"QuestWarpGate pointer width");
_Static_assert(offsetof(nox_player_update_data_t, quest_exit) ==
	(sizeof(void*) == 4 ? 312 : 400), "QuestExit offset");
_Static_assert(offsetof(nox_player_update_data_t, quest_warp_gate) ==
	(sizeof(void*) == 4 ? 316 : 408), "QuestWarpGate offset");
_Static_assert(sizeof(nox_player_update_data_t) == (sizeof(void*) == 4 ? 320 : 416),
	"partial PlayerUpdate size");

int32_t sub_4E8E60(void) { return INT32_C(-2147483601); }

static int32_t (*const quest_exit_countdown_signature)(void) = sub_4E8E60;

int main(void) {
	nox_player_update_data_t update = {0};
	nox_object_t exit_object = {0};
	nox_object_t warp_object = {0};
	update.quest_exit = &exit_object;
	update.quest_warp_gate = &warp_object;
	if (update.quest_exit != &exit_object || update.quest_warp_gate != &warp_object) {
		return 1;
	}
	return quest_exit_countdown_signature() == INT32_C(-2147483601) ? 0 : 2;
}
