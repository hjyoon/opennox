#include <assert.h>
#include <limits.h>
#include <stdint.h>

#include "../GAME4_3.h"

typedef int32_t (*read_use_fn)(nox_object_t*, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "ReadUse result must remain exact int32");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_useRead_53F7C0, read_use_fn: 1, default: 0),
	"ReadUse must preserve both native object pointers");

static nox_object_t* observed_owner;
static nox_object_t* observed_readable;
static unsigned int observed_calls;

int32_t nox_xxx_useRead_53F7C0(
		nox_object_t* owner, nox_object_t* readable) {
	observed_owner = owner;
	observed_readable = readable;
	++observed_calls;
	return INT32_MIN;
}

int main(void) {
	nox_object_t owner = {0};
	nox_object_t readable = {0};
	read_use_fn const use = nox_xxx_useRead_53F7C0;

	assert(use(&owner, &readable) == INT32_MIN);
	assert(observed_owner == &owner);
	assert(observed_readable == &readable);
	assert(use(NULL, NULL) == INT32_MIN);
	assert(observed_owner == NULL);
	assert(observed_readable == NULL);
	assert(observed_calls == 2);
	return 0;
}
