#include <stddef.h>

#include "../object_drop_bounded_4ed810.h"

struct float2 {
	float x;
	float y;
};

typedef int (*object_drop_bounded_callback_t)(
	nox_object_t*,
	nox_object_t*,
	float2*);

_Static_assert(sizeof(int) == 4, "bounded drop result must remain 32-bit");
_Static_assert(sizeof(float2) == 8, "float2 must contain two binary32 coordinates");
_Static_assert(offsetof(float2, x) == 0, "float2 X offset");
_Static_assert(offsetof(float2, y) == 4, "float2 Y offset");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_drop_4ED810, object_drop_bounded_callback_t: 1, default: 0),
	"bounded drop must use two native object pointers and one point pointer");

int main(void) {
	return 0;
}
