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

// The paired Pentagram pointer widens naturally on 64-bit targets. The fields
// before it retain their GAME.EXE offsets, while the animation tail moves by
// one pointer-width delta.
typedef struct nox_pentagram_update_data_t {
	uint8_t state;
	uint8_t reserved_1[3];
	uint32_t triggered;
	uint8_t animation_frame;
	uint8_t animation_tick;
	uint8_t reserved_10[2];
	nox_object_t* destination;
	uint8_t reserved_after_destination[4];
	uint8_t animation_step;
	uint8_t reserved_after_step[3];
} nox_pentagram_update_data_t;
_Static_assert(sizeof(nox_pentagram_update_data_t) == (sizeof(void*) == 4 ? 24 : 32),
	"wrong native size of Pentagram update data!");
_Static_assert(offsetof(nox_pentagram_update_data_t, triggered) == 4,
	"wrong offset of Pentagram update trigger!");
_Static_assert(offsetof(nox_pentagram_update_data_t, animation_frame) == 8,
	"wrong offset of Pentagram animation frame!");
_Static_assert(offsetof(nox_pentagram_update_data_t, destination) ==
	(sizeof(void*) == 4 ? 12 : 16), "wrong native offset of Pentagram destination!");
_Static_assert(offsetof(nox_pentagram_update_data_t, animation_step) ==
	(sizeof(void*) == 4 ? 20 : 28), "wrong native offset of Pentagram animation step!");

void nox_xxx_collidePentagram_4EAB20(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

#endif // NOX_PENTAGRAM_COLLIDE_4EAB20_H
