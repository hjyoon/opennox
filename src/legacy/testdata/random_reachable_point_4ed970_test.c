#include <stddef.h>

#include "../random_reachable_point_4ed970.h"

struct float2 {
	float field_0;
	float field_4;
};

typedef float2* (*random_reachable_point_callback_t)(
	float,
	float2*,
	float2*);

_Static_assert(sizeof(float) == 4, "radius must remain binary32");
_Static_assert(sizeof(float2) == 8, "point must remain two binary32 values");
_Static_assert(offsetof(float2, field_0) == 0, "point X offset changed");
_Static_assert(offsetof(float2, field_4) == 4, "point Y offset changed");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&sub_4ED970, random_reachable_point_callback_t: 1, default: 0),
	"random reachable point must use one float and two native point pointers");

int main(void) {
	return 0;
}
