#ifndef NOX_AWARD_SPELL_COLLIDE_4EAD20_H
#define NOX_AWARD_SPELL_COLLIDE_4EAD20_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_award_spell_collide_data_t {
	uint32_t spell;
} nox_award_spell_collide_data_t;

_Static_assert(sizeof(nox_award_spell_collide_data_t) == 4, "wrong size of AwardSpell collide data!");
_Static_assert(offsetof(nox_award_spell_collide_data_t, spell) == 0, "wrong offset of AwardSpell spell ID!");

int nox_xxx_collideSpellPedestal_4EAD20(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

#endif // NOX_AWARD_SPELL_COLLIDE_4EAD20_H
