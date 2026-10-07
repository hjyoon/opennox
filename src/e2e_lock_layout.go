package opennox

import (
	"math"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

// This is placement input only. A perpendicular 128-unit offset keeps the
// outside Door farther from both idle casters while avoiding the stock Door's
// collision footprint overlap with the original tile*23 +/-34 group query.
func e2eLockLayout(origin, direction types.Pointf) [3]types.Pointf {
	center := origin.Add(direction.Mul(64))
	center = types.Ptf(float32(math.Round(float64(center.X)/23)*23), float32(math.Round(float64(center.Y)/23)*23))
	side := types.Ptf(-direction.Y, direction.X)
	return [3]types.Pointf{center, center.Add(direction.Mul(16)), center.Add(side.Mul(128))}
}

// Moderate stock-map tile coordinates make these integer bounds exact as
// float32. The query is an input/precondition check, not a supplied lock result.
func e2eLockGroupBounds(update *server.DoorUpdateData) types.Rectf {
	x, y := int32(uint32(update.TileX)*23), int32(uint32(update.TileY)*23)
	return types.Rectf{Min: types.Ptf(float32(int64(x)-34), float32(int64(y)-34)), Max: types.Ptf(float32(int64(x)+34), float32(int64(y)+34))}
}
