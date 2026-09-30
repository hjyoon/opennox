package legacy

import (
	"image"

	"github.com/opennox/opennox/v1/client/noxrender"
)

// drawClipRectNative49F6F0 preserves the signed 32-bit SetRect/intersection
// arithmetic of 0049F6F0 while accessing the native-width render rectangles.
// An empty intersection leaves all state intact; success does not enable
// clipping and stores both the half-open image and inclusive primitive bounds.
func drawClipRectNative49F6F0(data *noxrender.RenderData, x, y, w, h int32) bool {
	rc := image.Rectangle{
		Min: image.Pt(int(x), int(y)),
		Max: image.Pt(int(x+w), int(y+h)),
	}
	rc = rc.Intersect(data.Rect3())
	if rc.Empty() {
		return false
	}
	data.SetClipRect(rc)
	rc.Max = rc.Max.Sub(image.Pt(1, 1))
	data.SetClipRect2(rc)
	return true
}
