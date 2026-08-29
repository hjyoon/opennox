#ifndef NOX_XFER_TRANSPORTER_4F5300_H
#define NOX_XFER_TRANSPORTER_4F5300_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_transporter_update_data_t {
	uint8_t reserved_0[12];
	uint32_t target_pe32;
	uint32_t target_extent;
} nox_transporter_update_data_t;

_Static_assert(sizeof(nox_transporter_update_data_t) == 20,
	"Transporter update data must remain twenty bytes");
_Static_assert(sizeof(int32_t) == 4,
	"TransporterXfer result must remain an exact 32-bit value");

int32_t nox_xxx_XFerTransporter_4F5300(
	nox_object_t* object,
	void* context);

#endif // NOX_XFER_TRANSPORTER_4F5300_H
