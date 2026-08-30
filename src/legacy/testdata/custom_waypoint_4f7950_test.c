// Suppress unrelated Win32-only assertions while parsing the shared header,
// then verify GAME.EXE 004F7950/004F79A0/004F9A80/004F9AB0's native slots.
#define _Static_assert(...)
#include "../GAME4.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(offsetof(nox_player_update_data_t, custom_waypoints) ==
	(sizeof(void*) == 4 ? 168 : 200), "custom-waypoint array offset");
_Static_assert(sizeof(((nox_player_update_data_t*)0)->custom_waypoints) ==
	3 * sizeof(void*), "custom-waypoint array width");
_Static_assert(offsetof(nox_player_update_data_t, custom_waypoint_write) ==
	(sizeof(void*) == 4 ? 180 : 224), "custom-waypoint write-index offset");
_Static_assert(offsetof(nox_player_update_data_t, custom_waypoint_read) ==
	(sizeof(void*) == 4 ? 181 : 225), "custom-waypoint read-index offset");
_Static_assert(sizeof(nox_player_update_data_t) ==
	(sizeof(void*) == 4 ? 320 : 416), "partial PlayerUpdate size");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&sub_4F7950), void (*)(nox_object_t*)),
	"custom-waypoint cleanup signature");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&nox_xxx_playerSetCustomWP_4F79A0),
	void (*)(nox_object_t*, float, float)),
	"custom-waypoint setter signature");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&sub_4F9A80), int (*)(nox_object_t*)),
	"custom-waypoint presence signature");
_Static_assert(__builtin_types_compatible_p(
	__typeof__(&sub_4F9AB0), int (*)(nox_object_t*)),
	"custom-waypoint steering signature");

int main(void) {
	nox_player_update_data_t update = {0};
	nox_object_t waypoint = {0};
	update.custom_waypoint_read = UINT8_C(2);
	update.custom_waypoint_write = UINT8_C(1);
	update.custom_waypoints[2] = &waypoint;
	if (update.custom_waypoints[update.custom_waypoint_read] != &waypoint) {
		return 1;
	}
	if (sizeof(void*) == 8 && (uintptr_t)&waypoint <= UINT32_MAX) {
		return 2;
	}
	return 0;
}
