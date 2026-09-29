package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "memmap.h"

extern void nox_xxx_mapGenMakeInfo_4D5DB0(void* p);

typedef struct nox_test_mapgen_make_info_4D5DB0_result {
	void* data;
	uintptr_t source_address;
} nox_test_mapgen_make_info_4D5DB0_result;

static nox_test_mapgen_make_info_4D5DB0_result nox_test_mapgen_make_info_4D5DB0(void) {
	nox_test_mapgen_make_info_4D5DB0_result out = {0};
	unsigned char previous[0x5B8];
	void* source = getMemAt(0x973F18, 2408);
	void* data = calloc(1, 0x5B8);
	if (!source || !data) {
		free(data);
		return out;
	}
	memcpy(previous, source, sizeof(previous));
	memset(source, 0, sizeof(previous));
	nox_xxx_mapGenMakeInfo_4D5DB0(source);
	memcpy(data, source, sizeof(previous));
	memcpy(source, previous, sizeof(previous));
	out.data = data;
	out.source_address = (uintptr_t)source;
	return out;
}
*/
import "C"

// MapgenMakeInfoViaC4D5DB0 exercises the production C-to-Go ABI with the
// actual map-info buffer used by both random-map generation entry points.
func MapgenMakeInfoViaC4D5DB0() ([]byte, uintptr) {
	got := C.nox_test_mapgen_make_info_4D5DB0()
	if got.data == nil {
		return nil, uintptr(got.source_address)
	}
	defer C.free(got.data)
	return C.GoBytes(got.data, C.int(0x5B8)), uintptr(got.source_address)
}
