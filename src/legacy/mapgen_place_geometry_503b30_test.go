package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func TestMapgenPlaceGeometry503B30MatchesOriginalCornerOrder(t *testing.T) {
	// GAME.EXE 00503B8A..00503C9C transforms the original point, then the
	// both-offset, X-only and Y-only corners. 004D3C80 orders these as top,
	// left, right, bottom before 00428170 derives the clipped rectangle.
	got := mapgenPlaceGeometryFixture503B30(100, 200, 46, 92, 3000, 3000, 20, 30)
	if want := [8]int32{3201, 2994, 3169, 3026, 3266, 3059, 3234, 3091}; got.corners != want {
		t.Errorf("ordered corners = %v, want %v", got.corners, want)
	}
	if want := [4]int32{3169, 2994, 3266, 3091}; got.bounds != want {
		t.Errorf("bounds = %v, want %v", got.bounds, want)
	}
	if want := [2]int32{169, 26}; got.offset != want {
		t.Errorf("placement offset = %v, want %v", got.offset, want)
	}
	if want := [2]uint32{math.Float32bits(460), math.Float32bits(690)}; got.wallSpanBits != want {
		t.Errorf("wall span bits = %x, want %x", got.wallSpanBits, want)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && runtime.GOOS != "windows" && got.recordAddress <= 0xffffffff {
		t.Fatalf("wire record address = %#x, want above 4 GiB on this 64-bit host", got.recordAddress)
	}
}

func TestMapgenPlaceGeometry503B30ClampsAndWrapsQwordLow(t *testing.T) {
	got := mapgenPlaceGeometryFixture503B30(-10000, -10000, 46, 92,
		math.MinInt32, math.MaxInt32, 0x40000000, 0x80000000)
	if want := [8]int32{82, 2923, 82, 2956, 82, 2988, 82, 3021}; got.corners != want {
		t.Errorf("clamped corners = %v, want %v", got.corners, want)
	}
	if want := [4]int32{82, 2923, 82, 3021}; got.bounds != want {
		t.Errorf("clamped bounds = %v, want %v", got.bounds, want)
	}
	if want := [2]int32{-2147483566, -2147480691}; got.offset != want {
		t.Errorf("wrapped offsets = %v, want %v", got.offset, want)
	}
	if want := [2]uint32{math.Float32bits(-1073741824), math.Float32bits(-2147483648)}; got.wallSpanBits != want {
		t.Errorf("wrapped wall span bits = %x, want %x", got.wallSpanBits, want)
	}
}

func TestMapgenPlaceOffset503B30InvalidFistpUsesLowZero(t *testing.T) {
	cases := []struct {
		name   string
		fixed  float32
		origin int32
		want   int32
	}{
		{name: "signed qword low wrap", fixed: 0x1p31, want: math.MinInt32},
		{name: "positive qword overflow", fixed: 0x1p63},
		{name: "negative qword overflow", fixed: -0x1p63, origin: math.MaxInt32},
		{name: "NaN", fixed: float32(math.NaN())},
		{name: "positive infinity", fixed: float32(math.Inf(1))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := mapgenPlaceOffsetFixture503B30(tc.fixed, tc.origin); got != tc.want {
				t.Fatalf("offset(%v, %d) = %d, want %d", tc.fixed, tc.origin, got, tc.want)
			}
		})
	}
}

func TestMapgenFixCoords4D3D90PreservesOriginalX87Rounding(t *testing.T) {
	// GAME.EXE 004D3DA8..004D3DCC uses FLDS/FADDS/FMULS/FSTPS with the
	// binary32 coefficient at 00583B80. Rounding the sum to float32 before
	// multiplying shifts X by one ULP for this input.
	xBits, yBits, ok := mapgenFixedCoordsFixture4D3D90(2260.84130859375, 810.8399658203125)
	if !ok {
		t.Fatal("coordinate transform returned false")
	}
	if xBits != 0x45a0480e || yBits != 0x44f15637 {
		t.Fatalf("fixed coordinate bits = (%08x, %08x), want (45a0480e, 44f15637)", xBits, yBits)
	}
}

func TestMapgenFixCoords4D3D90ClampsUnorderedAndInfinite(t *testing.T) {
	for _, tc := range []struct {
		name         string
		atX, atY     float32
		wantX, wantY uint32
	}{
		{name: "NaN", atX: float32(math.NaN()), atY: 12, wantX: math.Float32bits(82.5), wantY: math.Float32bits(81.5)},
		{name: "positive infinity", atX: float32(math.Inf(1)), atY: 12, wantX: math.Float32bits(5851.5), wantY: math.Float32bits(81.5)},
		{name: "negative infinity", atX: float32(math.Inf(-1)), atY: 12, wantX: math.Float32bits(82.5), wantY: math.Float32bits(5852.5)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			xBits, yBits, ok := mapgenFixedCoordsFixture4D3D90(tc.atX, tc.atY)
			if !ok || xBits != tc.wantX || yBits != tc.wantY {
				t.Fatalf("fixed coordinate bits = (%08x, %08x, %t), want (%08x, %08x, true)", xBits, yBits, ok, tc.wantX, tc.wantY)
			}
		})
	}
}
