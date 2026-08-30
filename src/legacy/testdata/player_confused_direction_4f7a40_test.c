#include "../player_confused_direction_4f7a40.h"

#include <stdint.h>

_Static_assert(
	_Generic(
		&nox_xxx_playerConfusedGetDirection_4F7A40,
		int (*)(nox_object_t*): 1,
		default: 0),
	"confused-direction native-pointer signature");

static nox_object_t* seen_unit;

int nox_xxx_playerConfusedGetDirection_4F7A40(nox_object_t* unit) {
	seen_unit = unit;
	return 0xa5;
}

int main(void) {
	uint64_t storage = UINT64_C(0);
	nox_object_t* unit = (nox_object_t*)(void*)&storage;
	if (sizeof(void*) == 8 && (uintptr_t)unit <= UINT32_MAX) {
		return 1;
	}
	if (nox_xxx_playerConfusedGetDirection_4F7A40(unit) != 0xa5) {
		return 2;
	}
	if (seen_unit != unit) {
		return 3;
	}
	return 0;
}
