#if defined(NOX_ABI_FREESTANDING)
#define assert(expr) ((void)sizeof(expr))
#else
#include <assert.h>
#endif
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../summon_unit_5016c0.h"

typedef nox_object_t* (*summon_unit_fn)(
	int32_t,
	float*,
	nox_object_t*,
	uint8_t);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4,
	"005016C0 type IDs must remain signed dwords");
_Static_assert(sizeof(float) == 4,
	"005016C0 positions must contain two 32-bit floats");
_Static_assert(sizeof(uint8_t) == 1,
	"005016C0 directions must remain unsigned bytes");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_unitDoSummonAt_5016C0, summon_unit_fn: 1, default: 0),
	"005016C0 must preserve int32/float*/object*/uint8 ABI");

struct nox_object_t {
	uintptr_t marker;
};

static int32_t observed_type_id;
static float* observed_position;
static nox_object_t* observed_owner;
static uint8_t observed_direction;
static nox_object_t* summon_result;

nox_object_t* nox_xxx_unitDoSummonAt_5016C0(
	int32_t type_id,
	float* position,
	nox_object_t* owner,
	uint8_t direction) {
	observed_type_id = type_id;
	observed_position = position;
	observed_owner = owner;
	observed_direction = direction;
	return summon_result;
}

int main(void) {
	float position[2] = {12.5f, -7.25f};
	nox_object_t owner = {.marker = UINTPTR_MAX};
	nox_object_t created = {.marker = UINTPTR_MAX - 1};
	summon_unit_fn const summon = nox_xxx_unitDoSummonAt_5016C0;

	summon_result = &created;
	assert(summon(INT32_MIN, position, &owner, UINT8_MAX) == &created);
	assert(observed_type_id == INT32_MIN);
	assert(observed_position == position);
	assert(observed_position[0] == 12.5f);
	assert(observed_position[1] == -7.25f);
	assert(observed_owner == &owner);
	assert(observed_owner->marker == UINTPTR_MAX);
	assert(observed_direction == UINT8_MAX);

	summon_result = NULL;
	assert(summon(INT32_MAX, NULL, NULL, 0) == NULL);
	assert(observed_type_id == INT32_MAX);
	assert(observed_position == NULL);
	assert(observed_owner == NULL);
	assert(observed_direction == 0);
	return 0;
}
