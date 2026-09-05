#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_projectile_create_4fdda0.h"

typedef nox_object_t* (*create_spell_projectile_fn)(
	nox_object_t*,
	nox_object_t*,
	int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "spell ID must remain a signed dword");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_createSpellFly_4FDDA0, create_spell_projectile_fn: 1, default: 0),
	"004FDDA0 must preserve native object pointers and a signed dword spell ID");

struct nox_object_t {
	uintptr_t marker;
};

static nox_object_t* observed_source;
static nox_object_t* observed_target;
static int32_t observed_spell_id;
static nox_object_t* next_result;

nox_object_t* nox_xxx_createSpellFly_4FDDA0(
	nox_object_t* source,
	nox_object_t* target,
	int32_t spell_id) {
	observed_source = source;
	observed_target = target;
	observed_spell_id = spell_id;
	return next_result;
}

int main(void) {
	nox_object_t source = {.marker = UINTPTR_MAX};
	nox_object_t target = {.marker = UINTPTR_MAX - 1};
	nox_object_t projectile = {.marker = UINTPTR_MAX - 2};
	create_spell_projectile_fn const create = nox_xxx_createSpellFly_4FDDA0;

	next_result = &projectile;
	assert(create(&source, &target, INT32_MIN) == &projectile);
	assert(observed_source == &source);
	assert(observed_target == &target);
	assert(observed_spell_id == INT32_MIN);
	assert(observed_source->marker == UINTPTR_MAX);
	assert(observed_target->marker == UINTPTR_MAX - 1);

	next_result = NULL;
	assert(create(NULL, NULL, INT32_MAX) == NULL);
	assert(observed_source == NULL);
	assert(observed_target == NULL);
	assert(observed_spell_id == INT32_MAX);
	return 0;
}
