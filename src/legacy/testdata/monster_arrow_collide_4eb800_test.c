// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert only MonsterArrowCollide's fixed data, callback, and parser ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#include "../GAME4_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_monster_arrow_collide_data_t) == 8,
	"MonsterArrow collide-data size");
_Static_assert(offsetof(nox_monster_arrow_collide_data_t, coop_damage) == 0,
	"MonsterArrow Coop damage offset");
_Static_assert(offsetof(nox_monster_arrow_collide_data_t, other_damage) == 4,
	"MonsterArrow other damage offset");
_Static_assert(offsetof(nox_object_t, obj_flags) ==
	(sizeof(void*) == 4 ? 16 : 20), "object flags offset");
_Static_assert(offsetof(nox_object_t, collide_data) ==
	(sizeof(void*) == 4 ? 700 : 776), "object collide-data offset");
_Static_assert(offsetof(nox_object_t, func_damage) ==
	(sizeof(void*) == 4 ? 716 : 808), "object damage callback offset");

_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_collideMonsterArrow_4EB800),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"MonsterArrow collide callback pointer width");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&sub_536E80),
	int (*)(char*, nox_monster_arrow_collide_data_t*)),
	"MonsterArrow parser data width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideMonsterArrow_4EB800(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int main(void) {
	nox_monster_arrow_collide_data_t data = {
		.coop_damage = INT32_C(-31),
		.other_damage = INT32_C(47),
	};
	nox_object_t source = {.collide_data = &data};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};
	char positive[] = "17 -9";
	char negative[] = "   -31 47";

	nox_xxx_collideMonsterArrow_4EB800(&source, &target, collision);
	if (seen_source != &source || seen_target != &target ||
		seen_collision != collision || source.collide_data != &data) {
		return 1;
	}
	if (sub_536E80(positive, &data) != 1 ||
		data.coop_damage != 17 || data.other_damage != 17) {
		return 2;
	}
	if (sub_536E80(negative, &data) != 1 ||
		data.coop_damage != -31 || data.other_damage != -31) {
		return 3;
	}
	return 0;
}
