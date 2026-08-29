#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../xfer_transporter_4f5300.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef int32_t (*transporter_xfer_fn_4F5300)(nox_object_t*, void*);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"TransporterXfer result must remain int32");
_Static_assert(sizeof(nox_transporter_update_data_t) == 20,
	"Transporter update data must remain twenty bytes");
_Static_assert(offsetof(nox_transporter_update_data_t, target_pe32) == 12,
	"Transporter PE32 target slot must remain at byte 12");
_Static_assert(offsetof(nox_transporter_update_data_t, target_extent) == 16,
	"Transporter target extent must remain at byte 16");
_Static_assert(
	_Generic(&nox_xxx_XFerTransporter_4F5300,
		transporter_xfer_fn_4F5300: 1, default: 0),
	"TransporterXfer must preserve two native pointers and its exact int32 result");

static nox_object_t* observed_object;
static void* observed_context;

int32_t nox_xxx_XFerTransporter_4F5300(
	nox_object_t* object,
	void* context) {
	observed_object = object;
	observed_context = context;
	return object != NULL;
}

int main(void) {
	nox_object_t object = {.marker = UINTPTR_MAX};
	void* const context = (void*)(uintptr_t)(UINTPTR_MAX - 15u);
	transporter_xfer_fn_4F5300 const transfer =
		nox_xxx_XFerTransporter_4F5300;

	assert(transfer(&object, context) == 1);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(observed_context == context);

	assert(transfer(NULL, NULL) == 0);
	assert(observed_object == NULL);
	assert(observed_context == NULL);
	return 0;
}
