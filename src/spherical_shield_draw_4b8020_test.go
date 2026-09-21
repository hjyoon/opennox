//go:build !server

package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

func TestDrawSphericalShield4B8020HighAddress(t *testing.T) {
	shield := &client.Drawable{PosVec: image.Pt(1, 2)}
	target := &client.Drawable{PosVec: image.Pt(300, 400)}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(shield)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(target)) <= uintptr(^uint32(0))) {
		t.Skipf("allocator returned a low drawable address: shield=%p target=%p", shield, target)
	}
	shield.UnionEffect().Field_108 = 0x9234
	moved, animated, deleted := 0, 0, 0
	got := drawSphericalShield4B8020(shield, sphericalShieldDrawHooks4B8020{
		byCode: func(code uint16) *client.Drawable {
			if code != 0x9234 {
				t.Fatalf("net code = %#x", code)
			}
			return target
		},
		move: func(dr *client.Drawable, x, y int) {
			moved++
			if dr != shield || x != 300 || y != 403 {
				t.Fatalf("move = %p (%d,%d)", dr, x, y)
			}
			dr.PosVec = image.Pt(x, y)
		},
		animate: func(dr *client.Drawable) int {
			animated++
			if dr != shield {
				t.Fatalf("animated %p, want %p", dr, shield)
			}
			return 7
		},
		delete: func(*client.Drawable) { deleted++ },
	})
	if got != 7 || moved != 1 || animated != 1 || deleted != 0 || shield.PosVec != image.Pt(300, 403) {
		t.Fatalf("result=%d moved=%d animated=%d deleted=%d pos=%v", got, moved, animated, deleted, shield.PosVec)
	}
}

func TestDrawSphericalShield4B8020MissingTarget(t *testing.T) {
	shield := &client.Drawable{}
	deleted := 0
	got := drawSphericalShield4B8020(shield, sphericalShieldDrawHooks4B8020{
		byCode: func(uint16) *client.Drawable { return nil },
		move: func(*client.Drawable, int, int) {
			t.Fatal("missing target moved shield")
		},
		animate: func(*client.Drawable) int {
			t.Fatal("missing target animated shield")
			return 0
		},
		delete: func(dr *client.Drawable) {
			deleted++
			if dr != shield {
				t.Fatalf("deleted %p, want %p", dr, shield)
			}
		},
	})
	if got != 0 || deleted != 1 {
		t.Fatalf("result=%d deletes=%d", got, deleted)
	}
}

func TestSphericalShieldDrawDispatch4B8020(t *testing.T) {
	fn := legacy.Get_nox_thing_spherical_shield_draw()
	if fn == nil {
		t.Fatal("spherical-shield callback pointer is nil")
	}
	dr := &client.Drawable{DrawFuncPtr: fn}
	if dr.DrawFuncPtr != fn {
		t.Fatal("spherical-shield callback pointer was not retained at native width")
	}
	dr.DrawFuncPtr = nil
	if got, ok := (*Client)(nil).callSphericalShieldDraw4B8020(dr, nil); ok || got != 0 {
		t.Fatalf("non-shield dispatch = (%d, %t)", got, ok)
	}
}
