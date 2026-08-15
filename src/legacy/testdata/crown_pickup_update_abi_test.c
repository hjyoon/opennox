#include <stddef.h>
#include <stdint.h>

#include "../crown_pickup_4f3400.h"
#include "../crown_update_53e1d0.h"

static uint32_t (*const crown_pickup_signature_4f3400)(
	nox_object_t*, nox_object_t*, int32_t, int32_t) = sub_4F3400;
static void (*const crown_update_signature_53e1d0)(nox_object_t*) =
	nox_xxx_updateCrown_53E1D0;
static void (*const crown_target_setter_signature_53e1d0)(
	nox_object_t*, nox_object_t*) =
	nox_server_crownUpdateDataSetPickupTarget_53E1D0;

_Static_assert(offsetof(nox_crown_update_data_t, field_0) == 0,
	"Crown field 0 moved");
_Static_assert(offsetof(nox_crown_update_data_t, pickup_target) == sizeof(void*),
	"Crown pickup target is not pointer-native");
_Static_assert(offsetof(nox_crown_update_data_t, field_2) == 2 * sizeof(void*),
	"Crown field 2 moved");
_Static_assert(sizeof(nox_crown_update_data_t) == (sizeof(void*) == 4 ? 12 : 24),
	"Crown update record size is not architecture-correct");

int crown_pickup_update_abi_test(void) {
	return crown_pickup_signature_4f3400 != 0 &&
		crown_update_signature_53e1d0 != 0 &&
		crown_target_setter_signature_53e1d0 != 0;
}
