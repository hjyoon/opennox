#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../spell_duration_find_active_target_4ff2d0.h"

typedef void* (*spell_duration_find_active_target_fn)(int32_t, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "004FF2D0 spell ID must remain exact int32");
_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"004FF2D0 target and result must remain native-width");
_Static_assert(
	_Generic(&sub_4FF2D0, spell_duration_find_active_target_fn: 1, default: 0),
	"004FF2D0 must preserve its exact int32 and two native-pointer ABI");

struct nox_object_t {
	uintptr_t marker;
};

static unsigned char first_record;
static unsigned char second_record;
static int32_t observed_spell_id;
static nox_object_t* observed_target;
static unsigned int observed_calls;

void* sub_4FF2D0(int32_t spell_id, nox_object_t* target) {
	observed_spell_id = spell_id;
	observed_target = target;
	++observed_calls;
	if (spell_id == INT32_MIN && target != NULL) {
		return &first_record;
	}
	if (spell_id == INT32_MAX && target == NULL) {
		return &second_record;
	}
	return NULL;
}

int main(void) {
	nox_object_t target = {.marker = UINTPTR_MAX};
	spell_duration_find_active_target_fn const find = sub_4FF2D0;

	assert(find(INT32_MIN, &target) == &first_record);
	assert(observed_spell_id == INT32_MIN);
	assert(observed_target == &target);
	assert(find(INT32_MAX, NULL) == &second_record);
	assert(observed_spell_id == INT32_MAX);
	assert(observed_target == NULL);
	assert(find(INT32_C(-1), &target) == NULL);
	assert(observed_calls == 3);
	assert(target.marker == UINTPTR_MAX);
	return 0;
}
