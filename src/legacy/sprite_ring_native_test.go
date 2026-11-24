package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestSpriteRing4C32A0UsesSignedFixedPointVertices(t *testing.T) {
	// Coefficients from the original PE32 table. Its routine shifts the int32 products with SAR.
	cases := [...]struct {
		index uintptr
		coeff [2]int32
		x, y  int32
	}{
		{0, [2]int32{16, 0}, 35, 0},
		{1, [2]int32{15, 6}, 32, 13},
		{5, [2]int32{-5, 15}, -11, 32},
		{8, [2]int32{-15, 0}, -33, 0},
		{10, [2]int32{-10, -10}, -22, -22},
		{15, [2]int32{15, -5}, 32, -11},
	}
	for _, c := range cases {
		entry := memmap.PtrT[[2]int32](0x587000, 192088+128*c.index)
		old := *entry
		t.Cleanup(func() { *entry = old })
		*entry = c.coeff
		x, y := spriteRingVertex(35, c.index)
		if x != c.x || y != c.y {
			t.Errorf("vertex %d = (%d, %d), want (%d, %d)", c.index, x, y, c.x, c.y)
		}
	}
}
