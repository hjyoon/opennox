// Keep this fixture independent from the Win32-only aggregate legacy headers
// so every supported target frontend can compile the retained public ABI.
#include "../skull_init_4f0450.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>

typedef struct direction_init_data {
	int32_t x;
	int32_t y;
} direction_init_data;

typedef struct skull_update_data {
	uint32_t scan_delay;
	uint32_t fire_delay;
	uint8_t target_ready;
	uint8_t field_9[3];
	uint32_t projectile_type;
	char projectile_name[32];
	uint8_t enabled;
	uint8_t field_49[3];
} skull_update_data;

struct nox_object_t {
	uint16_t direction_1;
	uint16_t direction_2;
	direction_init_data* init_data;
	skull_update_data* update_data;
};

typedef int32_t (*skull_init_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(int32_t) == 4, "direction components must remain signed dwords");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(direction_init_data) == 8, "direction init-data size");
_Static_assert(offsetof(direction_init_data, x) == 0, "direction X offset");
_Static_assert(offsetof(direction_init_data, y) == 4, "direction Y offset");
_Static_assert(sizeof(skull_update_data) == 52, "Skull update-data size");
_Static_assert(offsetof(skull_update_data, projectile_type) == 12, "projectile type offset");
_Static_assert(offsetof(skull_update_data, projectile_name) == 16, "projectile name offset");
_Static_assert(offsetof(skull_update_data, enabled) == 48, "enabled offset");
_Static_assert(
	_Generic(&nox_xxx_unitSkullInit_4F0450, skull_init_fn: 1, default: 0),
	"SkullInit must use one native object pointer and return an exact int32_t");

static char const* observed_name;
static unsigned int observed_lookups;

static int32_t lookup_type(char const* name) {
	observed_name = name;
	++observed_lookups;
	return INT32_MIN + 1;
}

int32_t nox_xxx_unitSkullInit_4F0450(nox_object_t* unit) {
	static uint32_t const table[9] = {
		UINT32_C(160), UINT32_C(192), UINT32_C(224),
		UINT32_C(128), UINT32_C(0), UINT32_C(0),
		UINT32_C(96), UINT32_C(64), UINT32_C(32),
	};
	direction_init_data* const init_data = unit->init_data;
	skull_update_data* const update_data = unit->update_data;
	int32_t const index = init_data->x + INT32_C(3) * init_data->y;
	uint32_t const angle = table[index + INT32_C(4)];
	int32_t projectile_type;

	unit->direction_2 = (uint16_t)angle;
	unit->direction_1 = (uint16_t)angle;
	projectile_type = lookup_type(update_data->projectile_name);
	update_data->projectile_type = (uint32_t)projectile_type;
	return projectile_type;
}

int main(void) {
	direction_init_data init_data = {.x = INT32_C(1), .y = -INT32_C(1)};
	skull_update_data update_data = {
		.scan_delay = UINT32_C(0x11111111),
		.fire_delay = UINT32_C(0x22222222),
		.target_ready = UINT8_C(0x33),
		.field_9 = {UINT8_C(0x41), UINT8_C(0x42), UINT8_C(0x43)},
		.projectile_type = UINT32_C(0x44444444),
		.projectile_name = "MercArcherArrow",
		.enabled = UINT8_C(0x55),
		.field_49 = {UINT8_C(0x61), UINT8_C(0x62), UINT8_C(0x63)},
	};
	nox_object_t unit = {
		.direction_1 = UINT16_C(0xAAAA),
		.direction_2 = UINT16_C(0xBBBB),
		.init_data = &init_data,
		.update_data = &update_data,
	};
	skull_init_fn const init = nox_xxx_unitSkullInit_4F0450;
	int32_t const result = init(&unit);

	if (result != INT32_MIN + 1 || observed_lookups != 1)
		return __LINE__;
	if (observed_name != update_data.projectile_name || strcmp(observed_name, "MercArcherArrow") != 0)
		return __LINE__;
	if (unit.direction_1 != UINT16_C(224) || unit.direction_2 != UINT16_C(224))
		return __LINE__;
	if (update_data.projectile_type != UINT32_C(0x80000001))
		return __LINE__;
	if (update_data.scan_delay != UINT32_C(0x11111111) ||
		update_data.fire_delay != UINT32_C(0x22222222) ||
		update_data.target_ready != UINT8_C(0x33) || update_data.enabled != UINT8_C(0x55))
		return __LINE__;
	if (update_data.field_9[0] != UINT8_C(0x41) || update_data.field_9[1] != UINT8_C(0x42) ||
		update_data.field_9[2] != UINT8_C(0x43) || update_data.field_49[0] != UINT8_C(0x61) ||
		update_data.field_49[1] != UINT8_C(0x62) || update_data.field_49[2] != UINT8_C(0x63))
		return __LINE__;
	return 0;
}
