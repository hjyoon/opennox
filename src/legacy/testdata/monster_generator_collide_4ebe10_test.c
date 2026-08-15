// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert MonsterGeneratorCollide's native object, update-data and callback
// ABI boundaries.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, obj_class) ==
	(sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, data_update) ==
	(sizeof(void*) == 4 ? 748 : 872), "object update-data offset");
_Static_assert(offsetof(nox_monster_generator_update_data_t, collision_flags) ==
	(sizeof(void*) == 4 ? 72 : 120), "generator collision flags offset");
_Static_assert(offsetof(nox_monster_generator_update_data_t, collision_func) ==
	(sizeof(void*) == 4 ? 76 : 124), "generator collision function offset");
_Static_assert(sizeof(nox_monster_generator_update_data_t) ==
	(sizeof(void*) == 4 ? 164 : 216), "generator update-data size");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_collideMonsterGen_4EBE10),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"MonsterGeneratorCollide callback three-pointer ABI");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideMonsterGen_4EBE10(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int main(void) {
	nox_monster_generator_update_data_t update = {
		.collision_flags = UINT32_C(0xa5a55a5a),
		.collision_func = -17,
	};
	nox_object_t source = {.data_update = &update};
	nox_object_t target = {.obj_class = UINT32_C(0x40000004)};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collideMonsterGen_4EBE10(&source, &target, collision);
	if (seen_source != &source || seen_target != &target ||
		seen_collision != collision || source.data_update != &update ||
		target.obj_class != UINT32_C(0x40000004) ||
		update.collision_flags != UINT32_C(0xa5a55a5a) ||
		update.collision_func != -17) {
		return 1;
	}
	return 0;
}
