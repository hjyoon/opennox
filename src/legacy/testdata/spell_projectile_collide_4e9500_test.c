// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert this callback's native pointer-width records and signature.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_spell_projectile_update_data_t) == (sizeof(void*) == 4 ? 28 : 40),
	"SpellProjectile update-data size");
_Static_assert(offsetof(nox_spell_projectile_update_data_t, owner) == 0,
	"SpellProjectile owner offset");
_Static_assert(offsetof(nox_spell_projectile_update_data_t, target) == sizeof(void*),
	"SpellProjectile target offset");
_Static_assert(offsetof(nox_spell_projectile_update_data_t, source) == 2 * sizeof(void*),
	"SpellProjectile source offset");
_Static_assert(offsetof(nox_spell_projectile_update_data_t, spell) == 3 * sizeof(void*),
	"SpellProjectile spell offset");
_Static_assert(offsetof(nox_spell_projectile_update_data_t, level) == 3 * sizeof(void*) + 4,
	"SpellProjectile level offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60),
	"object position offset");
_Static_assert(offsetof(nox_object_t, new_x) == (sizeof(void*) == 4 ? 64 : 68),
	"object new-position offset");
_Static_assert(offsetof(nox_object_t, prev_x) == (sizeof(void*) == 4 ? 72 : 76),
	"object previous-position offset");
_Static_assert(offsetof(nox_object_t, vel_x) == (sizeof(void*) == 4 ? 80 : 84),
	"object velocity offset");
_Static_assert(offsetof(nox_object_t, direction1) == (sizeof(void*) == 4 ? 124 : 128),
	"object direction offset");
_Static_assert(offsetof(nox_object_t, inv_next_item) == (sizeof(void*) == 4 ? 496 : 528),
	"object inventory-next offset");
_Static_assert(offsetof(nox_object_t, inv_first_item) == (sizeof(void*) == 4 ? 504 : 544),
	"object inventory-first offset");
_Static_assert(offsetof(nox_object_t, init_data) == (sizeof(void*) == 4 ? 692 : 760),
	"object init-data offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");
_Static_assert(offsetof(nox_player_update_data_t, state) == 88,
	"player-update state offset");
_Static_assert(offsetof(nox_player_update_data_t, field_59_0) == (sizeof(void*) == 4 ? 236 : 280),
	"player-update animation frame offset");
_Static_assert(offsetof(nox_player_update_data_t, player) == (sizeof(void*) == 4 ? 276 : 320),
	"player-update player offset");
_Static_assert(offsetof(nox_playerInfo, field_4) == 4,
	"player weapon-equip offset");
_Static_assert(
	__builtin_types_compatible_p(
		__typeof__(&nox_xxx_spellFlyCollide_4E9500),
		void (*)(nox_object_t*, nox_object_t*, float*)),
	"SpellProjectileCollide callback pointer width");

static nox_object_t* seen_projectile;
static nox_object_t* seen_other;
static float* seen_collision;

void nox_xxx_spellFlyCollide_4E9500(
	nox_object_t* projectile,
	nox_object_t* other,
	float* collision) {
	seen_projectile = projectile;
	seen_other = other;
	seen_collision = collision;
}

int main(void) {
	nox_object_t projectile = {0};
	nox_object_t other = {0};
	float collision[2] = {3.5f, -8.25f};
	nox_object_t owner = {0};
	nox_object_t source = {0};
	nox_spell_projectile_update_data_t update = {
		.owner = &owner,
		.target = &other,
		.source = &source,
		.spell = UINT32_C(0x89abcdef),
		.level = UINT32_C(0x10203040),
		.field_20 = UINT32_C(0x55667788),
		.field_24 = UINT32_C(0xa5a5a5a5),
	};

	nox_xxx_spellFlyCollide_4E9500(&projectile, &other, collision);
	if (seen_projectile != &projectile || seen_other != &other || seen_collision != collision) {
		return 1;
	}
	if (update.owner != &owner || update.target != &other || update.source != &source ||
		update.spell != UINT32_C(0x89abcdef) || update.level != UINT32_C(0x10203040) ||
		update.field_20 != UINT32_C(0x55667788) || update.field_24 != UINT32_C(0xa5a5a5a5)) {
		return 2;
	}
	return 0;
}
