//go:build !server

package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

func TestBoulderDrawNativeState4B9B50HighAddress(t *testing.T) {
	dr := &client.Drawable{PosVec: image.Pt(100, 200)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	if got := boulderDrawFrame4B9B50(dr); got != 0 {
		t.Fatalf("initial image = %d, want 0", got)
	}
	effect := dr.UnionEffect()
	if effect.Field_108 != 100 || effect.Field_109 != 200 {
		t.Fatalf("initial previous position = (%d, %d), want (100, 200)", effect.Field_108, effect.Field_109)
	}

	dr.PosVec = image.Pt(103, 204)
	effect.Field_110 = 7
	effect.Field_111 = 16
	if got := boulderDrawFrame4B9B50(dr); got != 23 {
		t.Fatalf("short movement image = %d, want 23", got)
	}
	if effect.Field_108 != 100 || effect.Field_109 != 200 {
		t.Fatalf("short movement changed previous position to (%d, %d)", effect.Field_108, effect.Field_109)
	}
}

func TestBoulderDrawRollingQuadrants4B9B50(t *testing.T) {
	tests := []struct {
		name      string
		pos       image.Point
		frame     uint32
		wantFrame uint32
		wantBank  uint32
		wantImage int
	}{
		{"left-down increments and wraps", image.Pt(90, 111), 15, 0, 0, 0},
		{"left-up decrements and wraps", image.Pt(90, 89), 0, 15, 16, 31},
		{"right-down increments and wraps", image.Pt(111, 111), 15, 0, 16, 16},
		{"right-up decrements and wraps", image.Pt(111, 89), 0, 15, 0, 15},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dr := &client.Drawable{PosVec: tc.pos}
			effect := dr.UnionEffect()
			effect.Field_108 = 100
			effect.Field_109 = 100
			effect.Field_110 = tc.frame
			if got := boulderDrawFrame4B9B50(dr); got != tc.wantImage {
				t.Fatalf("image = %d, want %d", got, tc.wantImage)
			}
			if effect.Field_110 != tc.wantFrame || effect.Field_111 != tc.wantBank {
				t.Fatalf("animation state = (%d, %d), want (%d, %d)", effect.Field_110, effect.Field_111, tc.wantFrame, tc.wantBank)
			}
			if effect.Field_108 != uint32(tc.pos.X) || effect.Field_109 != uint32(tc.pos.Y) {
				t.Fatalf("previous position = (%d, %d), want %v", effect.Field_108, effect.Field_109, tc.pos)
			}
		})
	}
}

func TestBoulderDrawDispatch4B9B50(t *testing.T) {
	fn := legacy.Get_nox_thing_boulder_draw()
	if fn == nil {
		t.Fatal("boulder callback pointer is nil")
	}
	dr := &client.Drawable{DrawFuncPtr: fn, PosVec: image.Pt(12, 34)}
	if got, ok := (*Client)(nil).callBoulderDraw4B9B50(dr, nil); !ok || got != 1 {
		t.Fatalf("boulder dispatch = (%d, %t), want (1, true)", got, ok)
	}
	dr.DrawFuncPtr = nil
	if got, ok := (*Client)(nil).callBoulderDraw4B9B50(dr, nil); ok || got != 0 {
		t.Fatalf("non-boulder dispatch = (%d, %t), want (0, false)", got, ok)
	}
}
