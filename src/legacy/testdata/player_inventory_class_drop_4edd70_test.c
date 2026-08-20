#include <stddef.h>

#include "../player_inventory_class_drop_4edd70.h"

typedef int (*player_inventory_class_drop_callback_t)(void);

_Static_assert(sizeof(int) == 4, "result must remain a full EAX value");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&sub_4EDD70, player_inventory_class_drop_callback_t: 1, default: 0),
	"player inventory class drop must keep its no-argument int ABI");

int main(void) {
	return 0;
}
