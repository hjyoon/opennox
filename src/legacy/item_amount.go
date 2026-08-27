package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>

#include "GAME3_1.h"

_Static_assert(sizeof(nox_gui_item_amount_callback_t) == sizeof(uintptr_t),
	"item-amount callback must occupy one native pointer slot");

enum {
	NOX_TEST_ITEM_AMOUNT_DIALOG,
	NOX_TEST_ITEM_AMOUNT_ITEM,
	NOX_TEST_ITEM_AMOUNT_IMAGE,
	NOX_TEST_ITEM_AMOUNT_ACCEPT,
	NOX_TEST_ITEM_AMOUNT_CANCEL,
};

static uintptr_t nox_test_item_amount_pointer_round_trip(int slot, uintptr_t value) {
	switch (slot) {
	case NOX_TEST_ITEM_AMOUNT_DIALOG: {
		nox_window* old = nox_gui_itemAmount_dialog_1319228;
		nox_gui_itemAmount_dialog_1319228 = (nox_window*)value;
		uintptr_t result = (uintptr_t)nox_gui_itemAmount_dialog_1319228;
		nox_gui_itemAmount_dialog_1319228 = old;
		return result;
	}
	case NOX_TEST_ITEM_AMOUNT_ITEM: {
		nox_drawable* old = nox_gui_itemAmount_item_1319256;
		nox_gui_itemAmount_item_1319256 = (nox_drawable*)value;
		uintptr_t result = (uintptr_t)nox_gui_itemAmount_item_1319256;
		nox_gui_itemAmount_item_1319256 = old;
		return result;
	}
	case NOX_TEST_ITEM_AMOUNT_IMAGE: {
		nox_video_bag_image_t* old = nox_gui_itemAmount_images_1319196[7];
		nox_gui_itemAmount_images_1319196[7] = (nox_video_bag_image_t*)value;
		uintptr_t result = (uintptr_t)nox_gui_itemAmount_images_1319196[7];
		nox_gui_itemAmount_images_1319196[7] = old;
		return result;
	}
	case NOX_TEST_ITEM_AMOUNT_ACCEPT:
	case NOX_TEST_ITEM_AMOUNT_CANCEL: {
		nox_gui_item_amount_callback_t callback;
		memcpy(&callback, &value, sizeof(callback));
		nox_gui_item_amount_callback_t* target = slot == NOX_TEST_ITEM_AMOUNT_ACCEPT
			? &nox_gui_itemAmount_accept_1319160
			: &nox_gui_itemAmount_cancel_1319100;
		nox_gui_item_amount_callback_t old = *target;
		*target = callback;
		uintptr_t result = 0;
		memcpy(&result, target, sizeof(result));
		*target = old;
		return result;
	}
	default:
		return 0;
	}
}

static int nox_test_item_amount_attrs_contract(uintptr_t base, uint32_t field_16) {
	nox_modifier_attrs_t attrs = {0};
	nox_drawable item = {0};
	for (int i = 0; i < 4; i++) {
		attrs.modifiers[i] = (void*)(base + (uintptr_t)(i * 0x101));
	}
	attrs.field_16 = field_16;
	nox_gui_item_amount_copy_attrs(&item, &attrs);
	for (int i = 0; i < 4; i++) {
		if (item.item_modifiers[i] != attrs.modifiers[i]) {
			return 0;
		}
	}
	uint32_t copied_field_16 = (uint16_t)item.item_field_112_0 |
		((uint32_t)(uint16_t)item.item_field_112_2 << 16);
	return copied_field_16 == field_16;
}

typedef struct nox_test_item_amount_callback_observation {
	int2 position;
	uint32_t item_id;
	uint32_t thing_type;
	uint32_t amount;
	uint32_t extra;
} nox_test_item_amount_callback_observation;

static nox_test_item_amount_callback_observation nox_test_item_amount_observed;

static void nox_test_item_amount_callback(int2* position, uint32_t item_id, uint32_t thing_type,
	uint32_t amount, uint32_t extra) {
	nox_test_item_amount_observed.position = *position;
	nox_test_item_amount_observed.item_id = item_id;
	nox_test_item_amount_observed.thing_type = thing_type;
	nox_test_item_amount_observed.amount = amount;
	nox_test_item_amount_observed.extra = extra;
}

