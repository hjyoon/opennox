#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../lifetime_update_53b8f0.h"

typedef void (*lifetime_update_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_updateLifetime_53B8F0, lifetime_update_fn: 1, default: 0),
	"0053B8F0 must receive one native object pointer");
_Static_assert(sizeof(nox_lifetime_update_data_t) == sizeof(uint32_t),
	"0053B8F0 duration record must remain one uint32_t");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_source;

void nox_xxx_updateLifetime_53B8F0(nox_object_t* source) {
	observed_source = source;
}

int main(void) {
	nox_object_t source = {.marker = UINTPTR_MAX};
	lifetime_update_fn const update = nox_xxx_updateLifetime_53B8F0;

	update(&source);
	assert(observed_source == &source);
	assert(observed_source->marker == UINTPTR_MAX);
	update(NULL);
	assert(observed_source == NULL);
	return 0;
}
