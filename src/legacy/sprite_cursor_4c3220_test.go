package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/client"
)

func TestSpriteHasState4C3220NativeDrawable(t *testing.T) {
	dr := &client.Drawable{NetCode32: 0x81234567}
	var gotPtr *client.Drawable
	var gotCode uint32
	got := spriteHasState4C3220(dr, func(code uint32) bool {
		gotPtr = dr
		gotCode = code
		return true
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if gotPtr != dr {
		t.Fatalf("drawable = %p, want native pointer %p", gotPtr, dr)
	}
	if gotCode != dr.NetCode32 {
		t.Fatalf("net code = %#x, want %#x", gotCode, dr.NetCode32)
	}
	if got := spriteHasState4C3220(dr, func(uint32) bool { return false }); got != 0 {
		t.Fatalf("missing result = %d, want 0", got)
	}
}
