#ifndef NOX_XFER_MOVER_4F5730_H
#define NOX_XFER_MOVER_4F5730_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_mover_update_data_t {
	uint8_t field_0;
	uint8_t reserved_1[3];
	float field_1;
	int32_t field_2;
	uint32_t waypoint_3_pe32;
	uint32_t waypoint_3_index;
	uint32_t waypoint_5_pe32;
	uint32_t waypoint_5_index;
	uint32_t target_pe32;
	uint32_t target_extent;
} nox_mover_update_data_t;

_Static_assert(sizeof(nox_mover_update_data_t) == 36,
	"Mover update data must remain thirty-six bytes");
_Static_assert(sizeof(int32_t) == 4,
	"MoverXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerMover_4F5730(
	nox_object_t* object,
	void* context);

#endif // NOX_XFER_MOVER_4F5730_H
