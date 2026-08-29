// Compile-only native-width contract for the Transporter transfer restored
// from GAME.EXE 004F5300.
#define _Static_assert(...) extern int nox_suppressed_static_assert
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_transporter_update_data_t) == 20,
	"Transporter update-data size");
_Static_assert(offsetof(nox_transporter_update_data_t, target_pe32) == 12,
	"Transporter PE32 target offset");
_Static_assert(offsetof(nox_transporter_update_data_t, target_extent) == 16,
	"Transporter target-extent offset");
_Static_assert(offsetof(nox_object_t, field_34) == (sizeof(void*) == 4 ? 136 : 140),
	"object field-34 offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");

typedef int32_t (*transporter_xfer_fn_4F5300)(nox_object_t*, void*);

_Static_assert(
	_Generic(&nox_xxx_XFerTransporter_4F5300,
		transporter_xfer_fn_4F5300: 1, default: 0),
	"TransporterXfer signature");

static transporter_xfer_fn_4F5300 const transporter_xfer_signature =
	nox_xxx_XFerTransporter_4F5300;

int main(void) {
	return transporter_xfer_signature == NULL;
}
