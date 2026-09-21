//go:build !server

package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

func TestArrowDrawKindFor4B7920(t *testing.T) {
	tests := []struct {
		name string
		fn   unsafe.Pointer
		want arrowDrawKind4B7920
	}{
		{"arrow", legacy.Get_nox_thing_arrow_draw(), arrowDrawStrong4B7920},
		{"weak arrow", legacy.Get_nox_thing_weak_arrow_draw(), arrowDrawWeak4B79D0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.fn == nil {
				t.Fatal("callback pointer is nil")
			}
			if got := arrowDrawKindFor4B7920(tc.fn); got != tc.want {
				t.Fatalf("kind = %d, want %d", got, tc.want)
			}
		})
	}
	if got := arrowDrawKindFor4B7920(nil); got != arrowDrawUnknown4B7920 {
		t.Fatalf("nil kind = %d", got)
	}
}

func TestDrawArrow4B7920HighAddress(t *testing.T) {
	drawCallback := legacy.Get_nox_thing_arrow_draw()
	updateCallback := legacy.Get_nox_xxx_updDrawMagicMissile_4CD9E0()
	dr := &client.Drawable{
		PosVec:              image.Pt(140, 260),
		Field_81:            uint32(100),
		Field_82:            uint32(200),
		DrawFuncPtr:         drawCallback,
		ClientUpdateFuncPtr: updateCallback,
	}
	tail := &client.Drawable{}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(tail)) <= uintptr(^uint32(0))) {
		t.Skip("allocator returned a low address")
	}

	vp := &noxrender.Viewport{}
	activated, decayed, slaveCalled := false, false, false
	hooks := arrowDrawHooks4B7920{
		typeID: func(name string) int {
			if name != "ArrowTailLink" {
				t.Fatalf("tail type = %q", name)
			}
			return 71
		},
		spawn: func(typeID int, pos image.Point) *client.Drawable {
			if typeID != 71 || pos != image.Pt(100, 200) {
				t.Fatalf("spawn = %d/%v", typeID, pos)
			}
			return tail
		},
		activate: func(got *client.Drawable) {
			if got != tail || decayed {
				t.Fatalf("activate = %p after decay=%t", got, decayed)
			}
			activated = true
		},
		decay: func(got *client.Drawable, lifetime int) {
			if got != tail || lifetime != 20 || !activated {
				t.Fatalf("decay = %p/%d after activate=%t", got, lifetime, activated)
			}
			decayed = true
		},
		slaveDraw: func(gotVP *noxrender.Viewport, got *client.Drawable) int {
			if gotVP != vp || got != dr || !decayed {
				t.Fatalf("slave draw = %p/%p after decay=%t", gotVP, got, decayed)
			}
			slaveCalled = true
			return 7
		},
		decayTime: 20,
	}

	if got := drawArrow4B7920(dr, vp, "ArrowTailLink", hooks); got != 7 {
		t.Fatalf("draw = %d, want 7", got)
	}
	if !activated || !decayed || !slaveCalled {
		t.Fatalf("activated=%t decayed=%t slave=%t", activated, decayed, slaveCalled)
	}
	if dr.Field_81 != 140 || dr.Field_82 != 260 {
		t.Fatalf("previous position = (%d, %d)", dr.Field_81, dr.Field_82)
	}
	if dr.DrawFuncPtr != drawCallback || dr.ClientUpdateFuncPtr != updateCallback {
		t.Fatalf("callbacks changed: draw=%p update=%p", dr.DrawFuncPtr, dr.ClientUpdateFuncPtr)
	}
	if effect := tail.UnionEffect(); effect.Field_108 != 140 || effect.Field_109 != 260 {
		t.Fatalf("tail endpoint = (%d, %d)", effect.Field_108, effect.Field_109)
	}
}

func TestDrawArrow4B7920ShortMoveAndSpawnFailure(t *testing.T) {
	t.Run("short move", func(t *testing.T) {
		dr := &client.Drawable{PosVec: image.Pt(105, 205), Field_81: 100, Field_82: 200}
		spawned := false
		got := drawArrow4B7920(dr, &noxrender.Viewport{}, "ArrowTailLink", arrowDrawHooks4B7920{
			typeID: func(string) int { return 1 },
			spawn: func(int, image.Point) *client.Drawable {
				spawned = true
				return nil
			},
			slaveDraw: func(*noxrender.Viewport, *client.Drawable) int { return 3 },
		})
		if got != 3 || spawned || dr.Field_81 != 100 || dr.Field_82 != 200 {
			t.Fatalf("draw=%d spawned=%t previous=(%d,%d)", got, spawned, dr.Field_81, dr.Field_82)
		}
	})

	t.Run("spawn failure", func(t *testing.T) {
		dr := &client.Drawable{PosVec: image.Pt(140, 260), Field_81: 100, Field_82: 200}
		activated, decayed := false, false
		got := drawArrow4B7920(dr, &noxrender.Viewport{}, "WeakArrowTailLink", arrowDrawHooks4B7920{
			typeID: func(name string) int {
				if name != "WeakArrowTailLink" {
					t.Fatalf("tail type = %q", name)
				}
				return 72
			},
			spawn:     func(int, image.Point) *client.Drawable { return nil },
			activate:  func(*client.Drawable) { activated = true },
			decay:     func(*client.Drawable, int) { decayed = true },
			slaveDraw: func(*noxrender.Viewport, *client.Drawable) int { return 5 },
		})
		if got != 5 || activated || decayed || dr.Field_81 != 100 || dr.Field_82 != 200 {
			t.Fatalf("draw=%d activated=%t decayed=%t previous=(%d,%d)", got, activated, decayed, dr.Field_81, dr.Field_82)
		}
	})
}
