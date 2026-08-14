// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert TeleportCollide's native object, data and callback ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>
#include <string.h>

_Static_assert(sizeof(nox_teleport_collide_data_t) == 8, "Teleport collide-data size");
_Static_assert(offsetof(nox_teleport_collide_data_t, destination_x) == 0, "Teleport X offset");
_Static_assert(offsetof(nox_teleport_collide_data_t, destination_y) == 4, "Teleport Y offset");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60), "object position offset");
_Static_assert(offsetof(nox_object_t, field_41) == (sizeof(void*) == 4 ? 164 : 168), "object X destination offset");
_Static_assert(offsetof(nox_object_t, field_42) == (sizeof(void*) == 4 ? 168 : 172), "object Y destination offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776), "object collide-data offset");
_Static_assert(__builtin_types_compatible_p(__typeof__(&sub_4EACA0),
											void (*)(nox_object_t*, nox_object_t*, float*)),
			   "TeleportCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void sub_4EACA0(nox_object_t* source, nox_object_t* target, float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

static void teleport_reference_destination(nox_object_t* source, nox_object_t* target) {
	nox_teleport_collide_data_t* data = source->collide_data;
	if (target == NULL || ((uint8_t)target->obj_class & UINT8_C(0x80)) != 0) {
		return;
	}
	float x = (float)data->destination_x;
	float y = (float)data->destination_y;
	memcpy(&target->field_41, &x, sizeof(x));
	memcpy(&target->field_42, &y, sizeof(y));
}

static float teleport_float_from_bits(uint32_t bits) {
	float value;
	memcpy(&value, &bits, sizeof(value));
	return value;
}

int main(void) {
	nox_teleport_collide_data_t data = {
		.destination_x = INT32_C(16777217),
		.destination_y = INT32_MIN,
	};
	nox_object_t source = {.collide_data = &data};
	nox_object_t target = {
		.obj_class = UINT32_C(0x80000004),
		.x = 10.5f,
		.y = -20.25f,
		.field_41 = UINT32_C(0x11223344),
		.field_42 = UINT32_C(0x55667788),
	};
	float collision[2] = {3.5f, -8.25f};

	sub_4EACA0(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}

	teleport_reference_destination(&source, &target);
	if (teleport_float_from_bits(target.field_41) != 16777216.0f ||
		teleport_float_from_bits(target.field_42) != (float)INT32_MIN) {
		return 2;
	}

	target.obj_class |= UINT32_C(0x80);
	target.field_41 = UINT32_C(0x11223344);
	target.field_42 = UINT32_C(0x55667788);
	teleport_reference_destination(&source, &target);
	if (target.field_41 != UINT32_C(0x11223344) || target.field_42 != UINT32_C(0x55667788)) {
		return 3;
	}
	if (collision[0] != 3.5f || collision[1] != -8.25f) {
		return 4;
	}
	return 0;
}
