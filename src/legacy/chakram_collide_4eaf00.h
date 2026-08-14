#ifndef NOX_CHAKRAM_COLLIDE_4EAF00_H
#define NOX_CHAKRAM_COLLIDE_4EAF00_H

#include "chakram_update_53dcc0.h"

#include <stddef.h>
#include <stdint.h>

typedef struct nox_chakram_attack_data_t {
	float damage;
	uint8_t damage_type;
	uint8_t padding_5[3];
	float radius;
	nox_object_t* owner;
	float x;
	float y;
	uint32_t field_24;
	nox_object_t* source;
} nox_chakram_attack_data_t;

_Static_assert(offsetof(nox_chakram_attack_data_t, damage) == 0,
	"wrong offset of Chakram attack damage!");
_Static_assert(offsetof(nox_chakram_attack_data_t, damage_type) == 4,
	"wrong offset of Chakram attack damage type!");
_Static_assert(offsetof(nox_chakram_attack_data_t, radius) == 8,
	"wrong offset of Chakram attack radius!");
_Static_assert(offsetof(nox_chakram_attack_data_t, owner) ==
	(sizeof(void*) == 4 ? 12 : 16), "wrong offset of Chakram attack owner!");
_Static_assert(offsetof(nox_chakram_attack_data_t, x) ==
	(sizeof(void*) == 4 ? 16 : 24), "wrong offset of Chakram attack X!");
_Static_assert(offsetof(nox_chakram_attack_data_t, y) ==
	(sizeof(void*) == 4 ? 20 : 28), "wrong offset of Chakram attack Y!");
_Static_assert(offsetof(nox_chakram_attack_data_t, field_24) ==
	(sizeof(void*) == 4 ? 24 : 32), "wrong offset of Chakram attack field 24!");
_Static_assert(offsetof(nox_chakram_attack_data_t, source) ==
	(sizeof(void*) == 4 ? 28 : 40), "wrong offset of Chakram attack source!");
_Static_assert(sizeof(nox_chakram_attack_data_t) ==
	(sizeof(void*) == 4 ? 32 : 48), "wrong size of Chakram attack data!");

void nox_xxx_collideChakram_4EAF00(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

#endif // NOX_CHAKRAM_COLLIDE_4EAF00_H
