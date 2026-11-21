package legacy

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

// ProjectileWallNormal537850 resolves the wall struck by the ray.
func ProjectileWallNormal537850(grid image.Point, ray types.Rectf) (types.Pointf, bool) {
	return GetServer().S().ProjectileWallNormal(grid, ray)
}
