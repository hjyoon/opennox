#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../unit_follow_5158c0.h"

typedef void (*unit_follow_fn)(nox_object_t*, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_unitSetFollow_5158C0, unit_follow_fn: 1, default: 0),
	"005158C0 must receive two native object pointers");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_unit;
static nox_object_t* observed_target;

void nox_xxx_unitSetFollow_5158C0(nox_object_t* unit, nox_object_t* target) {
	observed_unit = unit;
	observed_target = target;
}

static void expect(nox_object_t* unit, nox_object_t* target) {
	assert(observed_unit == unit);
	assert(observed_target == target);
}

int main(void) {
	nox_object_t unit = {.marker = UINTPTR_MAX};
	nox_object_t target = {.marker = UINTPTR_MAX - 1};
	unit_follow_fn const follow = nox_xxx_unitSetFollow_5158C0;

	follow(&unit, &target);
	expect(&unit, &target);
	assert(observed_unit->marker == UINTPTR_MAX);
	assert(observed_target->marker == UINTPTR_MAX - 1);
	follow(NULL, &target);
	expect(NULL, &target);
	follow(&unit, NULL);
	expect(&unit, NULL);
	follow(NULL, NULL);
	expect(NULL, NULL);
	return 0;
}
