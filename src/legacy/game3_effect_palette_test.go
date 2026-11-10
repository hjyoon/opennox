package legacy

import (
	"testing"
	"unsafe"
)

func TestGame3HarpoonGeometryOriginalC(t *testing.T) {
	// All axes, diagonal ties, quadrants, zero length, translated coordinates,
	// and short/long slopes are compared to execution of the untouched C body.
	for _, origin := range []game3LinePoint{{0, 0}, {-317, 901}, {4000, -5000}} {
		for _, delta := range []game3LinePoint{
			{0, 0}, {0, 17}, {0, -17}, {23, 0}, {-23, 0},
			{19, 19}, {-19, 19}, {19, -19}, {-19, -19},
			{1, 37}, {-1, 37}, {37, 1}, {37, -1}, {-37, -1}, {-1, -37},
		} {
			end := game3LinePoint{origin.x + delta.x, origin.y + delta.y}
			want, wantReturn, address := game3HarpoonOriginalControl(origin, end)
			if unsafe.Sizeof(uintptr(0)) > 4 && address <= 0xffffffff {
				t.Fatalf("original C control requires native high pointers: %#x", address)
			}
			var got [15]int32
			calls := 0
			returned := game3HarpoonLines(origin, end, func(a, b game3LinePoint, edge bool) int32 {
				if calls == 3 {
					t.Fatal("unexpected fourth line")
				}
				color := int32(0x1723)
				if edge {
					color = 0x4567
				}
				copy(got[5*calls:], []int32{a.x, a.y, b.x, b.y, color})
				calls++
				return 0x100 + int32(calls)
			})
			if calls != 3 || got != want || returned != wantReturn {
				t.Fatalf("line %v -> %v: calls=%d rows=%v return=%#x; original rows=%v return=%#x",
					origin, end, calls, got, returned, want, wantReturn)
			}
		}
	}
}
