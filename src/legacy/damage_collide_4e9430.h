#ifndef NOX_DAMAGE_COLLIDE_4E9430_H
#define NOX_DAMAGE_COLLIDE_4E9430_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_damage_collide_data_t {
	uint8_t damage;
	uint8_t reserved[3];
	int32_t damage_type;
} nox_damage_collide_data_t;
_Static_assert(sizeof(nox_damage_collide_data_t) == 8,
	"wrong size of DamageCollide data structure!");
_Static_assert(offsetof(nox_damage_collide_data_t, damage) == 0,
	"wrong offset of DamageCollide damage!");
_Static_assert(offsetof(nox_damage_collide_data_t, damage_type) == 4,
	"wrong offset of DamageCollide damage type!");

void nox_xxx_collideDamage_4E9430(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);
int nox_xxx_collideDamageLoad_536E10(char* args, nox_damage_collide_data_t* data);

#endif // NOX_DAMAGE_COLLIDE_4E9430_H
