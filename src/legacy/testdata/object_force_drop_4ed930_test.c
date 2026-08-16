#include <stddef.h>

#include "../object_force_drop_4ed930.h"

typedef int (*object_force_drop_callback_t)(
	nox_object_t*,
	nox_object_t*);

_Static_assert(sizeof(int) == 4, "force-drop result must remain 32-bit");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_invForceDropItem_4ED930, object_force_drop_callback_t: 1, default: 0),
	"force-drop must use two native object pointers");

int main(void) {
	return 0;
}
