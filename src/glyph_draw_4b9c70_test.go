//go:build !server

package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func TestGlyphDrawState4B9C70(t *testing.T) {
	local := &client.Drawable{PosVec: image.Pt(100, 100)}
	dr := &client.Drawable{PosVec: image.Pt(100, 100)}
	if got := legacy.GlyphDrawStateFor(dr, local, true, 32); !got.Visible || got.Alpha != 200 || got.Colorize {
		t.Fatalf("near glyph state = %+v", got)
	}
	dr.PosVec = image.Pt(175, 100)
	if got := legacy.GlyphDrawStateFor(dr, local, true, 32); !got.Visible || got.Alpha != 150 || got.Colorize {
		t.Fatalf("mid-distance glyph state = %+v", got)
	}
	dr.PosVec = image.Pt(250, 100)
	if got := legacy.GlyphDrawStateFor(dr, local, true, 32); got.Visible {
		t.Fatalf("far glyph state = %+v", got)
	}
	dr.ObjFlags = 0x40000000
	if got := legacy.GlyphDrawStateFor(dr, local, true, 32); !got.Visible || got.Alpha != 255 {
		t.Fatalf("forced-visible glyph state = %+v", got)
	}
}

func TestGlyphDrawStateInfravision4B9C70(t *testing.T) {
	local := &client.Drawable{Buffs: 1 << server.ENCHANT_INFRAVISION}
	dr := &client.Drawable{PosVec: image.Pt(500, 500)}
	if got := legacy.GlyphDrawStateFor(dr, local, true, 32); !got.Visible || got.Alpha != 255 || !got.Colorize {
		t.Fatalf("high-depth infravision state = %+v", got)
	}
	if got := legacy.GlyphDrawStateFor(dr, local, true, 8); !got.Visible || got.Alpha != 128 || !got.Colorize {
		t.Fatalf("low-depth infravision state = %+v", got)
	}
}

func TestGlyphDrawDispatch4B9C70HighAddress(t *testing.T) {
	fn := legacy.Get_nox_thing_glyph_draw()
	if fn == nil {
		t.Fatal("glyph callback pointer is nil")
	}
	dr := &client.Drawable{DrawFuncPtr: fn}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low drawable address: %p", dr)
	}
	if dr.DrawFuncPtr != fn {
		t.Fatal("glyph callback pointer was not retained at native width")
	}
	dr.DrawFuncPtr = nil
	if got, ok := (*Client)(nil).callGlyphDraw4B9C70(dr, nil); ok || got != 0 {
		t.Fatalf("non-glyph dispatch = (%d, %t)", got, ok)
	}
}
