#include <assert.h>
#include <stddef.h>
#include <stdint.h>

#include "../poison_state_4ee7e0.h"

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;
static int32_t observed_first;
static int32_t observed_second;

int32_t nox_xxx_activatePoison_4EE7E0(nox_object_t* unit, int32_t increment, int32_t maximum) {
	observed_unit = unit;
	observed_first = increment;
	observed_second = maximum;
	return INT32_MIN;
}

void nox_xxx_updatePoison_4EE8F0(nox_object_t* unit, int32_t amount) {
	observed_unit = unit;
	observed_first = amount;
}

void nox_xxx_removePoison_4EE9D0(nox_object_t* unit) {
	observed_unit = unit;
}

void nox_xxx_setSomePoisonData_4EEA90(nox_object_t* unit, int32_t value) {
	observed_unit = unit;
	observed_first = value;
}

int main(void) {
	uintptr_t storage = UINTPTR_MAX;
	nox_object_t* const unit = (nox_object_t*)&storage;

	assert(nox_xxx_activatePoison_4EE7E0(unit, INT32_MIN, INT32_MAX) == INT32_MIN);
	assert(observed_unit == unit);
	assert(observed_first == INT32_MIN);
	assert(observed_second == INT32_MAX);

	nox_xxx_updatePoison_4EE8F0(unit, INT32_MIN);
	assert(observed_unit == unit);
	assert(observed_first == INT32_MIN);

	nox_xxx_setSomePoisonData_4EEA90(unit, INT32_MAX);
	assert(observed_unit == unit);
	assert(observed_first == INT32_MAX);

	nox_xxx_removePoison_4EE9D0(NULL);
	assert(observed_unit == NULL);
	return 0;
}
