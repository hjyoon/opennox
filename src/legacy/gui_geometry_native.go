package legacy

import (
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/ntype"
)

func guiGeometryStyle46AE10(win *gui.Window, enabled int32) *gui.Window {
	if win == nil {
		return nil
	}
	if enabled != 0 {
		win.DrawData().Field0 |= 2
	} else {
		win.DrawData().Field0 &^= 2
	}
	return win
}

func guiGeometryPointInRect4281F0(point *ntype.Point32, bounds *[4]int32) int32 {
	if point.X >= bounds[0] && point.X <= bounds[2] && point.Y >= bounds[1] && point.Y <= bounds[3] {
		return 1
	}
	return 0
}

func guiGeometryFloatInRect428220(point *[2]float32, bounds *[4]float32) int32 {
	// Stock x87 rejects unordered lower comparisons but accepts unordered
	// upper comparisons (TEST AH,0x41 after FCOMP/FNSTSW).
	if point[0] >= bounds[0] && !(point[0] > bounds[2]) && point[1] >= bounds[1] && !(point[1] > bounds[3]) {
		return 1
	}
	return 0
}

func guiGeometryToScreen4739E0(vp *noxrender.Viewport, point, out *ntype.Point32) int32 {
	out.X = point.X + int32(vp.Screen.Min.X) - int32(vp.World.Min.X)
	// Point and output can overlap. The original loads Point.Y after storing X.
	y := point.Y
	out.Y = y + int32(vp.Screen.Min.Y) - int32(vp.World.Min.Y)
	return y
}

func guiGeometryToWorld473A10(vp *noxrender.Viewport, point, out *ntype.Point32) int32 {
	out.X = point.X + int32(vp.World.Min.X) - int32(vp.Screen.Min.X)
	y := point.Y
	out.Y = y + int32(vp.World.Min.Y) - int32(vp.Screen.Min.Y)
	return y
}
