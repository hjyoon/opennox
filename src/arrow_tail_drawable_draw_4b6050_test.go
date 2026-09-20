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

func TestArrowTailDrawKindFor4B6050(t *testing.T) {
	tests := []struct {
		name string
		fn   unsafe.Pointer
		want arrowTailDrawKind4B6050
	}{
		{"arrow", legacy.Get_nox_thing_arrow_tail_link_draw(), arrowTailDrawStrong4B6050},
		{"weak arrow", legacy.Get_nox_thing_weak_arrow_tail_link_draw(), arrowTailDrawWeak4B6120},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.fn == nil {
				t.Fatal("callback pointer is nil")
			}
			if got := arrowTailDrawKindFor4B6050(tc.fn); got != tc.want {
				t.Fatalf("kind = %d, want %d", got, tc.want)
			}
		})
	}
	if got := arrowTailDrawKindFor4B6050(nil); got != arrowTailDrawUnknown4B6050 {
		t.Fatalf("nil kind = %d", got)
	}
}

func TestArrowTailDrawStateFor4B6050HighAddress(t *testing.T) {
	dr := &client.Drawable{
		PosVec:   image.Pt(150, 260),
		ZVal:     12,
		ZVal2:    4,
		Deadline: 130,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	dr.UnionEffect().Field_108 = uint32(120)
	dr.UnionEffect().Field_109 = uint32(240)
	vp := &noxrender.Viewport{
		Screen: image.Rect(10, 20, 650, 500),
		World:  image.Rect(100, 200, 740, 680),
	}
	state := arrowTailDrawStateFor4B6050(dr, vp, 100, 60)
	if !state.alive || state.point != image.Pt(60, 60) || state.tail != image.Pt(30, 40) || state.colorIndex != 32 {
		t.Fatalf("tail state = %+v", state)
	}
	if state := arrowTailDrawStateFor4B6050(dr, vp, 130, 60); state.alive {
		t.Fatalf("expired tail is alive: %+v", state)
	}
	if state := arrowTailDrawStateFor4B6050(dr, vp, 100, 0); state.alive {
		t.Fatalf("zero-span tail is alive: %+v", state)
	}
}
