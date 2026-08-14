#ifndef NOX_PENTAGRAM_COLLIDE_4EAB20_H
#define NOX_PENTAGRAM_COLLIDE_4EAB20_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

// This is the exact pointer-independent prefix touched by GAME.EXE 004EAB20,
// not a claim that the complete Pentagram update record is eight bytes.
typedef struct nox_pentagram_update_data_prefix_t {
	uint8_t reserved_0[4];
	uint32_t triggered;
} nox_pentagram_update_data_prefix_t;
_Static_assert(sizeof(nox_pentagram_update_data_prefix_t) == 8,
	"wrong size of Pentagram update-data prefix!");
_Static_assert(offsetof(nox_pentagram_update_data_prefix_t, triggered) == 4,
	"wrong offset of Pentagram triggered field!");

void nox_xxx_collidePentagram_4EAB20(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

#endif // NOX_PENTAGRAM_COLLIDE_4EAB20_H
