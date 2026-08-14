#ifndef NOX_MANA_DRAIN_COLLIDE_4E9490_H
#define NOX_MANA_DRAIN_COLLIDE_4E9490_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_mana_drain_collide_data_t {
	uint8_t amount;
	uint8_t reserved[7];
} nox_mana_drain_collide_data_t;
_Static_assert(sizeof(nox_mana_drain_collide_data_t) == 8,
	"wrong size of ManaDrainCollide data structure!");
_Static_assert(offsetof(nox_mana_drain_collide_data_t, amount) == 0,
	"wrong offset of ManaDrainCollide amount!");

void nox_xxx_collideManadrain_4E9490(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);
int nox_xxx_collideManaDrainLoad_536E50(
	char* args,
	nox_mana_drain_collide_data_t* data);

#endif // NOX_MANA_DRAIN_COLLIDE_4E9490_H
