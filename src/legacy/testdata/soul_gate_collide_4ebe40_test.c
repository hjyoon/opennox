// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert only SoulGateCollide's native callback and record boundaries.
#include <stdio.h>

#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, obj_class) ==
	(sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, x) ==
	(sizeof(void*) == 4 ? 56 : 60), "object position offset");
_Static_assert(offsetof(nox_object_t, collide_data) ==
	(sizeof(void*) == 4 ? 700 : 776), "object collide-data offset");
_Static_assert(offsetof(nox_object_t, data_update) ==
	(sizeof(void*) == 4 ? 748 : 872), "object update-data offset");
_Static_assert(offsetof(nox_player_update_data_t, player) ==
	(sizeof(void*) == 4 ? 276 : 320), "player pointer offset");
_Static_assert(offsetof(nox_player_update_data_t, soul_gate) ==
	(sizeof(void*) == 4 ? 308 : 376), "SoulGate pointer offset");
_Static_assert(sizeof(nox_player_update_data_t) ==
	(sizeof(void*) == 4 ? 320 : 400), "partial player update-data size");
_Static_assert(sizeof(nox_soul_gate_collide_data_t) == 4,
	"SoulGate collide-data size");
_Static_assert(offsetof(nox_soul_gate_collide_data_t, last_used_frame) == 0,
	"SoulGate last-used frame offset");
_Static_assert(sizeof(void*) != 4 || offsetof(nox_playerInfo, field_4792) == 4792,
	"Win32 Player Quest-state offset");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&sub_4EBE40),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"SoulGateCollide callback three-pointer ABI");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void sub_4EBE40(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int main(void) {
	nox_soul_gate_collide_data_t data = {
		.last_used_frame = UINT32_C(0xfedcba98),
	};
	nox_player_update_data_t update = {
		.soul_gate = NULL,
	};
	nox_object_t source = {
		.x = 12.5f,
		.y = -7.25f,
		.collide_data = &data,
	};
	nox_object_t target = {
		.obj_class = UINT32_C(0x40000004),
		.data_update = &update,
	};
	float collision[2] = {3.5f, -8.25f};

	sub_4EBE40(&source, &target, collision);
	if (seen_source != &source || seen_target != &target ||
		seen_collision != collision || source.collide_data != &data ||
		target.data_update != &update || update.soul_gate != NULL ||
		data.last_used_frame != UINT32_C(0xfedcba98) ||
		collision[0] != 3.5f || collision[1] != -8.25f) {
		return 1;
	}
	return 0;
}