static int nox_test_item_amount_callback_contract(void) {
	uint32_t* thing_type = getMemU32Ptr(0x5D4594, 1319240);
	uint32_t* item_id = getMemU32Ptr(0x5D4594, 1319244);
	uint32_t* extra = getMemU32Ptr(0x5D4594, 1319252);
	uint32_t old_thing_type = *thing_type;
	uint32_t old_item_id = *item_id;
	uint32_t old_extra = *extra;
	nox_gui_item_amount_callback_t old_accept = nox_gui_itemAmount_accept_1319160;

	*thing_type = UINT32_C(0x89ABCDEF);
	*item_id = UINT32_C(0x76543210);
	*extra = UINT32_C(0x13579BDF);
	nox_gui_itemAmount_accept_1319160 = nox_test_item_amount_callback;
	memset(&nox_test_item_amount_observed, 0, sizeof(nox_test_item_amount_observed));
	int2 position = {.field_0 = -12345, .field_4 = 67890};
	nox_gui_item_amount_call_callback(nox_gui_itemAmount_accept_1319160, &position, UINT32_C(0x2468ACE0));

	int result = nox_test_item_amount_observed.position.field_0 == position.field_0 &&
		nox_test_item_amount_observed.position.field_4 == position.field_4 &&
		nox_test_item_amount_observed.item_id == *item_id &&
		nox_test_item_amount_observed.thing_type == *thing_type &&
		nox_test_item_amount_observed.amount == UINT32_C(0x2468ACE0) &&
		nox_test_item_amount_observed.extra == *extra;

	*thing_type = old_thing_type;
	*item_id = old_item_id;
	*extra = old_extra;
	nox_gui_itemAmount_accept_1319160 = old_accept;
	return result;
}

static size_t nox_test_item_amount_attrs_size(void) { return sizeof(nox_modifier_attrs_t); }
static size_t nox_test_item_amount_attrs_field_offset(void) { return offsetof(nox_modifier_attrs_t, field_16); }
static size_t nox_test_item_amount_drawable_modifiers_offset(void) { return offsetof(nox_drawable, item_modifiers); }
static size_t nox_test_item_amount_drawable_field_low_offset(void) { return offsetof(nox_drawable, item_field_112_0); }
static size_t nox_test_item_amount_drawable_field_high_offset(void) { return offsetof(nox_drawable, item_field_112_2); }
static size_t nox_test_item_amount_callback_size(void) { return sizeof(nox_gui_item_amount_callback_t); }
*/
import "C"

// ItemAmountCancel4BFE40 closes the quantity dialog if it is active. The
// restored C implementation stores every callback and window at native
// pointer width, so this boundary is safe on 64-bit clients.
func ItemAmountCancel4BFE40() int {
	return int(C.sub_4BFE40())
}

func itemAmountPointerRoundTrip(slot int, value uintptr) uintptr {
	return uintptr(C.nox_test_item_amount_pointer_round_trip(C.int(slot), C.uintptr_t(value)))
}

func itemAmountAttrsContract(base uintptr, field16 uint32) bool {
	return C.nox_test_item_amount_attrs_contract(C.uintptr_t(base), C.uint32_t(field16)) != 0
}

func itemAmountCallbackContract() bool {
	return C.nox_test_item_amount_callback_contract() != 0
}

func itemAmountNativeLayout() (attrsSize, attrsField, drawableModifiers, drawableFieldLow, drawableFieldHigh, callbackSize uintptr) {
	return uintptr(C.nox_test_item_amount_attrs_size()),
		uintptr(C.nox_test_item_amount_attrs_field_offset()),
		uintptr(C.nox_test_item_amount_drawable_modifiers_offset()),
		uintptr(C.nox_test_item_amount_drawable_field_low_offset()),
		uintptr(C.nox_test_item_amount_drawable_field_high_offset()),
		uintptr(C.nox_test_item_amount_callback_size())
}
