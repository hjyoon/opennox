#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../player_mana_refresh_4eecf0.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef uintptr_t (*player_mana_refresh_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "mana bytes must remain eight bits");
_Static_assert(sizeof(uintptr_t) == sizeof(void*), "return register must preserve native pointers");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_playerManaRefresh_4EECF0, player_mana_refresh_fn: 1, default: 0),
	"player mana refresh must use a native object pointer and native return register");

static nox_object_t* observed_unit;
static uintptr_t next_result;

uintptr_t nox_xxx_playerManaRefresh_4EECF0(nox_object_t* unit) {
	observed_unit = unit;
	return next_result;
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	player_mana_refresh_fn const refresh = nox_xxx_playerManaRefresh_4EECF0;

	next_result = UINTPTR_MAX;
	assert(refresh(&unit) == UINTPTR_MAX);
	assert(observed_unit == &unit);

	next_result = (uintptr_t)0x1234;
	assert(refresh(NULL) == (uintptr_t)0x1234);
	assert(observed_unit == NULL);
	return 0;
}
