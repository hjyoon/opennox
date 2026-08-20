#include <stddef.h>

#include "../chest_open_4edf00.h"

typedef void (*chest_open_callback_t)(
	nox_object_t*,
	nox_object_t*);

_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_chest_4EDF00, chest_open_callback_t: 1, default: 0),
	"ChestOpen must use two native object pointers and return void");

int main(void) {
	return 0;
}
