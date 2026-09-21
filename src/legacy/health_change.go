package legacy

/*
#include <stddef.h>
#include <stdint.h>
#include <string.h>

#include "GAME2_3.h"
#include "common/alloc/classes/alloc_class.h"

extern void* nox_alloc_healthChange_1301772;
extern void* dword_5d4594_1301776;
extern void* dword_5d4594_1301780;

static void* nox_test_health_change_saved_allocator;
static void* nox_test_health_change_saved_head;
static void* nox_test_health_change_saved_font;

static int nox_test_health_change_begin(void) {
	nox_test_health_change_saved_allocator = nox_alloc_healthChange_1301772;
	nox_test_health_change_saved_head = dword_5d4594_1301776;
	nox_test_health_change_saved_font = dword_5d4594_1301780;
	nox_alloc_healthChange_1301772 = nox_new_alloc_class(
		"HealthChangeTest", (int)sizeof(nox_health_change), 8);
	dword_5d4594_1301776 = NULL;
	dword_5d4594_1301780 = NULL;
	return nox_alloc_healthChange_1301772 != NULL;
}

static void nox_test_health_change_end(void) {
	if (nox_alloc_healthChange_1301772 != NULL) {
		nox_free_alloc_class((nox_alloc_class*)nox_alloc_healthChange_1301772);
	}
	nox_alloc_healthChange_1301772 = nox_test_health_change_saved_allocator;
	dword_5d4594_1301776 = nox_test_health_change_saved_head;
	dword_5d4594_1301780 = nox_test_health_change_saved_font;
}

typedef struct nox_test_health_change_snapshot {
	uintptr_t address;
	uint32_t drawable_id;
	int16_t delta;
	uint32_t frame;
	uintptr_t next;
	uintptr_t prev;
} nox_test_health_change_snapshot;

static nox_test_health_change_snapshot nox_test_health_change_snapshot_at(uintptr_t address) {
	nox_test_health_change_snapshot out = {0};
	nox_health_change* change = (nox_health_change*)address;
	if (change == NULL) {
		return out;
	}
	out.address = (uintptr_t)change;
	out.drawable_id = change->drawable_id;
	out.delta = change->delta;
	out.frame = change->frame;
	out.next = (uintptr_t)change->next;
	out.prev = (uintptr_t)change->prev;
	return out;
}

static nox_test_health_change_snapshot nox_test_health_change_head(void) {
	return nox_test_health_change_snapshot_at((uintptr_t)dword_5d4594_1301776);
}

static int nox_test_health_change_head_empty(void) {
	return dword_5d4594_1301776 == NULL;
}

static uintptr_t nox_test_health_change_pointer_round_trip(int slot, uintptr_t value) {
	switch (slot) {
	case 0: {
		void* old = dword_5d4594_1301776;
		dword_5d4594_1301776 = (void*)value;
		uintptr_t result = (uintptr_t)dword_5d4594_1301776;
		dword_5d4594_1301776 = old;
		return result;
	}
	case 1: {
		void* old = dword_5d4594_1301780;
		dword_5d4594_1301780 = (void*)value;
		uintptr_t result = (uintptr_t)dword_5d4594_1301780;
		dword_5d4594_1301780 = old;
		return result;
	}
	case 2:
	case 3: {
		nox_health_change change = {0};
		if (slot == 2) {
			change.next = (nox_health_change*)value;
			return (uintptr_t)change.next;
		}
		change.prev = (nox_health_change*)value;
		return (uintptr_t)change.prev;
	}
	default:
		return 0;
	}
}

static void nox_test_health_change_draw(
	uint32_t drawable_id,
	intptr_t viewport_x,
	intptr_t viewport_y,
	intptr_t viewport_field_4,
	intptr_t viewport_field_5,
	intptr_t drawable_x,
	intptr_t drawable_y,
	uint16_t drawable_z,
	float drawable_field_25
) {
	nox_draw_viewport_t viewport = {0};
	nox_drawable drawable;
	memset(&drawable, 0, sizeof(drawable));
	viewport.x1 = viewport_x;
	viewport.y1 = viewport_y;
	viewport.field_4 = viewport_field_4;
	viewport.field_5 = viewport_field_5;
	drawable.pos.x = drawable_x;
	drawable.pos.y = drawable_y;
	drawable.z = drawable_z;
	drawable.field_25 = drawable_field_25;
	drawable.field_32 = drawable_id;
	sub_49A6A0(&viewport, &drawable);
}

static size_t nox_test_health_change_size(void) {
	return sizeof(nox_health_change);
}

static size_t nox_test_health_change_next_offset(void) {
	return offsetof(nox_health_change, next);
}

static size_t nox_test_health_change_prev_offset(void) {
	return offsetof(nox_health_change, prev);
}
*/
import "C"

type healthChangeSnapshot struct {
	address    uintptr
	drawableID uint32
	delta      int16
	frame      uint32
	next       uintptr
	prev       uintptr
}

func healthChangeTestBegin() bool {
	return C.nox_test_health_change_begin() != 0
}

func healthChangeTestEnd() {
	C.nox_test_health_change_end()
}

func healthChangeHeadSnapshot() healthChangeSnapshot {
	v := C.nox_test_health_change_head()
	return healthChangeSnapshotFromC(v)
}

func healthChangeSnapshotAt(address uintptr) healthChangeSnapshot {
	v := C.nox_test_health_change_snapshot_at(C.uintptr_t(address))
	return healthChangeSnapshotFromC(v)
}

func healthChangeSnapshotFromC(v C.nox_test_health_change_snapshot) healthChangeSnapshot {
	return healthChangeSnapshot{
		address:    uintptr(v.address),
		drawableID: uint32(v.drawable_id),
		delta:      int16(v.delta),
		frame:      uint32(v.frame),
		next:       uintptr(v.next),
		prev:       uintptr(v.prev),
	}
}

func healthChangeHeadEmpty() bool {
	return C.nox_test_health_change_head_empty() != 0
}

func healthChangePointerRoundTrip(slot int, value uintptr) uintptr {
	return uintptr(C.nox_test_health_change_pointer_round_trip(C.int(slot), C.uintptr_t(value)))
}

func healthChangeDraw(drawableID uint32, viewportX, viewportY, viewportField4, viewportField5, drawableX, drawableY int, drawableZ uint16, drawableField25 float32) {
	C.nox_test_health_change_draw(
		C.uint32_t(drawableID),
		C.intptr_t(viewportX),
		C.intptr_t(viewportY),
		C.intptr_t(viewportField4),
		C.intptr_t(viewportField5),
		C.intptr_t(drawableX),
		C.intptr_t(drawableY),
		C.uint16_t(drawableZ),
		C.float(drawableField25),
	)
}

func healthChangeNativeLayout() (size, nextOffset, prevOffset uintptr) {
	return uintptr(C.nox_test_health_change_size()),
		uintptr(C.nox_test_health_change_next_offset()),
		uintptr(C.nox_test_health_change_prev_offset())
}
