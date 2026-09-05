#include <assert.h>
#include <float.h>
#include <limits.h>
#include <stdint.h>
#include <string.h>

#include "../spell_gesture_cancel_4fe680.h"

typedef void (*spell_gesture_cancel_fn)(nox_object_t*, float);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(
	sizeof(float) == 4 && FLT_RADIX == 2 && FLT_MANT_DIG == 24 && FLT_MAX_EXP == 128,
	"radius must remain an IEEE binary32 argument");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(
	_Generic(&nox_xxx_spell_4FE680, spell_gesture_cancel_fn: 1, default: 0),
	"004FE680 must preserve one native object pointer and one binary32 argument");

struct nox_object_t {
	uint32_t marker;
};

static nox_object_t* observed_source;
static uint32_t observed_radius_bits;

void nox_xxx_spell_4FE680(nox_object_t* source, float radius) {
	observed_source = source;
	memcpy(&observed_radius_bits, &radius, sizeof(observed_radius_bits));
}

int main(void) {
	nox_object_t source = {0};
	uint32_t const radius_bits = UINT32_C(0xc3960001);
	float radius;
	memcpy(&radius, &radius_bits, sizeof(radius));
	spell_gesture_cancel_fn const cancel = nox_xxx_spell_4FE680;

	cancel(&source, radius);
	assert(observed_source == &source);
	assert(observed_radius_bits == radius_bits);
	return 0;
}
