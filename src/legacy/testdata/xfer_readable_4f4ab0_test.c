#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../xfer_readable_4f4ab0.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef struct nox_readable_use_data_t {
	uint8_t text[256];
	uint32_t transient_read_state;
} nox_readable_use_data_t;

typedef int32_t (*readable_xfer_fn_4F4AB0)(nox_object_t*, void*);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4, "ReadableXfer result must remain int32");
_Static_assert(sizeof(nox_readable_use_data_t) == 260,
	"Readable use data must remain 260 bytes");
_Static_assert(offsetof(nox_readable_use_data_t, transient_read_state) == 256,
	"Readable transient state must remain at byte 256");
_Static_assert(
	_Generic(&nox_xxx_XFerReadable_4F4AB0,
		readable_xfer_fn_4F4AB0: 1, default: 0),
	"ReadableXfer must preserve two native pointers and its exact int32 result");

static nox_object_t* observed_object;
static void* observed_context;

int32_t nox_xxx_XFerReadable_4F4AB0(nox_object_t* object, void* context) {
	observed_object = object;
	observed_context = context;
	return object != NULL;
}

int main(void) {
	nox_object_t object = {.marker = UINTPTR_MAX};
	void* const context = (void*)(uintptr_t)(UINTPTR_MAX - 15u);
	readable_xfer_fn_4F4AB0 const transfer = nox_xxx_XFerReadable_4F4AB0;

	assert(transfer(&object, context) == 1);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(observed_context == context);

	assert(transfer(NULL, NULL) == 0);
	assert(observed_object == NULL);
	assert(observed_context == NULL);
	return 0;
}
