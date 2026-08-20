#include <stddef.h>
#include <stdint.h>

#include "../drop_all_items_4eda40.h"

typedef int32_t (*drop_all_items_callback_t)(nox_object_t*);

_Static_assert(sizeof(int32_t) == 4, "result must remain a full EAX value");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_dropAllItems_4EDA40, drop_all_items_callback_t: 1, default: 0),
	"drop-all-items must use one native object pointer and a fixed-width result");

int main(void) {
	return 0;
}
