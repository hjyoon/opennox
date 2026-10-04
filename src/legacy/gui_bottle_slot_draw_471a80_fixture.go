//go:build !server

package legacy

/*
#include "GAME2_1.h"
#include "memmap.h"

extern nox_drawable nox_gui_bottle_drawables[3];
extern void* dword_5d4594_1096288;
extern uint32_t nox_color_white_2523948;

static int64_t nox_test_bottle_observed_471A80[30];
static int nox_test_bottle_slot_471A80;
static int nox_test_bottle_mutate_viewport_471A80;
static int nox_test_bottle_new_count_471A80;

static int nox_test_bottle_draw_471A80(uint32_t* raw, nox_drawable* dr) {
    nox_draw_viewport_t* vp = (nox_draw_viewport_t*)raw;
    int64_t* out = nox_test_bottle_observed_471A80;
    out[0] = vp->x1;
    out[1] = vp->y1;
    out[2] = vp->x2;
    out[3] = vp->y2;
    out[4] = vp->field_4;
    out[5] = vp->field_5;
    out[6] = vp->field_6;
    out[7] = vp->field_7;
    out[8] = vp->width;
    out[9] = vp->height;
    out[10] = vp->field_10;
    out[11] = vp->field_11;
    out[12] = vp->field_12;
    out[13] = (uintptr_t)dr;
    out[14] = (uintptr_t)vp;
    // Inspect the GUI sprite's coordinate operands without invoking signed
    // int overflow for the deliberately extreme field-conversion cases.
    out[17] = (int64_t)(int32_t)vp->x1 + dr->pos.x - (int32_t)vp->field_4;
    out[18] = (int64_t)(int32_t)vp->y1 + dr->pos.y - (int32_t)vp->field_5;
    out[19]++;
    if (nox_test_bottle_mutate_viewport_471A80) {
        vp->y1 = -123;
    }
    if (nox_test_bottle_new_count_471A80 >= 0) {
        *getMemU16Ptr(0x5D4594, 1090312 + 536 * nox_test_bottle_slot_471A80) =
            nox_test_bottle_new_count_471A80;
    }
    return -77; // The stock slot procedure still returns 1, not this value.
}

static int nox_test_bottle_calls_471A80(void) {
    return nox_test_bottle_observed_471A80[19];
}

static void nox_test_bottle_slot_draw_471A80(nox_window* win, const uint32_t* words,
        int slot, uint16_t count, int callback, int mutate_viewport, int new_count,
        const uint16_t* binding, int64_t* out) {
    uint32_t* record = getMemU32Ptr(0x5D4594, 1091904);
    uint32_t saved_record[15];
    memcpy(saved_record, record, sizeof(saved_record));
    record[0] = UINT32_C(0x11223344);
    memcpy(record + 1, words, 13 * sizeof(uint32_t));
    record[14] = UINT32_C(0x55667788);

    uint8_t saved_slots[3][20];
    nox_drawable saved_drawables[3];
    memcpy(saved_drawables, nox_gui_bottle_drawables, sizeof(saved_drawables));
    for (int i = 0; i < 3; i++) {
        memcpy(saved_slots[i], getMemAt(0x5D4594, 1090300 + 536 * i), 20);
    }
    uint8_t* slot_data = getMemAt(0x5D4594, 1090300 + 536 * slot);
    memset(slot_data, 0, 20);
    memcpy(slot_data, binding, 4 * sizeof(uint16_t));
    *getMemU32Ptr(0x5D4594, 1090308 + 536 * slot) = UINT32_C(0x12345678);
    *getMemU16Ptr(0x5D4594, 1090312 + 536 * slot) = count;

    nox_drawable* dr = &nox_gui_bottle_drawables[slot];
    memset(dr, 0, sizeof(*dr));
    dr->field_27 = UINT32_C(0x23456789);
    dr->flags30 = UINT32_C(0x40000000);
    dr->pos.x = -111;
    dr->pos.y = -222;
    dr->draw_func = callback ? nox_test_bottle_draw_471A80 : NULL;
    nox_drawable expected_drawable;
    memcpy(&expected_drawable, dr, sizeof(*dr));
    void* saved_widget = win->widget_data;
    void* saved_font = dword_5d4594_1096288;
    uint32_t saved_white = nox_color_white_2523948;
    win->widget_data = (void*)(intptr_t)slot;
    dword_5d4594_1096288 = NULL;
    nox_color_white_2523948 = UINT32_C(0xffff);
    memset(nox_test_bottle_observed_471A80, 0, sizeof(nox_test_bottle_observed_471A80));
    nox_test_bottle_slot_471A80 = slot;
    nox_test_bottle_mutate_viewport_471A80 = mutate_viewport;
    nox_test_bottle_new_count_471A80 = new_count;

    int result = nox_xxx_guiBottleSlotDrawFn_471A80(win, NULL);
    memcpy(out, nox_test_bottle_observed_471A80, sizeof(nox_test_bottle_observed_471A80));
    out[13] = (uintptr_t)dr;
    out[15] = dr->pos.x;
    out[16] = dr->pos.y;
    out[20] = result;
    out[21] = record[0] == UINT32_C(0x11223344) && record[14] == UINT32_C(0x55667788) &&
        memcmp(record + 1, words, 13 * sizeof(uint32_t)) == 0;
    out[22] = 1;
    out[28] = 1;
    for (int i = 0; i < 3; i++) {
        if (i != slot) {
            out[22] &= memcmp(&nox_gui_bottle_drawables[i], &saved_drawables[i], sizeof(*dr)) == 0;
            out[28] &= memcmp(getMemAt(0x5D4594, 1090300 + 536 * i), saved_slots[i], 20) == 0;
        }
    }
    expected_drawable.pos.x = dr->pos.x;
    expected_drawable.pos.y = dr->pos.y;
    out[23] = memcmp(dr, &expected_drawable, sizeof(*dr)) == 0;
    out[24] = *getMemU32Ptr(0x5D4594, 1090308 + 536 * slot) == UINT32_C(0x12345678);
    out[25] = memcmp(slot_data, binding, 4 * sizeof(uint16_t)) == 0;
    out[26] = dr->flags30;
    out[27] = *getMemU16Ptr(0x5D4594, 1090312 + 536 * slot);

    win->widget_data = saved_widget;
    out[29] = win->widget_data == saved_widget;
    dword_5d4594_1096288 = saved_font;
    nox_color_white_2523948 = saved_white;
    memcpy(record, saved_record, sizeof(saved_record));
    memcpy(nox_gui_bottle_drawables, saved_drawables, sizeof(saved_drawables));
    for (int i = 0; i < 3; i++) {
        memcpy(getMemAt(0x5D4594, 1090300 + 536 * i), saved_slots[i], 20);
    }
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

func bottleSlotDrawC471A80(win *gui.Window, words [13]uint32, slot int, count uint16,
	callback, mutateViewport bool, newCount int, binding [4]uint16) [30]int64 {
	var out [30]int64
	C.nox_test_bottle_slot_draw_471A80((*C.nox_window)(win.C()),
		(*C.uint32_t)(unsafe.Pointer(&words[0])), C.int(slot), C.uint16_t(count),
		C.int(bool2int(callback)), C.int(bool2int(mutateViewport)), C.int(newCount),
		(*C.uint16_t)(unsafe.Pointer(&binding[0])), (*C.int64_t)(unsafe.Pointer(&out[0])))
	return out
}

func bottleSlotDrawCallCount471A80() int {
	return int(C.nox_test_bottle_calls_471A80())
}
