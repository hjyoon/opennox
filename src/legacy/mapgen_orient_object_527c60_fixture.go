package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"

typedef struct nox_test_mapgen_orient_object_527c60_result {
	uintptr_t object_address;
	uintptr_t update_address;
	uint32_t object_token;
	int native_result[4];
	uint32_t native_frames[4][3];
	int invalid_direction_result;
	int missing_update_result;
	int legacy_result;
	uint32_t legacy_frames[3];
} nox_test_mapgen_orient_object_527c60_result;

static nox_test_mapgen_orient_object_527c60_result nox_test_mapgen_orient_object_527c60(void) {
	nox_test_mapgen_orient_object_527c60_result out = {0};
	nox_object_t* object = (nox_object_t*)calloc(1, sizeof(*object));
	uint32_t* update = (uint32_t*)calloc(100, sizeof(*update));
	if (!object || !update) {
		free(update);
		free(object);
		return out;
	}

	out.object_address = (uintptr_t)object;
	out.update_address = (uintptr_t)update;
	object->obj_class = 0x80;
	object->data_update = update;

	const int directions[4] = {1, 3, 5, 7};
	for (int i = 0; i < 4; ++i) {
		memset(update, 0xA5, 100 * sizeof(*update));
		out.native_result[i] = nox_mapgenOrientObjNative_527C60(object, directions[i]);
		out.native_frames[i][0] = update[1];
		out.native_frames[i][1] = update[2];
		out.native_frames[i][2] = update[3];
	}
	out.invalid_direction_result = nox_mapgenOrientObjNative_527C60(object, 2);
	object->data_update = NULL;
	out.missing_update_result = nox_mapgenOrientObjNative_527C60(object, 1);
	object->data_update = update;

	out.object_token = nox_mapgenLegacyPtrRegister(object);
	memset(update, 0, 100 * sizeof(*update));
	out.legacy_result = nox_xxx_mapGenOrientObj_527C60((int)out.object_token, 3);
	out.legacy_frames[0] = update[1];
	out.legacy_frames[1] = update[2];
	out.legacy_frames[2] = update[3];

	nox_mapgenLegacyPtrForget(object);
	free(update);
	free(object);
	return out;
}
*/
import "C"

type mapgenOrientObjectResult527C60 struct {
	objectAddress          uintptr
	updateAddress          uintptr
	objectToken            uint32
	nativeResult           [4]int
	nativeFrames           [4][3]uint32
	invalidDirectionResult int
	missingUpdateResult    int
	legacyResult           int
	legacyFrames           [3]uint32
}

func mapgenOrientObjectFixture527C60() mapgenOrientObjectResult527C60 {
	got := C.nox_test_mapgen_orient_object_527c60()
	out := mapgenOrientObjectResult527C60{
		objectAddress:          uintptr(got.object_address),
		updateAddress:          uintptr(got.update_address),
		objectToken:            uint32(got.object_token),
		invalidDirectionResult: int(got.invalid_direction_result),
		missingUpdateResult:    int(got.missing_update_result),
		legacyResult:           int(got.legacy_result),
	}
	for i := range out.nativeResult {
		out.nativeResult[i] = int(got.native_result[i])
		for j := range out.nativeFrames[i] {
			out.nativeFrames[i][j] = uint32(got.native_frames[i][j])
		}
	}
	for i := range out.legacyFrames {
		out.legacyFrames[i] = uint32(got.legacy_frames[i])
	}
	return out
}
