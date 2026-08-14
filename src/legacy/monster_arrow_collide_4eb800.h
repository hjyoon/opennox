#ifndef NOX_MONSTER_ARROW_COLLIDE_4EB800_H
#define NOX_MONSTER_ARROW_COLLIDE_4EB800_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_monster_arrow_collide_data_t {
	int32_t coop_damage;
	int32_t other_damage;
} nox_monster_arrow_collide_data_t;

_Static_assert(sizeof(nox_monster_arrow_collide_data_t) == 8,
	"wrong size of MonsterArrow collide data!");
_Static_assert(offsetof(nox_monster_arrow_collide_data_t, coop_damage) == 0,
	"wrong offset of MonsterArrow Coop damage!");
_Static_assert(offsetof(nox_monster_arrow_collide_data_t, other_damage) == 4,
	"wrong offset of MonsterArrow other damage!");

void nox_xxx_collideMonsterArrow_4EB800(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);
int sub_536E80(char* args, nox_monster_arrow_collide_data_t* data);

#endif // NOX_MONSTER_ARROW_COLLIDE_4EB800_H
