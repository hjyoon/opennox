package input

import (
	"image"
	"math"

	"github.com/opennox/libs/types"
)

// ScaleViewport converts a viewport measured in drawable pixels to the
// logical window coordinates used by mouse events. These coordinate systems
// differ on high-DPI displays (for example, a Retina drawable is commonly
// twice the logical SDL window size).
func ScaleViewport(view image.Rectangle, drawable, logical image.Point) image.Rectangle {
	if drawable.X <= 0 || drawable.Y <= 0 || logical.X <= 0 || logical.Y <= 0 || drawable == logical {
		return view
	}
	scale := func(v, from, to int) int {
		return int(math.Round(float64(v) * float64(to) / float64(from)))
	}
	return image.Rect(
		scale(view.Min.X, drawable.X, logical.X),
		scale(view.Min.Y, drawable.Y, logical.Y),
		scale(view.Max.X, drawable.X, logical.X),
		scale(view.Max.Y, drawable.Y, logical.Y),
	)
}

type window struct {
	view  image.Rectangle // viewport inside the window
	draw  image.Point     // size of the image that game draws
	scale types.Pointf    // scale factors calculated from sizes above
}

func (win *window) init(def image.Point) {
	win.view = image.Rectangle{Max: def}
	win.draw = def
	win.scale = types.Pointf{X: 1, Y: 1}
}

func (win *window) updateScale() {
	win.scale.X = float32(win.draw.X) / float32(win.view.Dx())
	win.scale.Y = float32(win.draw.Y) / float32(win.view.Dy())
}

// toDrawSpace remaps window position to position on the video buffer
func (win *window) toDrawSpace(p image.Point) image.Point {
	p.X -= win.view.Min.X
	p.Y -= win.view.Min.Y
	p.X = int(float32(p.X) * win.scale.X)
	p.Y = int(float32(p.Y) * win.scale.Y)
	return clamp(image.Rectangle{Max: win.draw}, p)
}

func clamp(r image.Rectangle, p image.Point) image.Point {
	if p.X < r.Min.X {
		p.X = r.Min.X
	}
	if p.Y < r.Min.Y {
		p.Y = r.Min.Y
	}
	if r.Max.X != 0 && p.X >= r.Max.X {
		p.X = r.Max.X
	}
	if r.Max.Y != 0 && p.Y >= r.Max.Y {
		p.Y = r.Max.Y
	}
	return p
}

func clampf(r image.Rectangle, p types.Pointf) types.Pointf {
	if p.X < float32(r.Min.X) {
		p.X = float32(r.Min.X)
	}
	if p.Y < float32(r.Min.Y) {
		p.Y = float32(r.Min.Y)
	}
	if r.Max.X != 0 && p.X >= float32(r.Max.X) {
		p.X = float32(r.Max.X)
	}
	if r.Max.Y != 0 && p.Y >= float32(r.Max.Y) {
		p.Y = float32(r.Max.Y)
	}
	return p
}

func (win *window) SetWinSize(rect image.Rectangle) {
	if rect.Dx() == 0 || rect.Dy() == 0 {
		return
	}
	win.view = rect
	win.updateScale()
}

func (win *window) SetDrawWinSize(sz image.Point) {
	if sz.X == 0 || sz.Y == 0 {
		return
	}
	win.draw = sz
	win.updateScale()
}
