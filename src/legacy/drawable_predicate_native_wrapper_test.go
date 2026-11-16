//go:build !server

package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestDrawablePredicateNativeWrapper(t *testing.T) {
	drawablePredicateNativeRun(t, []int{3}, func(_ int, index int, owner unsafe.Pointer) int32 {
		dr := (*client.Drawable)(owner)
		switch index {
		case 0:
			return int32(Nox_xxx_sprite_4756E0_drawable(dr))
		case 1:
			return int32(Nox_xxx_sprite_475740_drawable(dr))
		case 2:
			return int32(Nox_xxx_sprite_4757A0_drawable(dr))
		case 3:
			return int32(Sub_4757D0_drawable(dr))
		default:
			panic("invalid original Drawable wrapper")
		}
	}, false)
	t.Log("Original Drawable wrappers:13460 contexts x3 placements x4 entries; whole64KiB readonly after every call")
}
