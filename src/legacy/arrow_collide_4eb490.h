#ifndef NOX_ARROW_COLLIDE_4EB490_H
#define NOX_ARROW_COLLIDE_4EB490_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_arrow_collide_data_t {
	uint32_t field_0;
	nox_object_t* owner;
} nox_arrow_collide_data_t;

typedef struct nox_arrow_attack_data_t {
	float damage;
	uint8_t damage_type;
	uint8_t padding_5[3];
	float radius;
	nox_object_t* owner;
	float x;
	float y;
	uint32_t field_24;
	nox_object_t* source;
} nox_arrow_attack_data_t;

_Static_assert(offsetof(nox_arrow_collide_data_t, field_0) == 0,
	"wrong offset of Arrow collide field zero!");
_Static_assert(offsetof(nox_arrow_collide_data_t, owner) == sizeof(void*),
	"wrong offset of Arrow collide owner!");
_Static_assert(sizeof(nox_arrow_collide_data_t) == 2 * sizeof(void*),
	"wrong size of Arrow collide data!");

_Static_assert(offsetof(nox_arrow_attack_data_t, damage) == 0,
	"wrong offset of Arrow attack damage!");
_Static_assert(offsetof(nox_arrow_attack_data_t, damage_type) == 4,
	"wrong offset of Arrow attack damage type!");
_Static_assert(offsetof(nox_arrow_attack_data_t, radius) == 8,
	"wrong offset of Arrow attack radius!");
_Static_assert(offsetof(nox_arrow_attack_data_t, owner) ==
	(sizeof(void*) == 4 ? 12 : 16), "wrong offset of Arrow attack owner!");
_Static_assert(offsetof(nox_arrow_attack_data_t, x) ==
	(sizeof(void*) == 4 ? 16 : 24), "wrong offset of Arrow attack X!");
_Static_assert(offsetof(nox_arrow_attack_data_t, y) ==
	(sizeof(void*) == 4 ? 20 : 28), "wrong offset of Arrow attack Y!");
_Static_assert(offsetof(nox_arrow_attack_data_t, field_24) ==
	(sizeof(void*) == 4 ? 24 : 32), "wrong offset of Arrow attack field 24!");
_Static_assert(offsetof(nox_arrow_attack_data_t, source) ==
	(sizeof(void*) == 4 ? 28 : 40), "wrong offset of Arrow attack source!");
_Static_assert(sizeof(nox_arrow_attack_data_t) ==
	(sizeof(void*) == 4 ? 32 : 48), "wrong size of Arrow attack data!");

void nox_xxx_collideArrow_4EB490(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

void nox_server_arrowCollideDataSetOwner_4EB490(
	nox_object_t* source,
	nox_object_t* owner);

#endif // NOX_ARROW_COLLIDE_4EB490_H
