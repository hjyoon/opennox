package legacy

/*
#include <stdint.h>
#include <string.h>

#include "GAME3_2.h"
#include "GAME4.h"

typedef struct nox_test_mapgen_geometry_503b30_result {
	uintptr_t record_address;
	int32_t corners[8];
	int32_t bounds[4];
	int32_t offset[2];
	uint32_t wall_span_bits[2];
} nox_test_mapgen_geometry_503b30_result;

typedef struct nox_test_mapgen_fixed_coords_4d3d90_result {
	uint32_t x_bits;
	uint32_t y_bits;
	int ok;
} nox_test_mapgen_fixed_coords_4d3d90_result;

static nox_test_mapgen_fixed_coords_4d3d90_result nox_test_mapgen_fixed_coords_4d3d90(float x, float y) {
	float2 at = {x, y};
	float2 fixed = {0};
	nox_test_mapgen_fixed_coords_4d3d90_result out = {0};
	out.ok = nox_xxx_mapGenFixCoords_4D3D90(&at, &fixed);
	memcpy(&out.x_bits, &fixed.field_0, sizeof(fixed.field_0));
	memcpy(&out.y_bits, &fixed.field_4, sizeof(fixed.field_4));
	return out;
}

static nox_test_mapgen_geometry_503b30_result nox_test_mapgen_geometry_503b30(
		float x, float y, float width, float height, int32_t map_x, int32_t map_y,
		uint32_t wall_x, uint32_t wall_y) {
	uint8_t record[76] = {0};
	memcpy(record + 64, &width, sizeof(width));
	memcpy(record + 68, &height, sizeof(height));
	float2 at = {x, y};
	float2 fixed = {0};
	nox_xxx_mapGenFixCoords_4D3D90(&at, &fixed);
	nox_test_mapgen_geometry_503b30_result out = {.record_address = (uintptr_t)record};
	int4 bounds = {0};
	nox_mapgenBuildPlaceBounds_503B30(&at, &fixed, record, out.corners, &bounds);
	out.bounds[0] = bounds.field_0;
	out.bounds[1] = bounds.field_4;
	out.bounds[2] = bounds.field_8;
	out.bounds[3] = bounds.field_C;
	out.offset[0] = nox_mapgenPlaceOffset_503B30(fixed.field_0, map_x);
	out.offset[1] = nox_mapgenPlaceOffset_503B30(fixed.field_4, map_y);
	float span_x = nox_mapgenWallSpan_503B30(wall_x);
	float span_y = nox_mapgenWallSpan_503B30(wall_y);
	memcpy(&out.wall_span_bits[0], &span_x, sizeof(span_x));
	memcpy(&out.wall_span_bits[1], &span_y, sizeof(span_y));
	return out;
}
*/
import "C"

type mapgenPlaceGeometryResult503B30 struct {
	recordAddress uintptr
	corners       [8]int32
	bounds        [4]int32
	offset        [2]int32
	wallSpanBits  [2]uint32
}

func mapgenPlaceGeometryFixture503B30(x, y, width, height float32, mapX, mapY int32, wallX, wallY uint32) mapgenPlaceGeometryResult503B30 {
	v := C.nox_test_mapgen_geometry_503b30(C.float(x), C.float(y), C.float(width), C.float(height),
		C.int32_t(mapX), C.int32_t(mapY), C.uint32_t(wallX), C.uint32_t(wallY))
	out := mapgenPlaceGeometryResult503B30{recordAddress: uintptr(v.record_address)}
	for i := range out.corners {
		out.corners[i] = int32(v.corners[i])
	}
	for i := range out.bounds {
		out.bounds[i] = int32(v.bounds[i])
	}
	for i := range out.offset {
		out.offset[i] = int32(v.offset[i])
		out.wallSpanBits[i] = uint32(v.wall_span_bits[i])
	}
	return out
}

func mapgenPlaceOffsetFixture503B30(fixed float32, origin int32) int32 {
	return int32(C.nox_mapgenPlaceOffset_503B30(C.float(fixed), C.int32_t(origin)))
}

func mapgenFixedCoordsFixture4D3D90(x, y float32) (xBits, yBits uint32, ok bool) {
	v := C.nox_test_mapgen_fixed_coords_4d3d90(C.float(x), C.float(y))
	return uint32(v.x_bits), uint32(v.y_bits), v.ok != 0
}
