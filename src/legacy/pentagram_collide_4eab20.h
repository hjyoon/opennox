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

// PentagramUpdate is also serialized by TransporterXfer. Keep both PE32
// dwords fixed: destination_pe32 is transient, destination_extent is the map
// link written on load, and native code widens the resolved object in a Go
// sidecar instead of overwriting either field.
typedef struct nox_pentagram_update_data_t {
	uint8_t state;
	uint8_t reserved_1[3];
	uint32_t triggered;
	uint8_t animation_frame;
	uint8_t animation_tick;
	uint8_t reserved_10[2];
	uint32_t destination_pe32;
	uint32_t destination_extent;
	uint8_t animation_step;
	uint8_t reserved_after_step[3];
} nox_pentagram_update_data_t;
_Static_assert(sizeof(nox_pentagram_update_data_t) == 24,
	"wrong fixed size of Pentagram update data!");
_Static_assert(offsetof(nox_pentagram_update_data_t, triggered) == 4,
	"wrong offset of Pentagram update trigger!");
_Static_assert(offsetof(nox_pentagram_update_data_t, animation_frame) == 8,
	"wrong offset of Pentagram animation frame!");
_Static_assert(offsetof(nox_pentagram_update_data_t, destination_pe32) == 12,
	"wrong fixed offset of Pentagram PE32 destination!");
_Static_assert(offsetof(nox_pentagram_update_data_t, destination_extent) == 16,
	"wrong fixed offset of Pentagram destination extent!");
_Static_assert(offsetof(nox_pentagram_update_data_t, animation_step) == 20,
	"wrong fixed offset of Pentagram animation step!");

void nox_xxx_collidePentagram_4EAB20(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);
int nox_xxx_updateTeleportPentagram_53BEF0(nox_object_t* pentagram);
void nox_xxx_fnPentagramTeleport_53C060(nox_object_t* unit, void* destination);
int nox_xxx_updateInvisiblePentagram_53C0C0(nox_object_t* pentagram);
void sub_53C140(nox_object_t* unit, void* destination);

#endif // NOX_PENTAGRAM_COLLIDE_4EAB20_H
