#include "../defs.h"

#include <stddef.h>
#include <stdint.h>

#ifdef NOX_OBJECT_PIXIE_TARGET_CLEAR_4E81D0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable exactly the object and Pixie record fields used here.
#undef _Static_assert
_Static_assert(sizeof(((nox_object_t*)0)->typ_ind) == 2, "object type-index width");
_Static_assert(offsetof(nox_object_t, typ_ind) == (sizeof(void*) == 4 ? 4 : 8), "object type-index offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872), "object update-data offset");
_Static_assert(sizeof(nox_pixie_update_data_t) == (sizeof(void*) == 4 ? 28 : 40), "Pixie update-data size");
_Static_assert(offsetof(nox_pixie_update_data_t, owner) == 0, "Pixie owner offset");
_Static_assert(offsetof(nox_pixie_update_data_t, target) == (sizeof(void*) == 4 ? 4 : 8), "Pixie target offset");
_Static_assert(offsetof(nox_pixie_update_data_t, deadline) == (sizeof(void*) == 4 ? 20 : 28), "Pixie deadline offset");
_Static_assert(offsetof(nox_pixie_update_data_t, last_owner_visible_frame) == (sizeof(void*) == 4 ? 24 : 32),
	"Pixie last-owner-visible offset");
#endif

static void object_pixie_target_clear_fixture_4e81d0(nox_object_t* obj, uint32_t pixie_type) {
	if (obj && obj->typ_ind == pixie_type) {
		((nox_pixie_update_data_t*)obj->data_update)->target = NULL;
	}
}

int main(void) {
	nox_object_t object = {0};
	nox_object_t owner = {0};
	nox_object_t target = {0};
	nox_pixie_update_data_t update = {.owner = &owner, .target = &target, .deadline = UINT32_C(17)};
	object.typ_ind = UINT16_C(37);
	object.data_update = &update;
	object_pixie_target_clear_fixture_4e81d0(&object, UINT32_C(37));
	return update.target != NULL || update.owner != &owner || update.deadline != UINT32_C(17);
}
