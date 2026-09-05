#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_cast_by_user_4fdd20.h"

typedef int32_t (*cast_spell_by_user_fn)(
	int32_t,
	nox_object_t*,
	nox_spell_accept_arg_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "spell ID and result must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(offsetof(nox_spell_accept_arg_t, obj) == 0, "object must remain the first argument field");
_Static_assert(offsetof(nox_spell_accept_arg_t, x) == sizeof(void*), "X must follow the native pointer");
_Static_assert(
	offsetof(nox_spell_accept_arg_t, y) == sizeof(void*) + sizeof(float),
	"Y must follow X");
_Static_assert(
	sizeof(nox_spell_accept_arg_t) == sizeof(void*) + 2 * sizeof(float),
	"spell argument must preserve its native layout");
_Static_assert(
	_Generic(&nox_xxx_castSpellByUser_4FDD20, cast_spell_by_user_fn: 1, default: 0),
	"004FDD20 must preserve native pointers and signed dword spell/result values");

struct nox_object_t {
	uintptr_t marker;
};

static int32_t observed_spell_id;
static nox_object_t* observed_caster;
static nox_spell_accept_arg_t* observed_arg;
static int32_t next_result;

int32_t nox_xxx_castSpellByUser_4FDD20(
	int32_t spell_id,
	nox_object_t* caster,
	nox_spell_accept_arg_t* arg) {
	observed_spell_id = spell_id;
	observed_caster = caster;
	observed_arg = arg;
	return next_result;
}

int main(void) {
	nox_object_t caster = {.marker = UINTPTR_MAX};
	nox_object_t target = {.marker = UINTPTR_MAX - 1};
	nox_spell_accept_arg_t arg = {.obj = &target, .x = -123.5F, .y = 456.25F};
	cast_spell_by_user_fn const cast = nox_xxx_castSpellByUser_4FDD20;

	next_result = INT32_MIN;
	assert(cast(INT32_MAX, &caster, &arg) == INT32_MIN);
	assert(observed_spell_id == INT32_MAX);
	assert(observed_caster == &caster);
	assert(observed_arg == &arg);
	assert(observed_arg->obj == &target);
	assert(observed_arg->x == -123.5F);
	assert(observed_arg->y == 456.25F);

	next_result = INT32_MAX;
	assert(cast(INT32_MIN, NULL, NULL) == INT32_MAX);
	assert(observed_spell_id == INT32_MIN);
	assert(observed_caster == NULL);
	assert(observed_arg == NULL);
	return 0;
}
