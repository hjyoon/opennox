// Keep this fixture independent from the Win32-only aggregate legacy headers
// so the retained ABI can be compiled by every supported target frontend.
#include "../skeleton_init_4f0370.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

typedef void (*skeleton_init_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_unitSkeletonInit_4F0370, skeleton_init_fn: 1, default: 0),
	"SkeletonInit must use one native object pointer and return void");

static nox_object_t* observed_unit;
static unsigned int observed_calls;

void nox_xxx_unitSkeletonInit_4F0370(nox_object_t* unit) {
	observed_unit = unit;
	++observed_calls;
}

static int check_call(skeleton_init_fn init, nox_object_t* unit) {
	init(unit);
	if (observed_unit != unit)
		return __LINE__;
	return 0;
}

int main(void) {
	unsigned char first_storage = 0;
	unsigned char second_storage = 0;
	nox_object_t* const first = (nox_object_t*)(void*)&first_storage;
	nox_object_t* const second = (nox_object_t*)(void*)&second_storage;
	skeleton_init_fn const init = nox_xxx_unitSkeletonInit_4F0370;
	int line;

	line = check_call(init, first);
	if (line != 0)
		return line;
	line = check_call(init, second);
	if (line != 0)
		return line;
	line = check_call(init, NULL);
	if (line != 0)
		return line;
	line = check_call(init, first);
	if (line != 0)
		return line;
	if (observed_calls != 4)
		return __LINE__;
	return 0;
}
