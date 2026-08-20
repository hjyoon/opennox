#include <stddef.h>

#include "../potion_drop_4edde0.h"

struct float2 {
	float x;
	float y;
};

typedef int (*potion_drop_callback_t)(
	nox_object_t*,
	nox_object_t*,
	float2*);

_Static_assert(sizeof(int) == 4, "PotionDrop result must remain 32-bit");
_Static_assert(sizeof(float2) == 8, "float2 must contain two binary32 coordinates");
_Static_assert(offsetof(float2, x) == 0, "float2 X offset");
_Static_assert(offsetof(float2, y) == 4, "float2 Y offset");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&sub_4EDDE0, potion_drop_callback_t: 1, default: 0),
	"PotionDrop must use two native object pointers and one point pointer");

int main(void) {
	return 0;
}
