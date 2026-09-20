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

func TestSimpleProjectileDrawKindFor4B9D70(t *testing.T) {
	tests := []struct {
		name string
		fn   unsafe.Pointer
		want simpleProjectileDrawKind4B9D70
	}{
		{"spider spit", legacy.Get_nox_thing_spider_spit_draw(), simpleProjectileDrawSpiderSpit4B9D70},
		{"black powder", legacy.Get_nox_thing_black_powder_draw(), simpleProjectileDrawBlackPowder4B9ED0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.fn == nil {
				t.Fatal("callback pointer is nil")
			}
			if got := simpleProjectileDrawKindFor4B9D70(tc.fn); got != tc.want {
				t.Fatalf("kind = %d, want %d", got, tc.want)
			}
		})
	}
	if got := simpleProjectileDrawKindFor4B9D70(nil); got != simpleProjectileDrawUnknown4B9D70 {
		t.Fatalf("nil kind = %d", got)
	}
}

func TestSimpleProjectileDrawGeometryHighAddress(t *testing.T) {
	dr := &client.Drawable{
		PosVec:  image.Pt(150, 260),
		Field_8: uint32(120),
		Field_9: uint32(240),
		ZVal:    12,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	vp := &noxrender.Viewport{
		Screen: image.Rect(10, 20, 650, 500),
		World:  image.Rect(100, 200, 740, 680),
	}
	if got := blackPowderPoint4B9ED0(dr, vp); got != image.Pt(60, 80) {
		t.Fatalf("black powder point = %v", got)
	}

	state := spiderSpitDrawStateFor4B9D70(dr, vp)
	wantHorizontal := spiderSpitDrawState4B9D70{
		point: image.Pt(60, 68),
		tail:  image.Pt(30, 48),
		grayLines: [2]projectileLine4B9D70{
			{from: image.Pt(60, 69), to: image.Pt(30, 49)},
			{from: image.Pt(60, 69), to: image.Pt(30, 47)},
		},
	}
	if state != wantHorizontal {
		t.Fatalf("horizontal state = %+v, want %+v", state, wantHorizontal)
	}

	dr.Field_8 = uint32(155)
	dr.Field_9 = uint32(200)
	state = spiderSpitDrawStateFor4B9D70(dr, vp)
	wantVertical := spiderSpitDrawState4B9D70{
		point: image.Pt(60, 68),
		tail:  image.Pt(65, 8),
		grayLines: [2]projectileLine4B9D70{
			{from: image.Pt(61, 68), to: image.Pt(64, 8)},
			{from: image.Pt(59, 68), to: image.Pt(64, 8)},
		},
	}
	if state != wantVertical {
		t.Fatalf("vertical state = %+v, want %+v", state, wantVertical)
	}
}
