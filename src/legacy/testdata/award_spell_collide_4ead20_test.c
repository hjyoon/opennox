// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert AwardSpellCollide's native object, data and callback ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_award_spell_collide_data_t) == 4, "AwardSpell collide-data size");
_Static_assert(offsetof(nox_award_spell_collide_data_t, spell) == 0, "AwardSpell ID offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776), "object collide-data offset");
_Static_assert(__builtin_types_compatible_p(__typeof__(&nox_xxx_collideSpellPedestal_4EAD20),
											int (*)(nox_object_t*, nox_object_t*, float*)),
			   "AwardSpellCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

int nox_xxx_collideSpellPedestal_4EAD20(
	nox_object_t* source,
	nox_object_t* target,
	float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
	return -INT32_C(19088743);
}

static nox_object_t* grant_target;
static uint32_t grant_spell;
static int32_t grant_mode;
static int32_t grant_fourth;
static int32_t grant_fifth;

static int32_t award_spell_grant_reference(
	nox_object_t* target,
	uint32_t spell,
	int32_t mode,
	int32_t fourth,
	int32_t fifth) {
	grant_target = target;
	grant_spell = spell;
	grant_mode = mode;
	grant_fourth = fourth;
	grant_fifth = fifth;
	return -INT32_C(305419896);
}

static int32_t award_spell_reference(nox_object_t* source, nox_object_t* target) {
	if (target == NULL) {
		return 0;
	}
	nox_award_spell_collide_data_t* data = source->collide_data;
	return award_spell_grant_reference(target, data->spell, 1, 0, 0);
}

int main(void) {
	nox_award_spell_collide_data_t data = {.spell = UINT32_C(0xf1234567)};
	nox_object_t source = {.collide_data = &data};
	nox_object_t target = {0};
	float collision[2] = {3.5f, -8.25f};

	int callback_result = nox_xxx_collideSpellPedestal_4EAD20(&source, &target, collision);
	if (callback_result != -INT32_C(19088743) || seen_source != &source ||
		seen_target != &target || seen_collision != collision) {
		return 1;
	}

	int32_t result = award_spell_reference(&source, &target);
	if (result != -INT32_C(305419896) || grant_target != &target ||
		grant_spell != UINT32_C(0xf1234567) || grant_mode != 1 ||
		grant_fourth != 0 || grant_fifth != 0) {
		return 2;
	}
	if (award_spell_reference(NULL, NULL) != 0) {
		return 3;
	}
	if (collision[0] != 3.5f || collision[1] != -8.25f) {
		return 4;
	}
	return 0;
}
