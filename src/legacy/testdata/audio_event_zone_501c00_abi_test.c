#if defined(NOX_ABI_FREESTANDING)
#define assert(expr) ((void)sizeof(expr))
#else
#include <assert.h>
#endif
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../audio_event_zone_501c00.h"

typedef uint8_t (*audio_event_zone_fn)(float*, nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(
	_Generic(&sub_501C00, audio_event_zone_fn: 1, default: 0),
	"00501C00 must receive native pointers and return an unsigned byte");

struct nox_object_t {
	uintptr_t marker;
};

static float* observed_position;
static nox_object_t* observed_object;

uint8_t sub_501C00(float* position, nox_object_t* object) {
	observed_position = position;
	observed_object = object;
	return UINT8_C(0xE7);
}

int main(void) {
	float position[2] = {123.5f, -456.75f};
	nox_object_t object = {.marker = UINTPTR_MAX};
	audio_event_zone_fn const zone = sub_501C00;

	assert(zone(position, &object) == UINT8_C(0xE7));
	assert(observed_position == position);
	assert(observed_position[0] == 123.5f);
	assert(observed_position[1] == -456.75f);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(zone(NULL, NULL) == UINT8_C(0xE7));
	assert(observed_position == NULL);
	assert(observed_object == NULL);
	return 0;
}
