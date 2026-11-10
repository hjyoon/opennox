#include "game4_map_archive.h"

// Const/token-only ABI adapters. All arithmetic and state decisions run in Go.
FILE* sub_502DA0(char* name) {
	// This historical scalar success token is not the actual FILE* owner.
	// Keep address 1 out of Go's pointer roots; nox_file_8 is the real handle.
	return game4MapOpen(name) ? (FILE*)1 : NULL;
}

void nox_mapgenBuildPlaceBounds_503B30(const float2* at, const float2* fixed_at,
	const void* record, int32_t corners[8], int4* bounds) {
	game4MapBuildPlaceBounds((void*)at, (void*)fixed_at, (void*)record, corners, bounds);
}

int nox_mapgenPlaceGeometryNative_503B30(float x, float y, float fixed_x, float fixed_y,
	const void* record, int32_t map_origin_x, int32_t map_origin_y, int32_t bounds[4], int32_t delta[2]) {
	return game4MapPlaceGeometry(x, y, fixed_x, fixed_y, (void*)record,
		map_origin_x, map_origin_y, bounds, delta);
}
