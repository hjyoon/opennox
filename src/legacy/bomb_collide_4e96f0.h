#ifndef NOX_BOMB_COLLIDE_4E96F0_H
#define NOX_BOMB_COLLIDE_4E96F0_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_bomb_collide_data_t {
	uint8_t reserved[8];
} nox_bomb_collide_data_t;
_Static_assert(sizeof(nox_bomb_collide_data_t) == 8,
	"wrong size of BombCollide data structure!");
_Static_assert(offsetof(nox_bomb_collide_data_t, reserved) == 0,
	"wrong offset of BombCollide reserved data!");

void nox_xxx_collideBomb_4E96F0(
	nox_object_t* bomb,
	nox_object_t* other,
	float* collision);

#endif // NOX_BOMB_COLLIDE_4E96F0_H
