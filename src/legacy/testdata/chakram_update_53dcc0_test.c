// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert only ChakramInMotionUpdate's native object, data and callback ABI.
#define _Static_assert(...)
#include "../GAME4_3.h"
#undef _Static_assert

#include <math.h>
#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_chakram_update_data_t, field_0) == 0,
	"Chakram field-0 offset");
_Static_assert(offsetof(nox_chakram_update_data_t, reflections) == 4,
	"Chakram reflections offset");
_Static_assert(offsetof(nox_chakram_update_data_t, return_target) == 8,
	"Chakram return-target offset");
_Static_assert(offsetof(nox_chakram_update_data_t, last_hit) == 8 + sizeof(void*),
	"Chakram last-hit offset");
_Static_assert(offsetof(nox_chakram_update_data_t, owner_x) == 8 + 2 * sizeof(void*),
	"Chakram owner-X offset");
_Static_assert(offsetof(nox_chakram_update_data_t, owner_y) == 12 + 2 * sizeof(void*),
	"Chakram owner-Y offset");
_Static_assert(offsetof(nox_chakram_update_data_t, return_state) == 16 + 2 * sizeof(void*),
	"Chakram return-state offset");
_Static_assert(sizeof(nox_chakram_update_data_t) == (sizeof(void*) == 4 ? 28 : 40),
	"Chakram update-data size");

_Static_assert(offsetof(nox_object_t, obj_flags) == (sizeof(void*) == 4 ? 16 : 20),
	"object flags offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60),
	"object position offset");
_Static_assert(offsetof(nox_object_t, vel_x) == (sizeof(void*) == 4 ? 80 : 84),
	"object velocity offset");
_Static_assert(offsetof(nox_object_t, field_32) == (sizeof(void*) == 4 ? 128 : 132),
	"object creation-frame offset");
_Static_assert(offsetof(nox_object_t, inv_first_item) == (sizeof(void*) == 4 ? 504 : 544),
	"object inventory-first offset");
_Static_assert(offsetof(nox_object_t, owner) == (sizeof(void*) == 4 ? 508 : 552),
	"object owner offset");
_Static_assert(offsetof(nox_object_t, speed_cur) == (sizeof(void*) == 4 ? 544 : 604),
	"object current-speed offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872),
	"object update-data offset");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_updateChakramInMotion_53DCC0),
	void (*)(nox_object_t*)),
	"Chakram update callback pointer width");

static nox_object_t* seen_source;

void nox_xxx_updateChakramInMotion_53DCC0(nox_object_t* source) {
	seen_source = source;
}

static uint32_t reference_frame;
static uint32_t reference_rate;
static int delete_calls;
static nox_object_t* map_seen_owner;
static nox_object_t* map_replacement;

static int reference_map_check(nox_object_t* source, nox_object_t* owner) {
	map_seen_owner = owner;
	if (map_replacement != NULL) {
		source->owner = map_replacement;
	}
	return 1;
}

static void reference_delayed_delete(nox_object_t* source) {
	(void)source;
	delete_calls++;
}

// Portable reference for the native fields and live/cached reads used by
// GAME.EXE 0053DCC0. The Go semantic suite covers every branch independently;
// this copy makes the C ABI fixture executable on native 32/64-bit layouts.
static void chakram_update_reference(nox_object_t* source) {
	nox_chakram_update_data_t* update = source->data_update;
	nox_object_t* item = source->inv_first_item;
	nox_object_t* owner;

	if (item == NULL || (item->obj_flags & UINT32_C(0x20)) != 0) {
		reference_delayed_delete(source);
		return;
	}
	if (update->last_hit != NULL &&
		(update->last_hit->obj_flags & UINT32_C(0x20)) != 0) {
		update->last_hit = NULL;
	}

	owner = source->owner;
	if (owner == NULL || (owner->obj_flags & UINT32_C(0x20)) != 0) {
		update->return_state = UINT8_C(1);
		update->return_target = NULL;
	} else {
		update->owner_x = owner->x;
		update->owner_y = owner->y;
		if (!reference_map_check(source, source->owner)) {
			update->return_target = NULL;
		} else if (update->return_state != 0) {
			goto lifetime;
		} else {
			update->return_target = source->owner;
		}
		if (update->return_state != 0) {
			goto lifetime;
		}
		if (update->return_target != NULL &&
			(update->return_target->obj_flags & UINT32_C(0x8020)) != 0) {
			update->return_target = NULL;
			update->return_state = UINT8_C(1);
		} else {
			double dx = (double)update->owner_x - (double)source->x;
			double dy_extended = (double)update->owner_y - (double)source->y;
			float dy = (float)dy_extended;
			float denominator = (float)(
				sqrt(dy_extended * (double)dy + dx * dx) + (double)(float)0.1f);
			source->vel_x = (float)(dx * (double)source->speed_cur / (double)denominator);
			source->vel_y = (float)((double)dy * (double)source->speed_cur /
				(double)denominator);
		}
	}

lifetime:
	if (reference_frame - source->field_32 > reference_rate * UINT32_C(5)) {
		update->return_state = UINT8_C(1);
		update->return_target = NULL;
	}
}

int main(void) {
	nox_chakram_update_data_t update = {0};
	nox_object_t item = {0};
	nox_object_t last = {.obj_flags = UINT32_C(0x20)};
	nox_object_t cached_owner = {.x = 13.0f, .y = 24.0f};
	nox_object_t live_owner = {.obj_flags = UINT32_C(0x8000)};
	nox_object_t source = {
		.x = 10.0f,
		.y = 20.0f,
		.field_32 = UINT32_MAX - UINT32_C(2),
		.inv_first_item = &item,
		.owner = &cached_owner,
		.speed_cur = 5.0f,
		.data_update = &update,
	};

	nox_xxx_updateChakramInMotion_53DCC0(&source);
	if (seen_source != &source) {
		return 1;
	}

	update.last_hit = &last;
	update.return_target = &cached_owner;
	map_replacement = &live_owner;
	reference_frame = UINT32_C(2);
	reference_rate = UINT32_C(1);
	chakram_update_reference(&source);
	if (map_seen_owner != &cached_owner || source.owner != &live_owner ||
		update.last_hit != NULL || update.owner_x != 13.0f || update.owner_y != 24.0f ||
		update.return_target != NULL || update.return_state != UINT8_C(1)) {
		return 2;
	}

	map_replacement = NULL;
	map_seen_owner = NULL;
	source.owner = &cached_owner;
	update.return_state = UINT8_C(2);
	update.return_target = &live_owner;
	reference_frame = UINT32_C(2);
	chakram_update_reference(&source);
	if (update.return_state != UINT8_C(2) || update.return_target != &live_owner) {
		return 3;
	}
	reference_frame = UINT32_C(3);
	chakram_update_reference(&source);
	if (update.return_state != UINT8_C(1) || update.return_target != NULL) {
		return 4;
	}

	update.return_state = 0;
	source.field_32 = UINT32_C(100);
	reference_frame = UINT32_C(101);
	reference_rate = UINT32_C(30);
	chakram_update_reference(&source);
	if (source.vel_x < 2.94f || source.vel_x > 2.95f ||
		source.vel_y < 3.92f || source.vel_y > 3.93f) {
		return 5;
	}

	source.inv_first_item = NULL;
	chakram_update_reference(&source);
	if (delete_calls != 1) {
		return 6;
	}
	return 0;
}
