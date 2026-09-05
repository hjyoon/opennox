#include <assert.h>
#include <float.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>

#include "../position_delta_4fea70.h"

typedef int32_t (*position_delta_fn)(nox_object_t*, float2*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "int32_t must remain four bytes");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(float) == 4 && FLT_RADIX == 2 && FLT_MANT_DIG == 24 && FLT_MAX_EXP == 128,
	"coordinates require IEEE-754 binary32");
_Static_assert(
	_Generic(&sub_4FEA70, position_delta_fn: 1, default: 0),
	"004FEA70 must receive native object and point pointers and return int32_t");

struct nox_object_t {
	uintptr_t marker;
};

struct float2 {
	float field_0;
	float field_4;
};

_Static_assert(sizeof(float2) == 8, "float2 must remain eight bytes");
_Static_assert(offsetof(float2, field_0) == 0, "float2 X moved");
_Static_assert(offsetof(float2, field_4) == 4, "float2 Y moved");

static nox_object_t* observed_object;
static float2* observed_point;
static uint32_t observed_x_bits;
static uint32_t observed_y_bits;

int32_t sub_4FEA70(nox_object_t* object, float2* point) {
	observed_object = object;
	observed_point = point;
	memcpy(&observed_x_bits, &point->field_0, sizeof(observed_x_bits));
	memcpy(&observed_y_bits, &point->field_4, sizeof(observed_y_bits));
	return INT32_C(0x1234567);
}

int main(void) {
	nox_object_t object = {.marker = UINTPTR_MAX};
	float2 point;
	uint32_t const x_bits = UINT32_C(0x41200001);
	uint32_t const y_bits = UINT32_C(0xc1700001);
	memcpy(&point.field_0, &x_bits, sizeof(x_bits));
	memcpy(&point.field_4, &y_bits, sizeof(y_bits));
	position_delta_fn const predicate = sub_4FEA70;

	assert(predicate(&object, &point) == INT32_C(0x1234567));
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(observed_point == &point);
	assert(observed_x_bits == x_bits);
	assert(observed_y_bits == y_bits);
	return 0;
}
