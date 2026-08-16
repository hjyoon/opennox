#include <stddef.h>

#include "../treasure_drop_4ed710.h"

struct float2 {
	float x;
	float y;
};

typedef int (*treasure_drop_callback_t)(
	nox_object_t*,
	nox_object_t*,
	float2*);

_Static_assert(sizeof(int) == 4, "TreasureDrop result must remain 32-bit");
_Static_assert(sizeof(float2) == 8, "float2 must contain two binary32 coordinates");
_Static_assert(offsetof(float2, x) == 0, "float2 X offset");
_Static_assert(offsetof(float2, y) == 4, "float2 Y offset");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_dropTreasure_4ED710, treasure_drop_callback_t: 1, default: 0),
	"TreasureDrop must use two native object pointers and one point pointer");

int main(void) {
	return 0;
}
