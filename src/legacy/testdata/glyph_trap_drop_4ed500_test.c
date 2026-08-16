#include <stddef.h>

#include "../glyph_drop_4ed500.h"
#include "../trap_drop_4ed580.h"

struct float2 {
	float x;
	float y;
};

typedef int (*drop_callback_t)(nox_object_t*, nox_object_t*, float2*);

_Static_assert(sizeof(float2) == 8, "float2 must contain two binary32 coordinates");
_Static_assert(offsetof(float2, x) == 0, "float2 X offset");
_Static_assert(offsetof(float2, y) == 4, "float2 Y offset");
_Static_assert(
	_Generic(&nox_GlyphDrop_4ED500, drop_callback_t: 1, default: 0),
	"GlyphDrop must use two native object pointers and one point pointer");
_Static_assert(
	_Generic(&nox_xxx_dropTrap_4ED580, drop_callback_t: 1, default: 0),
	"TrapDrop must use two native object pointers and one point pointer");

int main(void) {
	return 0;
}
