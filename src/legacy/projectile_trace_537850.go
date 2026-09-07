package legacy

/*
#include <stdint.h>
#include "defs.h"

int sub_57CDB0(int2* grid, float* ray, float2* normal);

typedef struct projectile_wall_normal_result_537850 {
	int32_t ok;
	float x;
	float y;
} projectile_wall_normal_result_537850;

// Keep every pointer consumed by sub_57CDB0 inside C-owned stack storage.
// Only fixed-width scalar values cross the Go/C boundary.
static projectile_wall_normal_result_537850 projectile_wall_normal_scalar_537850(
	int32_t grid_x, int32_t grid_y,
	float from_x, float from_y, float to_x, float to_y
) {
	int2 grid;
	float ray[4];
	float2 normal = {0};
	projectile_wall_normal_result_537850 result;
	grid.field_0 = grid_x;
	grid.field_4 = grid_y;
	ray[0] = from_x;
	ray[1] = from_y;
	ray[2] = to_x;
	ray[3] = to_y;
	result.ok = sub_57CDB0(&grid, ray, &normal) != 0;
	result.x = normal.field_0;
	result.y = normal.field_4;
	return result;
}
*/
import "C"

import (
	"image"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
)

// StoreProjectileTraceGrid537850 mirrors GAME.EXE's fixed-width wall-grid
// output without carrying an Object pointer through C.
func StoreProjectileTraceGrid537850(point image.Point) {
	*memmap.PtrT[ntype.Point32](0x5D4594, 2488612) = ntype.Point32{
		X: int32(point.X),
		Y: int32(point.Y),
	}
}

// ProjectileWallNormal537850 calls the still-legacy wall normal helper through
// a scalar-only adapter. The ray rectangle stores origin in Min and the wall
// hit point in Max, matching the original float[4] layout.
func ProjectileWallNormal537850(grid image.Point, ray types.Rectf) (types.Pointf, bool) {
	result := C.projectile_wall_normal_scalar_537850(
		C.int32_t(grid.X),
		C.int32_t(grid.Y),
		C.float(ray.Min.X),
		C.float(ray.Min.Y),
		C.float(ray.Max.X),
		C.float(ray.Max.Y),
	)
	return types.Ptf(float32(result.x), float32(result.y)), result.ok != 0
}
