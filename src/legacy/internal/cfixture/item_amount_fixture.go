package cfixture

/*
#include <stddef.h>
#include "GAME3_1.h"

enum {
	item_amount_attrs_size = sizeof(nox_modifier_attrs_t),
	item_amount_attrs_field = offsetof(nox_modifier_attrs_t, field_16),
	item_amount_drawable_modifiers = offsetof(nox_drawable, item_modifiers),
	item_amount_drawable_field_low = offsetof(nox_drawable, item_field_112_0),
	item_amount_drawable_field_high = offsetof(nox_drawable, item_field_112_2),
	item_amount_callback_size = sizeof(nox_gui_item_amount_callback_t),
};

static int item_amount_attrs_contract(uint8_t* modifiers, uint32_t field_16) {
	nox_modifier_attrs_t attrs = {0};
	nox_drawable item = {0};
	for (int i = 0; i < 4; i++) {
		attrs.modifiers[i] = &modifiers[i];
	}
	attrs.field_16 = field_16;
	nox_gui_item_amount_copy_attrs(&item, &attrs);
	for (int i = 0; i < 4; i++) {
		if (item.item_modifiers[i] != attrs.modifiers[i]) {
			return 0;
		}
	}
	uint32_t copied = (uint16_t)item.item_field_112_0 | ((uint32_t)(uint16_t)item.item_field_112_2 << 16);
	return copied == field_16;
}

static struct {
	int2 position;
	uint32_t item_id, thing_type, amount, extra;
	int callback;
} item_amount_observed;

static void item_amount_accept(int2* pos, uint32_t id, uint32_t thing, uint32_t amount, uint32_t extra) {
	item_amount_observed.position = *pos;
	item_amount_observed.item_id = id;
	item_amount_observed.thing_type = thing;
	item_amount_observed.amount = amount;
	item_amount_observed.extra = extra;
	item_amount_observed.callback = 1;
}

static void item_amount_cancel(int2* pos, uint32_t id, uint32_t thing, uint32_t amount, uint32_t extra) {
	item_amount_accept(pos, id, thing, amount, extra);
	item_amount_observed.callback = 2;
}

static nox_gui_item_amount_callback_t item_amount_test_callback(int cancel) {
	return cancel ? item_amount_cancel : item_amount_accept;
}

static int item_amount_callback_contract(nox_gui_item_amount_callback_t callback, int expected) {
	uint32_t* thing = getMemU32Ptr(0x5D4594, 1319240);
	uint32_t* id = getMemU32Ptr(0x5D4594, 1319244);
	uint32_t* extra = getMemU32Ptr(0x5D4594, 1319252);
	uint32_t old_thing = *thing, old_id = *id, old_extra = *extra;
	*thing = UINT32_C(0x89ABCDEF);
	*id = UINT32_C(0x76543210);
	*extra = UINT32_C(0x13579BDF);
	item_amount_observed.callback = 0;
	int2 position = {.field_0 = -12345, .field_4 = 67890};
	nox_gui_item_amount_call_callback(callback, &position, UINT32_C(0x2468ACE0));

	int result = item_amount_observed.callback == expected &&
		item_amount_observed.position.field_0 == position.field_0 &&
		item_amount_observed.position.field_4 == position.field_4 &&
		item_amount_observed.item_id == *id && item_amount_observed.thing_type == *thing &&
		item_amount_observed.amount == UINT32_C(0x2468ACE0) && item_amount_observed.extra == *extra;
	*thing = old_thing;
	*id = old_id;
	*extra = old_extra;
	return result;
}
*/
import "C"

import "unsafe"

func ItemAmountNativeLayout() (attrsSize, attrsField, modifiers, fieldLow, fieldHigh, callbackSize uintptr) {
	return C.item_amount_attrs_size, C.item_amount_attrs_field, C.item_amount_drawable_modifiers,
		C.item_amount_drawable_field_low, C.item_amount_drawable_field_high, C.item_amount_callback_size
}

func ItemAmountAttrsContract(modifiers *[4]byte, field16 uint32) bool {
	return C.item_amount_attrs_contract((*C.uint8_t)(unsafe.Pointer(&modifiers[0])), C.uint32_t(field16)) != 0
}

func ItemAmountCallbacks() (accept, cancel unsafe.Pointer) {
	return unsafe.Pointer(C.item_amount_test_callback(0)), unsafe.Pointer(C.item_amount_test_callback(1))
}

func ItemAmountCallbackContract(callback unsafe.Pointer, expected int) bool {
	fn := C.nox_gui_item_amount_callback_t(callback)
	return C.item_amount_callback_contract(fn, C.int(expected)) != 0
}
