#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../xfer_exit_4f4b90.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef struct nox_exit_collide_data_t {
	uint8_t map_name[80];
	float destination_x;
	float destination_y;
} nox_exit_collide_data_t;

typedef int32_t (*exit_xfer_fn_4F4B90)(nox_object_t*, void*);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4, "ExitXfer result must remain int32");
_Static_assert(sizeof(nox_exit_collide_data_t) == 88,
	"Exit collide data must remain 88 bytes");
_Static_assert(offsetof(nox_exit_collide_data_t, destination_x) == 80,
	"Exit destination X must remain at byte 80");
_Static_assert(offsetof(nox_exit_collide_data_t, destination_y) == 84,
	"Exit destination Y must remain at byte 84");
_Static_assert(
	_Generic(&nox_xxx_XFerExit_4F4B90,
		exit_xfer_fn_4F4B90: 1, default: 0),
	"ExitXfer must preserve two native pointers and its exact int32 result");

static nox_object_t* observed_object;
static void* observed_context;

int32_t nox_xxx_XFerExit_4F4B90(nox_object_t* object, void* context) {
	observed_object = object;
	observed_context = context;
	return object != NULL;
}

int main(void) {
	nox_object_t object = {.marker = UINTPTR_MAX};
	void* const context = (void*)(uintptr_t)(UINTPTR_MAX - 15u);
	exit_xfer_fn_4F4B90 const transfer = nox_xxx_XFerExit_4F4B90;

	assert(transfer(&object, context) == 1);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(observed_context == context);

	assert(transfer(NULL, NULL) == 0);
	assert(observed_object == NULL);
	assert(observed_context == NULL);
	return 0;
}
