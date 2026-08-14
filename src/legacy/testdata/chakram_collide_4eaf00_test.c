// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert only ChakramInMotionCollide's native object, data and callback ABI.
#include <stdio.h>

#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_chakram_update_data_t) ==
	(sizeof(void*) == 4 ? 28 : 40), "Chakram update-data size");
_Static_assert(offsetof(nox_chakram_update_data_t, return_target) == 8,
	"Chakram return-target offset");
_Static_assert(offsetof(nox_chakram_update_data_t, last_hit) ==
	8 + sizeof(void*), "Chakram last-hit offset");
_Static_assert(offsetof(nox_chakram_update_data_t, return_state) ==
	16 + 2 * sizeof(void*), "Chakram return-state offset");

_Static_assert(sizeof(nox_chakram_attack_data_t) ==
	(sizeof(void*) == 4 ? 32 : 48), "Chakram attack-data size");
_Static_assert(offsetof(nox_chakram_attack_data_t, owner) ==
	(sizeof(void*) == 4 ? 12 : 16), "Chakram attack owner offset");
_Static_assert(offsetof(nox_chakram_attack_data_t, x) ==
	(sizeof(void*) == 4 ? 16 : 24), "Chakram attack X offset");
_Static_assert(offsetof(nox_chakram_attack_data_t, field_24) ==
	(sizeof(void*) == 4 ? 24 : 32), "Chakram attack field-24 offset");
_Static_assert(offsetof(nox_chakram_attack_data_t, source) ==
	(sizeof(void*) == 4 ? 28 : 40), "Chakram attack source offset");

_Static_assert(offsetof(nox_object_t, material) == (sizeof(void*) == 4 ? 24 : 28),
	"object material offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60),
	"object position offset");
_Static_assert(offsetof(nox_object_t, vel_x) == (sizeof(void*) == 4 ? 80 : 84),
	"object velocity offset");
_Static_assert(offsetof(nox_object_t, inv_first_item) ==
	(sizeof(void*) == 4 ? 504 : 544), "object inventory-first offset");
_Static_assert(offsetof(nox_object_t, owner) ==
	(sizeof(void*) == 4 ? 508 : 552), "object owner offset");
_Static_assert(offsetof(nox_object_t, data_update) ==
	(sizeof(void*) == 4 ? 748 : 872), "object update-data offset");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_collideChakram_4EAF00),
	void (*)(nox_object_t*, nox_object_t*, float*)),
	"Chakram collide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideChakram_4EAF00(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

int main(void) {
	nox_object_t owner = {0};
	nox_object_t source = {.owner = &owner};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};
	nox_chakram_attack_data_t attack = {
		.damage = 7.25f,
		.damage_type = UINT8_C(9),
		.radius = 36.0f,
		.owner = &owner,
		.x = 12.5f,
		.y = -4.25f,
		.field_24 = UINT32_C(0x89abcdef),
		.source = &source,
	};

	nox_xxx_collideChakram_4EAF00(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}
	if (attack.owner != &owner || attack.source != &source ||
		attack.damage != 7.25f || attack.damage_type != UINT8_C(9) ||
		attack.radius != 36.0f || attack.x != 12.5f || attack.y != -4.25f ||
		attack.field_24 != UINT32_C(0x89abcdef)) {
		return 2;
	}
	return 0;
}
