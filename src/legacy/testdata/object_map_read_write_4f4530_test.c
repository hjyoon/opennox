#include <assert.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>

#include "../object_map_read_write_4f4530.h"

struct nox_object_t {
	uintptr_t marker;
};

typedef int32_t (*object_map_read_write_fn_4F4530)(nox_object_t*, int32_t);

_Static_assert(CHAR_BIT == 8, "serialized bytes must remain eight bits");
_Static_assert(sizeof(void*) == 4 || sizeof(void*) == 8, "unsupported pointer width");
_Static_assert(sizeof(int32_t) == 4, "map version and result must remain int32");
_Static_assert(
	_Generic(&nox_xxx_mapReadWriteObjData_4F4530, object_map_read_write_fn_4F4530: 1, default: 0),
	"object map serializer must preserve its native object pointer and exact int32 scalar");

static nox_object_t* observed_object;
static int32_t observed_map_version;

int32_t nox_xxx_mapReadWriteObjData_4F4530(nox_object_t* object, int32_t map_version) {
	observed_object = object;
	observed_map_version = map_version;
	return map_version;
}

int main(void) {
	nox_object_t object = {UINTPTR_MAX};
	object_map_read_write_fn_4F4530 const serializer =
		nox_xxx_mapReadWriteObjData_4F4530;

	assert(serializer(&object, INT32_MIN) == INT32_MIN);
	assert(observed_object == &object);
	assert(observed_object->marker == UINTPTR_MAX);
	assert(observed_map_version == INT32_MIN);

	assert(serializer(NULL, INT32_MAX) == INT32_MAX);
	assert(observed_object == NULL);
	assert(observed_map_version == INT32_MAX);
	return 0;
}
