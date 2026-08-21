// Keep this fixture independent from the Win32-only aggregate legacy headers
// so every supported target frontend can compile the retained public ABI.
#include "../spark_init_4f0390.h"

#include <limits.h>
#include <stddef.h>
#include <stdint.h>

struct nox_object_t {
	nox_spark_update_data_t* update;
};

typedef nox_spark_update_data_t* (*spark_init_fn)(nox_object_t*);

_Static_assert(CHAR_BIT == 8, "bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(nox_spark_update_data_t) == 16, "Spark update-data size");
_Static_assert(offsetof(nox_spark_update_data_t, lifetime_initial) == 0,
	"Spark initial lifetime offset");
_Static_assert(offsetof(nox_spark_update_data_t, lifetime_remaining) == 4,
	"Spark remaining lifetime offset");
_Static_assert(offsetof(nox_spark_update_data_t, field_8) == 8,
	"Spark unrelated field offset");
_Static_assert(offsetof(nox_spark_update_data_t, kind) == 12,
	"Spark kind offset");
_Static_assert(
	_Generic(&nox_xxx_unitSparkInit_4F0390, spark_init_fn: 1, default: 0),
	"SparkInit must use native object/update pointers");

static nox_object_t* observed_unit;

nox_spark_update_data_t* nox_xxx_unitSparkInit_4F0390(nox_object_t* unit) {
	nox_spark_update_data_t* const update = unit->update;
	observed_unit = unit;
	update->lifetime_remaining = 32;
	update->lifetime_initial = 32;
	return update;
}

int main(void) {
	nox_spark_update_data_t update = {
		.lifetime_initial = UINT32_C(0x11111111),
		.lifetime_remaining = UINT32_C(0x22222222),
		.field_8 = UINT32_C(0xA5A5A5A5),
		.kind = UINT32_C(0x5A5A5A5A),
	};
	nox_object_t unit = {.update = &update};
	spark_init_fn const init = nox_xxx_unitSparkInit_4F0390;
	nox_spark_update_data_t* const result = init(&unit);

	if (observed_unit != &unit || result != &update)
		return __LINE__;
	if (update.lifetime_initial != 32 || update.lifetime_remaining != 32)
		return __LINE__;
	if (update.field_8 != UINT32_C(0xA5A5A5A5) ||
		update.kind != UINT32_C(0x5A5A5A5A))
		return __LINE__;
	return 0;
}
