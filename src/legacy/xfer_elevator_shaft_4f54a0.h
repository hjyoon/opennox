#ifndef NOX_XFER_ELEVATOR_SHAFT_4F54A0_H
#define NOX_XFER_ELEVATOR_SHAFT_4F54A0_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_elevator_shaft_update_data_t {
	uint32_t field_0;
	uint32_t link_pe32;
	uint32_t elevator_extent;
	uint8_t field_3;
	uint8_t reserved_13[3];
} nox_elevator_shaft_update_data_t;

_Static_assert(sizeof(nox_elevator_shaft_update_data_t) == 16,
	"ElevatorShaft update data must remain sixteen bytes");
_Static_assert(sizeof(int32_t) == 4,
	"ElevatorShaftXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerElevatorShaft_4F54A0(
	nox_object_t* object,
	void* context);

#endif // NOX_XFER_ELEVATOR_SHAFT_4F54A0_H
