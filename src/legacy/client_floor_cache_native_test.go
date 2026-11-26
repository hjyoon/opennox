package legacy

import (
	"image"
	"runtime"
	"testing"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
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

func TestFloorEdgeCallbackUsesTypedNativePointers(t *testing.T) {
	original := floorEdgeDraw
	defer Set_func_587000_154944(original)

	Set_func_587000_154944(cfixture.FloorEdgeCallback())
	runtime.GC()
	if got := cfixture.FloorEdgeCallbackCalls(); got != 2 {
		t.Fatalf("two typed native edge callbacks in list order, got %d", got)
	}
	Set_func_587000_154944(Get_nox_xxx_drawNoEdges_4818F0())
	if got := cfixture.FloorEdgeCallbackCalls(); got != 0 {
		t.Fatalf("draw-no-edges callback count = %d, want 0", got)
	}
	Set_func_587000_154944(nil)
	if got := cfixture.FloorEdgeCallbackCalls(); got != 0 {
		t.Fatalf("nil edge callback count = %d, want 0", got)
	}
}
