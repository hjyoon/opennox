#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../xfer_trigger_4f4e50.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef struct nox_script_callback_t {
	uint32_t flags;
	int32_t function;
} nox_script_callback_t;

typedef struct nox_trigger_update_data_t {
	uint32_t flags;
	uint32_t field_4;
	uint8_t state;
	uint8_t field_9;
	uint8_t reserved_10[2];
	nox_script_callback_t script_collide;
	nox_script_callback_t script_activate;
	nox_script_callback_t script_deactivate;
	uint32_t sound_activate;
	uint32_t sound_deactivate;
	uint32_t class_include;
	uint32_t class_exclude;
	uint8_t team_include;
	uint8_t team_exclude;
	uint8_t colors[6];
} nox_trigger_update_data_t;

typedef int32_t (*trigger_xfer_fn_4F4E50)(nox_object_t*, void*);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8,
	"unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4,
	"TriggerXfer result must remain int32");
_Static_assert(sizeof(nox_script_callback_t) == 8,
	"script callback must remain eight bytes");
_Static_assert(offsetof(nox_script_callback_t, function) == 4,
	"script function must remain at byte 4");
_Static_assert(sizeof(nox_trigger_update_data_t) == 60,
	"Trigger update data must remain 60 bytes");
_Static_assert(offsetof(nox_trigger_update_data_t, flags) == 0,
	"Trigger flags must remain at byte 0");
_Static_assert(offsetof(nox_trigger_update_data_t, field_4) == 4,
	"Trigger PE32 transient slot must remain at byte 4");
_Static_assert(offsetof(nox_trigger_update_data_t, state) == 8,
	"Trigger state must remain at byte 8");
_Static_assert(offsetof(nox_trigger_update_data_t, field_9) == 9,
	"Trigger field 9 must remain at byte 9");
_Static_assert(offsetof(nox_trigger_update_data_t, script_collide) == 12,
	"Trigger collide callback must remain at byte 12");
_Static_assert(offsetof(nox_trigger_update_data_t, script_activate) == 20,
	"Trigger activate callback must remain at byte 20");
_Static_assert(offsetof(nox_trigger_update_data_t, script_deactivate) == 28,
	"Trigger deactivate callback must remain at byte 28");
_Static_assert(offsetof(nox_trigger_update_data_t, sound_activate) == 36,
	"Trigger activate sound must remain at byte 36");
_Static_assert(offsetof(nox_trigger_update_data_t, sound_deactivate) == 40,
	"Trigger deactivate sound must remain at byte 40");
_Static_assert(offsetof(nox_trigger_update_data_t, class_include) == 44,
	"Trigger included class must remain at byte 44");
_Static_assert(offsetof(nox_trigger_update_data_t, class_exclude) == 48,
	"Trigger excluded class must remain at byte 48");
_Static_assert(offsetof(nox_trigger_update_data_t, team_include) == 52,
	"Trigger included team must remain at byte 52");
_Static_assert(offsetof(nox_trigger_update_data_t, team_exclude) == 53,
	"Trigger excluded team must remain at byte 53");
_Static_assert(offsetof(nox_trigger_update_data_t, colors) == 54,
	"Trigger colors must remain at byte 54");
_Static_assert(
	_Generic(&nox_xxx_unitTriggerXfer_4F4E50,
		trigger_xfer_fn_4F4E50: 1, default: 0),
	"TriggerXfer must preserve two native pointers and its exact int32 result");

static nox_object_t* observed_object;
static void* observed_context;

int32_t nox_xxx_unitTriggerXfer_4F4E50(
	nox_object_t* object,
	void* context) {
	observed_object = object;
	observed_context = context;
	return object != NULL;
}

int main(void) {
	nox_object_t object = {.marker = UINTPTR_MAX};
	void* const context = (void*)(uintptr_t)(UINTPTR_MAX - 15u);
	trigger_xfer_fn_4F4E50 const transfer = nox_xxx_unitTriggerXfer_4F4E50;

	assert(transfer(&object, context) == 1);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(observed_context == context);

	assert(transfer(NULL, NULL) == 0);
	assert(observed_object == NULL);
	assert(observed_context == NULL);
	return 0;
}
