#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_accept_4fd400.h"

typedef int32_t (*spell_accept_fn)(
	int32_t,
	nox_object_t*,
	nox_object_t*,
	nox_object_t*,
	nox_spell_accept_arg_t*,
	int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "spell ID, level, and result must remain signed dwords");
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
	_Generic(&nox_xxx_spellAccept_4FD400, spell_accept_fn: 1, default: 0),
	"004FD400 must preserve native pointers and signed dword arguments/result");

struct nox_object_t {
	uintptr_t marker;
};

static int32_t observed_spell_id;
static nox_object_t* observed_second;
static nox_object_t* observed_third;
static nox_object_t* observed_fourth;
static nox_spell_accept_arg_t* observed_arg;
static int32_t observed_level;
static int32_t next_result;

int32_t nox_xxx_spellAccept_4FD400(
	int32_t spell_id,
	nox_object_t* second,
	nox_object_t* third,
	nox_object_t* fourth,
	nox_spell_accept_arg_t* arg,
	int32_t level) {
	observed_spell_id = spell_id;
	observed_second = second;
	observed_third = third;
	observed_fourth = fourth;
	observed_arg = arg;
	observed_level = level;
	return next_result;
}

int main(void) {
	nox_object_t second = {.marker = UINTPTR_MAX};
	nox_object_t third = {.marker = UINTPTR_MAX - 1};
	nox_object_t fourth = {.marker = UINTPTR_MAX - 2};
	nox_object_t target = {.marker = UINTPTR_MAX - 3};
	nox_spell_accept_arg_t arg = {.obj = &target, .x = -123.5F, .y = 456.25F};
	spell_accept_fn const accept = nox_xxx_spellAccept_4FD400;

	next_result = INT32_MIN;
	assert(accept(INT32_MAX, &second, &third, &fourth, &arg, INT32_MIN) == INT32_MIN);
	assert(observed_spell_id == INT32_MAX);
	assert(observed_second == &second);
	assert(observed_third == &third);
	assert(observed_fourth == &fourth);
	assert(observed_arg == &arg);
	assert(observed_arg->obj == &target);
	assert(observed_arg->x == -123.5F);
	assert(observed_arg->y == 456.25F);
	assert(observed_level == INT32_MIN);

	next_result = INT32_MAX;
	assert(accept(INT32_MIN, NULL, NULL, NULL, NULL, INT32_MAX) == INT32_MAX);
	assert(observed_spell_id == INT32_MIN);
	assert(observed_second == NULL);
	assert(observed_third == NULL);
	assert(observed_fourth == NULL);
	assert(observed_arg == NULL);
	assert(observed_level == INT32_MAX);
	return 0;
}
