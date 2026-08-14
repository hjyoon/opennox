// Suppress unrelated Win32-only assertions while parsing the shared headers,
// then assert TrapDoorCollide's native object, data and callback ABI.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <math.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>

_Static_assert(sizeof(nox_trap_door_script_callback_t) == 8, "TrapDoor script block size");
_Static_assert(sizeof(nox_trap_door_collide_data_t) == 28, "TrapDoor collide-data size");
_Static_assert(offsetof(nox_trap_door_collide_data_t, script) == 0, "TrapDoor script offset");
_Static_assert(offsetof(nox_trap_door_collide_data_t, fall_velocity_x) == 8, "TrapDoor X velocity offset");
_Static_assert(offsetof(nox_trap_door_collide_data_t, fall_velocity_y) == 12, "TrapDoor Y velocity offset");
_Static_assert(offsetof(nox_trap_door_collide_data_t, next_frame) == 16, "TrapDoor next-frame offset");
_Static_assert(offsetof(nox_trap_door_collide_data_t, delay) == 20, "TrapDoor delay offset");
_Static_assert(offsetof(nox_trap_door_collide_data_t, reserved_22) == 22, "TrapDoor reserved offset");
_Static_assert(offsetof(nox_trap_door_collide_data_t, activated) == 24, "TrapDoor activated offset");
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, obj_flags) == (sizeof(void*) == 4 ? 16 : 20), "object flags offset");
_Static_assert(offsetof(nox_object_t, x) == (sizeof(void*) == 4 ? 56 : 60), "object position offset");
_Static_assert(offsetof(nox_object_t, float_39) == (sizeof(void*) == 4 ? 156 : 160), "object fall-position offset");
_Static_assert(offsetof(nox_object_t, field_41) == (sizeof(void*) == 4 ? 164 : 168), "object X fall-velocity offset");
_Static_assert(offsetof(nox_object_t, field_42) == (sizeof(void*) == 4 ? 168 : 172), "object Y fall-velocity offset");
_Static_assert(offsetof(nox_object_t, shape) == (sizeof(void*) == 4 ? 172 : 176), "object shape offset");
_Static_assert(offsetof(nox_object_t, collide_data) == (sizeof(void*) == 4 ? 700 : 776), "object collide-data offset");
_Static_assert(__builtin_types_compatible_p(__typeof__(&nox_xxx_collideTrapDoor_4EAB60),
											void (*)(nox_object_t*, nox_object_t*, float*)),
			   "TrapDoorCollide callback pointer width");

static nox_object_t* seen_source;
static nox_object_t* seen_target;
static float* seen_collision;

void nox_xxx_collideTrapDoor_4EAB60(nox_object_t* source, nox_object_t* target, float* collision) {
	seen_source = source;
	seen_target = target;
	seen_collision = collision;
}

static void trap_door_reference_fall(nox_object_t* source, nox_object_t* target) {
	nox_trap_door_collide_data_t* data = source->collide_data;
	float velocity_x = (float)data->fall_velocity_x;
	float velocity_y = (float)data->fall_velocity_y;
	target->obj_flags |= UINT32_C(0x00060000);
	memcpy(&target->field_41, &velocity_x, sizeof(velocity_x));
	memcpy(&target->field_42, &velocity_y, sizeof(velocity_y));
	target->float_39 = source->x;
	target->float_40 = source->y;
}

static float trap_door_float_from_bits(uint32_t bits) {
	float value;
	memcpy(&value, &bits, sizeof(value));
	return value;
}

static void trap_door_reference_inactive(nox_object_t* source, nox_object_t* target, uint32_t frame) {
	nox_trap_door_collide_data_t* data = source->collide_data;
	if (target == NULL || (uint8_t)target->obj_class & UINT8_C(0x80) || data->activated != 0) {
		return;
	}
	if (data->delay != 0) {
		data->next_frame = frame + (uint32_t)data->delay;
	}
	data->activated = UINT32_C(1);
}

int main(void) {
	nox_trap_door_collide_data_t data = {
		.script = {.flags = UINT32_C(0xa55a5aa5), .func = -17},
		.fall_velocity_x = INT32_C(16777217),
		.fall_velocity_y = INT32_C(-31),
		.next_frame = UINT32_C(0x11223344),
		.delay = UINT16_C(7),
		.reserved_22 = UINT16_C(0x55aa),
	};
	nox_object_t source = {
		.obj_flags = UINT32_C(0x01000000),
		.x = 30.25f,
		.y = -10.5f,
		.collide_data = &data,
	};
	nox_object_t target = {
		.obj_flags = UINT32_C(0x20),
		.float_39 = 101.0f,
		.float_40 = 102.0f,
		.field_41 = UINT32_C(0x11223344),
		.field_42 = UINT32_C(0x55667788),
	};
	float collision[2] = {3.5f, -8.25f};

	nox_xxx_collideTrapDoor_4EAB60(&source, &target, collision);
	if (seen_source != &source || seen_target != &target || seen_collision != collision) {
		return 1;
	}

	trap_door_reference_fall(&source, &target);
	if (target.obj_flags != UINT32_C(0x60020) || trap_door_float_from_bits(target.field_41) != 16777216.0f ||
		trap_door_float_from_bits(target.field_42) != -31.0f || target.float_39 != source.x ||
		target.float_40 != source.y) {
		return 2;
	}
	if (data.script.flags != UINT32_C(0xa55a5aa5) || data.script.func != -17 || data.reserved_22 != UINT16_C(0x55aa)) {
		return 3;
	}

	source.obj_flags = 0;
	data.activated = 0;
	trap_door_reference_inactive(&source, &target, UINT32_C(0xfffffffc));
	if (data.next_frame != UINT32_C(3) || data.activated != UINT32_C(1)) {
		return 4;
	}
	trap_door_reference_inactive(&source, NULL, 10);
	if (data.next_frame != UINT32_C(3) || data.activated != UINT32_C(1)) {
		return 5;
	}
	if (collision[0] != 3.5f || collision[1] != -8.25f) {
		return 6;
	}
	return 0;
}
