#ifndef NOX_MONSTER_GENERATOR_COLLIDE_4EBE10_H
#define NOX_MONSTER_GENERATOR_COLLIDE_4EBE10_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

// Native-pointer representation of the 164-byte GAME.EXE update record. The
// first twelve object references widen together; the six words for the first
// three callbacks and the collision pair remain fixed-width, so the collision
// block moves as a unit.
typedef struct nox_monster_generator_update_data_t {
	nox_object_t* objects[12];
	uint32_t pre_collision[6];
	uint32_t collision_flags;
	int32_t collision_func;
	uint32_t tail[21];
} nox_monster_generator_update_data_t;

_Static_assert(offsetof(nox_monster_generator_update_data_t, collision_flags) ==
	(sizeof(void*) == 4 ? 72 : 120),
	"wrong offset of MonsterGenerator collision flags!");
_Static_assert(offsetof(nox_monster_generator_update_data_t, collision_func) ==
	(sizeof(void*) == 4 ? 76 : 124),
	"wrong offset of MonsterGenerator collision function!");
_Static_assert(sizeof(nox_monster_generator_update_data_t) ==
	(sizeof(void*) == 4 ? 164 : 216),
	"wrong size of MonsterGenerator update data!");

void nox_xxx_collideMonsterGen_4EBE10(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

#endif // NOX_MONSTER_GENERATOR_COLLIDE_4EBE10_H
