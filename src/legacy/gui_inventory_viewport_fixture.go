package legacy

/*
#include "GAME2_1.h"
#include "memmap.h"

static int nox_test_inventory_viewport_native(const uint32_t* words, int64_t* fields) {
    uint32_t* record = getMemU32Ptr(0x5D4594, 1049728);
    uint32_t saved[15];
    memcpy(saved, record, sizeof(saved));
    record[0] = UINT32_C(0x11223344);
    memcpy(record + 1, words, 13 * sizeof(uint32_t));
    record[14] = UINT32_C(0x55667788);
    const nox_draw_viewport_t* viewport = nox_client_inventory_viewport_native();
    fields[0] = viewport->x1;
    fields[1] = viewport->y1;
    fields[2] = viewport->x2;
    fields[3] = viewport->y2;
    fields[4] = viewport->field_4;
    fields[5] = viewport->field_5;
    fields[6] = viewport->field_6;
    fields[7] = viewport->field_7;
    fields[8] = viewport->width;
    fields[9] = viewport->height;
    fields[10] = viewport->field_10;
    fields[11] = viewport->field_11;
    fields[12] = viewport->field_12;
    int unchanged = record[0] == UINT32_C(0x11223344) && record[14] == UINT32_C(0x55667788) &&
        memcmp(record + 1, words, 13 * sizeof(uint32_t)) == 0;
    memcpy(record, saved, sizeof(saved));
    return unchanged;
}
*/
import "C"

import "unsafe"

func inventoryViewportNativeFields(words [13]uint32) ([13]int64, bool) {
	var fields [13]int64
	unchanged := C.nox_test_inventory_viewport_native(
		(*C.uint32_t)(unsafe.Pointer(&words[0])),
		(*C.int64_t)(unsafe.Pointer(&fields[0])),
	) != 0
	return fields, unchanged
}
