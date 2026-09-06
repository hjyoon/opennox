#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_duration_create_4feba0.h"

typedef int32_t (*spell_duration_create_fn)(
	int32_t,
	nox_object_t*,
	nox_object_t*,
	nox_object_t*,
	nox_spell_accept_arg_t*,
	int32_t,
	void*,
	void*,
	void*,
	int32_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4,
	"spell ID, level, duration, and result must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(offsetof(nox_spell_accept_arg_t, obj) == 0,
	"object must remain the first acceptance field");
_Static_assert(offsetof(nox_spell_accept_arg_t, x) == sizeof(void*),
	"acceptance X must follow the native pointer");
_Static_assert(
	offsetof(nox_spell_accept_arg_t, y) == sizeof(void*) + sizeof(float),
	"acceptance Y must follow X");
_Static_assert(
	sizeof(nox_spell_accept_arg_t) == sizeof(void*) + 2 * sizeof(float),
	"acceptance argument must preserve its native layout");
_Static_assert(
	_Generic(&nox_xxx_spellDurationBased_4FEBA0,
		spell_duration_create_fn: 1,
		default: 0),
	"004FEBA0 must preserve native pointers and signed dword values");

struct nox_object_t {
	uintptr_t marker;
};

static int32_t observed_spell_id;
static nox_object_t* observed_second;
static nox_object_t* observed_third;
static nox_object_t* observed_fourth;
static nox_spell_accept_arg_t* observed_arg;
static int32_t observed_level;
static void* observed_create;
static void* observed_update;
static void* observed_destroy;
static int32_t observed_duration;
static int32_t next_result;

int32_t nox_xxx_spellDurationBased_4FEBA0(
	int32_t spell_id,
	nox_object_t* second,
	nox_object_t* third,
	nox_object_t* fourth,
	nox_spell_accept_arg_t* arg,
	int32_t level,
	void* create,
	void* update,
	void* destroy,
	int32_t duration) {
	observed_spell_id = spell_id;
	observed_second = second;
	observed_third = third;
	observed_fourth = fourth;
	observed_arg = arg;
	observed_level = level;
	observed_create = create;
	observed_update = update;
	observed_destroy = destroy;
	observed_duration = duration;
	return next_result;
}

int main(void) {
	nox_object_t second = {.marker = UINTPTR_MAX};
	nox_object_t third = {.marker = UINTPTR_MAX - 1};
	nox_object_t fourth = {.marker = UINTPTR_MAX - 2};
	nox_object_t target = {.marker = UINTPTR_MAX - 3};
	nox_spell_accept_arg_t arg = {.obj = &target, .x = -123.5F, .y = 456.25F};
	uintptr_t create_marker = UINTPTR_MAX - 4;
	uintptr_t update_marker = UINTPTR_MAX - 5;
	uintptr_t destroy_marker = UINTPTR_MAX - 6;
	spell_duration_create_fn const create_spell = nox_xxx_spellDurationBased_4FEBA0;

	next_result = INT32_MIN;
	assert(create_spell(
			   INT32_MAX,
			   &second,
			   &third,
			   &fourth,
			   &arg,
			   INT32_MIN,
			   &create_marker,
			   &update_marker,
			   &destroy_marker,
			   INT32_MIN) == INT32_MIN);
	assert(observed_spell_id == INT32_MAX);
	assert(observed_second == &second);
	assert(observed_third == &third);
	assert(observed_fourth == &fourth);
	assert(observed_arg == &arg);
	assert(observed_arg->obj == &target);
	assert(observed_arg->x == -123.5F);
	assert(observed_arg->y == 456.25F);
	assert(observed_level == INT32_MIN);
	assert(observed_create == &create_marker);
	assert(observed_update == &update_marker);
	assert(observed_destroy == &destroy_marker);
	assert(observed_duration == INT32_MIN);

	next_result = INT32_MAX;
	assert(create_spell(
			   INT32_MIN,
			   NULL,
			   NULL,
			   NULL,
			   NULL,
			   INT32_MAX,
			   NULL,
			   NULL,
			   NULL,
			   INT32_MAX) == INT32_MAX);
	assert(observed_spell_id == INT32_MIN);
	assert(observed_second == NULL);
	assert(observed_third == NULL);
	assert(observed_fourth == NULL);
	assert(observed_arg == NULL);
	assert(observed_level == INT32_MAX);
	assert(observed_create == NULL);
	assert(observed_update == NULL);
	assert(observed_destroy == NULL);
	assert(observed_duration == INT32_MAX);
	return 0;
}
