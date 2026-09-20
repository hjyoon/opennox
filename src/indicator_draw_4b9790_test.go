//go:build !server

package opennox

import (
	"image"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

func TestIndicatorDrawKindFor4B9790(t *testing.T) {
	tests := []struct {
		name string
		fn   unsafe.Pointer
		want indicatorDrawKind4B9790
	}{
		{"player waypoint", legacy.Get_nox_thing_player_waypoint_draw(), indicatorDrawPlayerWaypoint4B9790},
		{"pressure plate", legacy.Get_nox_thing_pressure_plate_draw(), indicatorDrawPressurePlate4BBB30},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.fn == nil {
				t.Fatal("callback pointer is nil")
			}
			if got := indicatorDrawKindFor4B9790(tc.fn); got != tc.want {
				t.Fatalf("kind = %d, want %d", got, tc.want)
			}
		})
	}
	if got := indicatorDrawKindFor4B9790(nil); got != indicatorDrawUnknown4B9790 {
		t.Fatalf("nil kind = %d", got)
	}
}

func TestPlayerWaypointDrawState4B9790HighAddress(t *testing.T) {
	dr := &client.Drawable{PosVec: image.Pt(150, 260)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	vp := &noxrender.Viewport{
		Screen: image.Rect(10, 20, 650, 500),
		World:  image.Rect(100, 200, 740, 680),
	}
	state := playerWaypointDrawStateFor4B9790(dr, vp, 32)
	if state.center != image.Pt(50, 60) {
		t.Fatalf("center = %v, want (50,60)", state.center)
	}
	if got, want := state.circle[0], (indicatorLine4B9790{
		from: image.Pt(50+10*sincosTable16[0].X/16, 60+10*sincosTable16[0].Y/16),
		to:   image.Pt(50+10*sincosTable16[16].X/16, 60+10*sincosTable16[16].Y/16),
	}); got != want {
		t.Fatalf("first circle line = %+v, want %+v", got, want)
	}
	if got, want := state.circle[15].to, state.circle[0].from; got != want {
		t.Fatalf("circle does not close: %v != %v", got, want)
	}
	if got, want := state.lines[0], (indicatorLine4B9790{
		from: image.Pt(50+10*sincosTable16[64].X/16, 60+10*sincosTable16[64].Y/16),
		to:   image.Pt(50+10*sincosTable16[166].X/16, 60+10*sincosTable16[166].Y/16),
	}); got != want {
		t.Fatalf("first line = %+v, want %+v", got, want)
	}
}

func TestPressurePlateDrawState4BBB30HighAddress(t *testing.T) {
	dr := &client.Drawable{PosVec: image.Pt(150, 260)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	dr.Shape.Box.LeftTop = -10.5
	dr.Shape.Box.LeftBottom = -20.5
	dr.Shape.Box.LeftBottom2 = -30.5
	dr.Shape.Box.LeftTop2 = 40.5
	dr.Shape.Box.RightTop = 50.5
	dr.Shape.Box.RightBottom = -60.5
	dr.Shape.Box.RightBottom2 = 70.5
	dr.Shape.Box.RightTop2 = 80.5
	dr.UnionEffect().Field_108 = 0x04030201
	dr.UnionEffect().Field_109 = 0x00000605
	vp := &noxrender.Viewport{
		Screen: image.Rect(10, 20, 650, 500),
		World:  image.Rect(100, 200, 740, 680),
	}
	state := pressurePlateDrawStateFor4BBB30(dr, vp)
	if !state.visible {
		t.Fatal("colored pressure plate is invisible")
	}
	if state.colors != [2]noxcolor.RGBA5551{
		noxcolor.RGB5551Color(1, 2, 3),
		noxcolor.RGB5551Color(4, 5, 6),
	} {
		t.Fatalf("colors = %#v", state.colors)
	}
	want := [4]indicatorLine4B9790{
		{from: image.Pt(50, 60), to: image.Pt(110, 20)},
		{from: image.Pt(130, 160), to: image.Pt(110, 20)},
		{from: image.Pt(50, 60), to: image.Pt(30, 120)},
		{from: image.Pt(130, 160), to: image.Pt(30, 120)},
	}
	if state.lines[0] != want {
		t.Fatalf("first layer = %+v, want %+v", state.lines[0], want)
	}
	for i, line := range state.lines[1] {
		if line.from != want[i].from.Add(image.Pt(1, 0)) || line.to != want[i].to.Add(image.Pt(1, 0)) {
			t.Fatalf("second layer line %d = %+v", i, line)
		}
	}

	dr.UnionEffect().Field_108 = 0
	dr.UnionEffect().Field_109 = 0
	if state := pressurePlateDrawStateFor4BBB30(dr, vp); state.visible {
		t.Fatal("uncolored pressure plate is visible")
	}
}

func TestPressurePlateFloatToInt4BBB30(t *testing.T) {
	tests := []struct {
		value float32
		want  int
	}{
		{1.5, 2},
		{2.5, 2},
		{-1.5, -2},
		{-2.5, -2},
	}
	for _, tc := range tests {
		if got := pressurePlateFloatToInt4BBB30(tc.value); got != tc.want {
			t.Fatalf("round(%v) = %d, want %d", tc.value, got, tc.want)
		}
	}
}
