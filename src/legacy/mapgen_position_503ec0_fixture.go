package legacy

/*
#include <stdint.h>
#include <string.h>

#include "GAME4.h"
#include "memmap.h"

extern uint32_t dword_5d4594_1599476;
extern uint32_t dword_5d4594_1599480;
extern uint32_t dword_5d4594_3835396;

typedef struct nox_test_mapgen_position_503ec0_result {
	uintptr_t object_address;
	int valid;
	uint32_t output_bits[2];
	uint32_t object_bits[2];
	int gated;
	uint32_t gated_output_bits[2];
	uint32_t gated_object_bits[2];
	int null_gated;
} nox_test_mapgen_position_503ec0_result;

static nox_test_mapgen_position_503ec0_result nox_test_mapgen_position_503ec0(
		float x, float y, int32_t origin_x, int32_t origin_y) {
	uint32_t old_loaded = dword_5d4594_1599480;
	uint32_t old_selected = dword_5d4594_3835396;
	uint32_t old_placed = dword_5d4594_1599476;
	int32_t old_origin_x = *getMemIntPtr(0x5D4594, 1599508);
	int32_t old_origin_y = *getMemIntPtr(0x5D4594, 1599512);

	nox_object_t object = {0};
	object.x = x;
	object.y = y;
	float2 output = {123.25f, -456.5f};
	nox_test_mapgen_position_503ec0_result result = {
		.object_address = (uintptr_t)&object,
	};

	dword_5d4594_1599480 = 17;
	dword_5d4594_3835396 = 17;
	dword_5d4594_1599476 = 0;
	*getMemIntPtr(0x5D4594, 1599508) = origin_x;
	*getMemIntPtr(0x5D4594, 1599512) = origin_y;
	result.valid = sub_503EC0(&object, &output);
	memcpy(&result.output_bits[0], &output.field_0, sizeof(output.field_0));
	memcpy(&result.output_bits[1], &output.field_4, sizeof(output.field_4));
	memcpy(&result.object_bits[0], &object.x, sizeof(object.x));
	memcpy(&result.object_bits[1], &object.y, sizeof(object.y));

	output.field_0 = 123.25f;
	output.field_4 = -456.5f;
	object.x = -999.0f;
	object.y = 9999.0f;
	dword_5d4594_1599476 = 1;
	result.gated = sub_503EC0(&object, &output);
	memcpy(&result.gated_output_bits[0], &output.field_0, sizeof(output.field_0));
	memcpy(&result.gated_output_bits[1], &output.field_4, sizeof(output.field_4));
	memcpy(&result.gated_object_bits[0], &object.x, sizeof(object.x));
	memcpy(&result.gated_object_bits[1], &object.y, sizeof(object.y));
	result.null_gated = sub_503EC0(NULL, NULL);

	dword_5d4594_1599480 = old_loaded;
	dword_5d4594_3835396 = old_selected;
	dword_5d4594_1599476 = old_placed;
	*getMemIntPtr(0x5D4594, 1599508) = old_origin_x;
	*getMemIntPtr(0x5D4594, 1599512) = old_origin_y;
	return result;
}
*/
import "C"

type mapgenPositionResult503EC0 struct {
	objectAddress   uintptr
	valid           bool
	outputBits      [2]uint32
	objectBits      [2]uint32
	gated           bool
	gatedOutputBits [2]uint32
	gatedObjectBits [2]uint32
	nullGated       bool
}

func mapgenPositionFixture503EC0(x, y float32, originX, originY int32) mapgenPositionResult503EC0 {
	v := C.nox_test_mapgen_position_503ec0(
		C.float(x), C.float(y), C.int32_t(originX), C.int32_t(originY),
	)
	return mapgenPositionResult503EC0{
		objectAddress:   uintptr(v.object_address),
		valid:           v.valid != 0,
		outputBits:      [2]uint32{uint32(v.output_bits[0]), uint32(v.output_bits[1])},
		objectBits:      [2]uint32{uint32(v.object_bits[0]), uint32(v.object_bits[1])},
		gated:           v.gated != 0,
		gatedOutputBits: [2]uint32{uint32(v.gated_output_bits[0]), uint32(v.gated_output_bits[1])},
		gatedObjectBits: [2]uint32{uint32(v.gated_object_bits[0]), uint32(v.gated_object_bits[1])},
		nullGated:       v.null_gated != 0,
	}
}
