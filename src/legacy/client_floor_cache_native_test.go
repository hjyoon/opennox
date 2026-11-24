package legacy

import (
	"image"
	"testing"

	"github.com/opennox/opennox/v1/client/noxrender"
)

func TestFloorVisibleCellsUsesNativeViewportCoordinates(t *testing.T) {
	for _, tc := range []struct {
		name  string
		world image.Rectangle
		want  [4]int32
	}{
		{"world relative to screen", image.Rect(103, 169, 743, 649), [4]int32{1, 2, 21, 19}},
		{"far edge", image.Rect(6000, 6100, 6640, 6580), [4]int32{106, 110, 127, 127}},
		{"negative origin", image.Rect(-75, -30, 565, 450), [4]int32{0, 0, 20, 17}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			vp := noxrender.Viewport{Screen: image.Rect(10, 20, 650, 500), World: tc.world}
			x0, y0, x1, y1 := floorVisibleCells((*nox_draw_viewport_t)(vp.C()), 21, 17, 128)
			if got := [4]int32{x0, y0, x1, y1}; got != tc.want {
				t.Fatalf("visible floor cells = %v, want %v", got, tc.want)
			}
		})
	}
}
