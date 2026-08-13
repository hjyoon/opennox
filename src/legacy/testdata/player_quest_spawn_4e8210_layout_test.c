#include "../defs.h"

#include <stddef.h>
#include <stdint.h>

#ifdef NOX_PLAYER_QUEST_SPAWN_4E8210_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable exactly the object and SoulGate fields used here.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60), "object x offset");
_Static_assert(offsetof(nox_object_t, y) == (sizeof(void*) == 4 ? 60 : 64), "object y offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776),
	"object collide-data offset");
_Static_assert(sizeof(nox_soul_gate_collide_data_t) == 4, "SoulGate collide-data size");
_Static_assert(offsetof(nox_soul_gate_collide_data_t, last_used_frame) == 0,
	"SoulGate last-used frame offset");
#endif

static uint32_t player_quest_spawn_gate_frame_4e8210(const nox_object_t* gate) {
	return ((const nox_soul_gate_collide_data_t*)gate->collide_data)->last_used_frame;
}

int main(void) {
	nox_soul_gate_collide_data_t data = {.last_used_frame = UINT32_C(0xfedcba98)};
	nox_object_t gate = {0};
	gate.x = 17.0f;
	gate.y = -23.0f;
	gate.collide_data = &data;
	return player_quest_spawn_gate_frame_4e8210(&gate) != UINT32_C(0xfedcba98) ||
		gate.x != 17.0f || gate.y != -23.0f;
}
