// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert only ArrowCollide's native object, data and callback ABI.
#include <stdio.h>

#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_arrow_collide_data_t) ==
	(sizeof(void*) == 4 ? 8 : 16), "Arrow collide-data size");
_Static_assert(offsetof(nox_arrow_collide_data_t, field_0) == 0,
	"Arrow collide field-zero offset");
_Static_assert(offsetof(nox_arrow_collide_data_t, owner) == sizeof(void*),
	"Arrow collide owner offset");

_Static_assert(sizeof(nox_arrow_attack_data_t) ==
	(sizeof(void*) == 4 ? 32 : 48), "Arrow attack-data size");
_Static_assert(offsetof(nox_arrow_attack_data_t, owner) ==
	(sizeof(void*) == 4 ? 12 : 16), "Arrow attack owner offset");
_Static_assert(offsetof(nox_arrow_attack_data_t, x) ==
	(sizeof(void*) == 4 ? 16 : 24), "Arrow attack X offset");
_Static_assert(offsetof(nox_arrow_attack_data_t, field_24) ==
	(sizeof(void*) == 4 ? 24 : 32), "Arrow attack field-24 offset");
_Static_assert(offsetof(nox_arrow_attack_data_t, source) ==
	(sizeof(void*) == 4 ? 28 : 40), "Arrow attack source offset");

_Static_assert(offsetof(nox_object_t, typ_ind) ==
	(sizeof(void*) == 4 ? 4 : 8), "object type-index offset");
_Static_assert(offsetof(nox_object_t, obj_class) ==
	(sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, obj_flags) ==
	(sizeof(void*) == 4 ? 16 : 20), "object flags offset");
_Static_assert(offsetof(nox_object_t, x) ==
	(sizeof(void*) == 4 ? 56 : 60), "object position offset");
_Static_assert(offsetof(nox_object_t, shape) ==
	(sizeof(void*) == 4 ? 172 : 176), "object shape offset");
_Static_assert(offsetof(nox_object_t, owner) ==
	(sizeof(void*) == 4 ? 508 : 552), "object owner offset");
_Static_assert(offsetof(nox_object_t, health_data) ==
	(sizeof(void*) == 4 ? 556 : 616), "object health offset");
_Static_assert(offsetof(nox_object_t, collide_data) ==
	(sizeof(void*) == 4 ? 700 : 776), "object collide-data offset");
_Static_assert(offsetof(nox_object_t, func_damage) ==
	(sizeof(void*) == 4 ? 716 : 808), "object damage callback offset");

_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_collideArrow_4EB490),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"Arrow collide callback pointer width");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_server_arrowCollideDataSetOwner_4EB490),
	void (*)(nox_object_t*, nox_object_t*)),
	"Arrow collide-data setter pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideArrow_4EB490(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

void nox_server_arrowCollideDataSetOwner_4EB490(
	nox_object_t* source,
	nox_object_t* owner) {
	((nox_arrow_collide_data_t*)source->collide_data)->owner = owner;
}

int main(void) {
	nox_object_t owner = {0};
	nox_arrow_collide_data_t data = {.field_0 = UINT32_C(0x89abcdef)};
	nox_object_t source = {.collide_data = &data};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};
	nox_arrow_attack_data_t attack = {
		.damage = 7.25f,
		.damage_type = UINT8_C(11),
		.radius = 6.75f,
		.owner = &owner,
		.x = 12.5f,
		.y = -4.25f,
		.field_24 = UINT32_C(0),
		.source = &source,
	};

	nox_server_arrowCollideDataSetOwner_4EB490(&source, &owner);
	nox_xxx_collideArrow_4EB490(&source, &target, collision);
	if (seen_source != &source || seen_target != &target ||
		seen_collision != collision || data.owner != &owner) {
		return 1;
	}
	if (data.field_0 != UINT32_C(0x89abcdef) || attack.owner != &owner ||
		attack.source != &source || attack.damage != 7.25f ||
		attack.damage_type != UINT8_C(11) || attack.radius != 6.75f ||
		attack.x != 12.5f || attack.y != -4.25f || attack.field_24 != 0) {
		return 2;
	}
	return 0;
}
