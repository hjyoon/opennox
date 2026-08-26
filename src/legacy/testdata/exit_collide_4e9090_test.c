// Suppress unrelated Win32-only assertions while parsing the shared header,
// then assert the pointer-native callback and fixed 88-byte data boundary used
// by GAME.EXE 004E9090 and its registration record at 005CA238.
#define _Static_assert(...)
#include "../GAME3_3.h"
#undef _Static_assert

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_exit_collide_data_t) == 88, "ExitCollide data size");
_Static_assert(offsetof(nox_exit_collide_data_t, map_name) == 0,
	"ExitCollide map-name offset");
_Static_assert(offsetof(nox_exit_collide_data_t, destination_x) == 80,
	"ExitCollide destination-X offset");
_Static_assert(offsetof(nox_exit_collide_data_t, destination_y) == 84,
	"ExitCollide destination-Y offset");
_Static_assert(sizeof(((nox_object_t*)0)->obj_class) == 4,
	"object class width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_subclass) == 4,
	"object subclass width");
_Static_assert(sizeof(((nox_object_t*)0)->obj_flags) == 4,
	"object flags width");
_Static_assert(sizeof(((nox_object_t*)0)->inv_holder) == sizeof(void*),
	"inventory holder pointer width");
_Static_assert(sizeof(((nox_object_t*)0)->field_128) == sizeof(void*),
	"owned-list next pointer width");
_Static_assert(sizeof(((nox_object_t*)0)->field_129) == sizeof(void*),
	"owned-list head pointer width");
_Static_assert(sizeof(((nox_object_t*)0)->collide_data) == sizeof(void*),
	"collide-data pointer width");
_Static_assert(sizeof(((nox_object_t*)0)->data_update) == sizeof(void*),
	"update-data pointer width");

static nox_object_t* seen_exit;
static nox_object_t* seen_unit;
static float* seen_collision;

void nox_xxx_collideExit_4E9090(
	nox_object_t* exit,
	nox_object_t* unit,
	float* collision) {
	seen_exit = exit;
	seen_unit = unit;
	seen_collision = collision;
}

static void (*const exit_collide_signature)(nox_object_t*, nox_object_t*, float*) =
	nox_xxx_collideExit_4E9090;

int main(void) {
	nox_exit_collide_data_t data = {0};
	nox_object_t exit = {0};
	nox_object_t unit = {0};
	nox_object_t owned = {0};
	float collision[2] = {9.5f, -4.25f};

	data.map_name[0] = 'Q';
	exit.obj_subclass = UINT32_C(2);
	exit.collide_data = &data;
	unit.obj_class = UINT32_C(4);
	unit.field_129 = &owned;
	owned.field_128 = &exit;

	exit_collide_signature(&exit, &unit, collision);
	if (seen_exit != &exit || seen_unit != &unit || seen_collision != collision) {
		return 1;
	}
	if (exit.collide_data != &data || unit.field_129 != &owned ||
		owned.field_128 != &exit || data.map_name[0] != 'Q') {
		return 2;
	}

	exit_collide_signature(0, 0, 0);
	if (seen_exit != 0 || seen_unit != 0 || seen_collision != 0) {
		return 3;
	}
	return 0;
}
