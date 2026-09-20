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

func TestMagicDrawableDrawKindFor4B98A0(t *testing.T) {
	tests := []struct {
		name string
		fn   unsafe.Pointer
		want magicDrawableDrawKind4B98A0
	}{
		{"magic", legacy.Get_nox_thing_magic_draw(), magicDrawableDrawOrb4B98A0},
		{"missile", legacy.Get_nox_thing_magic_missle_draw(), magicDrawableDrawMissile4B99F0},
		{"tail", legacy.Get_nox_thing_magic_tail_link_draw(), magicDrawableDrawTail4B5E10},
		{"missile tail", legacy.Get_nox_thing_magic_missle_tail_link_draw(), magicDrawableDrawMissileTail4B5F30},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.fn == nil {
				t.Fatal("callback pointer is nil")
			}
			if got := magicDrawableDrawKindFor4B98A0(tc.fn); got != tc.want {
				t.Fatalf("kind = %d, want %d", got, tc.want)
			}
		})
	}
	if got := magicDrawableDrawKindFor4B98A0(nil); got != magicDrawableDrawUnknown4B98A0 {
		t.Fatalf("nil kind = %d", got)
	}
}

func TestMagicDrawNativeGeometry4B98A0HighAddress(t *testing.T) {
	dr := &client.Drawable{
		PosVec: image.Pt(150, 260),
		ZVal:   12,
		ZVal2:  4,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	vp := &noxrender.Viewport{
		Screen: image.Rect(10, 20, 650, 500),
		World:  image.Rect(100, 200, 740, 680),
	}
	if got, want := magicDrawPoint4B98A0(dr, vp), image.Pt(60, 64); got != want {
		t.Fatalf("draw point = %v, want %v", got, want)
	}
}

func TestMagicTailDrawStateFor4B5E10HighAddress(t *testing.T) {
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
	state := magicTailDrawStateFor4B5E10(dr, vp, 100, 60)
	if !state.alive || state.point != image.Pt(60, 64) || state.tail != image.Pt(30, 44) ||
		state.colorIndex != 32 || state.intensity != 10 {
		t.Fatalf("tail state = %+v", state)
	}
	state = magicTailDrawStateFor4B5E10(dr, vp, 120, 20)
	if !state.alive || state.colorIndex != 32 || state.intensity != 10 {
		t.Fatalf("missile tail state = %+v", state)
	}
	if state := magicTailDrawStateFor4B5E10(dr, vp, 130, 60); state.alive {
		t.Fatalf("expired tail is alive: %+v", state)
	}
	if state := magicTailDrawStateFor4B5E10(dr, vp, 100, 0); state.alive {
		t.Fatalf("zero-span tail is alive: %+v", state)
	}
}
