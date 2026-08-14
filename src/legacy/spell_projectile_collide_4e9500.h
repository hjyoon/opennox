#ifndef NOX_SPELL_PROJECTILE_COLLIDE_4E9500_H
#define NOX_SPELL_PROJECTILE_COLLIDE_4E9500_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_spell_projectile_update_data_t {
	nox_object_t* owner;
	nox_object_t* target;
	nox_object_t* source;
	uint32_t spell;
	uint32_t level;
	uint32_t field_20;
	uint32_t field_24;
} nox_spell_projectile_update_data_t;

_Static_assert(sizeof(nox_spell_projectile_update_data_t) == (sizeof(void*) == 4 ? 28 : 40),
	"wrong size of SpellProjectile update data");
_Static_assert(offsetof(nox_spell_projectile_update_data_t, owner) == 0,
	"wrong offset of SpellProjectile owner");
_Static_assert(offsetof(nox_spell_projectile_update_data_t, target) == sizeof(void*),
	"wrong offset of SpellProjectile target");
_Static_assert(offsetof(nox_spell_projectile_update_data_t, source) == 2 * sizeof(void*),
	"wrong offset of SpellProjectile source");
_Static_assert(offsetof(nox_spell_projectile_update_data_t, spell) == 3 * sizeof(void*),
	"wrong offset of SpellProjectile spell");
_Static_assert(offsetof(nox_spell_projectile_update_data_t, level) == 3 * sizeof(void*) + 4,
	"wrong offset of SpellProjectile level");

void nox_xxx_spellFlyCollide_4E9500(
	nox_object_t* projectile,
	nox_object_t* other,
	float* collision);

#endif // NOX_SPELL_PROJECTILE_COLLIDE_4E9500_H
