// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert UndeadKillerCollide's native object, collision-data and callback
// ABI boundaries.
#include <stdio.h>

#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_object_t, obj_class) ==
	(sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, obj_subclass) ==
	(sizeof(void*) == 4 ? 12 : 16), "object subclass offset");
_Static_assert(offsetof(nox_object_t, health_data) ==
	(sizeof(void*) == 4 ? 556 : 616), "object health-data offset");
_Static_assert(offsetof(nox_object_t, collide_data) ==
	(sizeof(void*) == 4 ? 700 : 776), "object collide-data offset");
_Static_assert(offsetof(nox_object_t, func_damage) ==
	(sizeof(void*) == 4 ? 716 : 808), "object damage callback offset");
_Static_assert(sizeof(nox_undead_killer_collide_data_t) == sizeof(void*),
	"UndeadKiller collide-data native pointer width");
_Static_assert(offsetof(nox_undead_killer_collide_data_t, spell) == 0,
	"UndeadKiller spell pointer offset");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_collideUndeadKiller_4EBD40),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"UndeadKillerCollide callback three-pointer ABI");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideUndeadKiller_4EBD40(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

static int damage_probe(
	nox_object_t* target,
	nox_object_t* parent,
	nox_object_t* source,
	int32_t damage,
	int32_t damage_type) {
	(void)target;
	(void)parent;
	(void)source;
	return damage ^ damage_type;
}

int main(void) {
	uint16_t health[10] = {UINT16_C(0x1234)};
	uint32_t spell_words[30] = {0};
	nox_undead_killer_collide_data_t data = {.spell = spell_words};
	nox_object_t source = {.collide_data = &data};
	nox_object_t target = {
		.obj_class = UINT32_C(0x102),
		.obj_subclass = UINT32_C(0x240),
		.health_data = health,
		.func_damage = damage_probe,
	};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collideUndeadKiller_4EBD40(&source, &target, collision);
	if (seen_source != &source || seen_target != &target ||
		seen_collision != collision || source.collide_data != &data ||
		data.spell != spell_words || target.health_data != health ||
		target.func_damage != damage_probe || target.obj_class != UINT32_C(0x102) ||
		target.obj_subclass != UINT32_C(0x240)) {
		return 1;
	}
	return 0;
}
