#ifndef NOX_TRAP_DOOR_COLLIDE_4EAB60_H
#define NOX_TRAP_DOOR_COLLIDE_4EAB60_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_trap_door_script_callback_t {
	uint32_t flags;
	int32_t func;
} nox_trap_door_script_callback_t;

typedef struct nox_trap_door_collide_data_t {
	nox_trap_door_script_callback_t script;
	int32_t fall_velocity_x;
	int32_t fall_velocity_y;
	uint32_t next_frame;
	uint16_t delay;
	uint16_t reserved_22;
	uint32_t activated;
} nox_trap_door_collide_data_t;

_Static_assert(sizeof(nox_trap_door_script_callback_t) == 8, "wrong size of TrapDoor script callback!");
_Static_assert(offsetof(nox_trap_door_script_callback_t, flags) == 0, "wrong offset of TrapDoor script flags!");
_Static_assert(offsetof(nox_trap_door_script_callback_t, func) == 4, "wrong offset of TrapDoor script function!");
_Static_assert(sizeof(nox_trap_door_collide_data_t) == 28, "wrong size of TrapDoor collide data!");
_Static_assert(offsetof(nox_trap_door_collide_data_t, script) == 0, "wrong offset of TrapDoor script block!");
_Static_assert(offsetof(nox_trap_door_collide_data_t, fall_velocity_x) == 8, "wrong offset of TrapDoor X velocity!");
_Static_assert(offsetof(nox_trap_door_collide_data_t, fall_velocity_y) == 12, "wrong offset of TrapDoor Y velocity!");
_Static_assert(offsetof(nox_trap_door_collide_data_t, next_frame) == 16, "wrong offset of TrapDoor next frame!");
_Static_assert(offsetof(nox_trap_door_collide_data_t, delay) == 20, "wrong offset of TrapDoor delay!");
_Static_assert(offsetof(nox_trap_door_collide_data_t, reserved_22) == 22, "wrong offset of TrapDoor reserved word!");
_Static_assert(offsetof(nox_trap_door_collide_data_t, activated) == 24, "wrong offset of TrapDoor activated flag!");

void nox_xxx_collideTrapDoor_4EAB60(nox_object_t* source, nox_object_t* target, float* collision);

#endif // NOX_TRAP_DOOR_COLLIDE_4EAB60_H
