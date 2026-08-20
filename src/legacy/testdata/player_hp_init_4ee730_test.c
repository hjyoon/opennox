#include <assert.h>
#include <stddef.h>
#include <stdint.h>

#include "../player_hp_init_4ee730.h"

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;

void nox_xxx_playerHP_4EE730(nox_object_t* unit) {
	observed_unit = unit;
}

int main(void) {
	uintptr_t storage = UINTPTR_MAX;
	nox_object_t* const unit = (nox_object_t*)&storage;
	void (*const initialize)(nox_object_t*) = nox_xxx_playerHP_4EE730;

	initialize(unit);
	assert(observed_unit == unit);
	assert((uintptr_t)observed_unit == (uintptr_t)unit);

	initialize(NULL);
	assert(observed_unit == NULL);
	return 0;
}
