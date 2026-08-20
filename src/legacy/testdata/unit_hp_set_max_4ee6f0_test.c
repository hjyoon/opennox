#include <assert.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_hp_set_max_4ee6f0.h"

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;

void nox_xxx_unitHPsetOnMax_4EE6F0(nox_object_t* unit) {
	observed_unit = unit;
}

int main(void) {
	uintptr_t storage = UINTPTR_MAX;
	nox_object_t* const unit = (nox_object_t*)&storage;
	void (*const restore)(nox_object_t*) = nox_xxx_unitHPsetOnMax_4EE6F0;

	restore(unit);
	assert(observed_unit == unit);
	assert((uintptr_t)observed_unit == (uintptr_t)unit);

	restore(NULL);
	assert(observed_unit == NULL);
	return 0;
}
