package legacy

import (
	"encoding/binary"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

func TestDrawableLightNativeABIOriginal(t *testing.T) {
	dr := new(client.Drawable)
	color := noxrender.RGB{R: -65537, G: 258, B: math.MinInt32}
	if ret := drawableLightColorCABI(dr, color); ret != dr || dr.LightFlags != 2 || dr.LightColor != color {
		t.Fatalf("color setter returned %p for %p, flags %#x, color %+v", ret, dr, dr.LightFlags, dr.LightColor)
	}
	if ret := drawableLightColorCABI(nil, color); ret != nil {
		t.Fatalf("color setter returned %p for nil drawable", ret)
	}

	// Outputs captured from the original PE32 code with the gameplay x87 control word.
	for _, tc := range []struct {
		degrees int32
		result  int64
	}{
		{-2147483648, -390937477119},
		{-1000000000, -182044448852},
		{-180001, -32768182},
		{-719, -130889},
		{-361, -65717},
		{-360, -65535},
		{-359, -65353},
		{-180, -32767},
		{-1, -181},
		{0, 0},
		{1, 182},
		{179, 32586},
		{180, 32768},
		{359, 65354},
		{360, 65536},
		{361, 65718},
		{719, 130890},
		{720, 131072},
		{10000, 1820444},
		{180001, 32768183},
		{1000000000, 182044448853},
		{2147483647, 390937476938},
	} {
		*dr = client.Drawable{LightPenumbra: 0x1234, Field_42: 0x4321, LightFlags: 9}
		if got := drawableLightAngleCABI(dr, tc.degrees, true); got != tc.result ||
			dr.LightDir != uint16(tc.result) || dr.LightPenumbra != 0x1234 || dr.Field_42 != 0 || dr.LightFlags != 9 {
			t.Fatalf("direction %d returned %d, dir %#x, size %#x, lock %#x, flags %#x",
				tc.degrees, got, dr.LightDir, dr.LightPenumbra, dr.Field_42, dr.LightFlags)
		}
		dr.Field_42 = 0x4321
		if got := drawableLightAngleCABI(dr, tc.degrees, false); got != tc.result ||
			dr.LightPenumbra != uint16(tc.result) || dr.Field_42 != 0x4321 || dr.LightFlags != 9 {
			t.Fatalf("size %d returned %d, size %#x, lock %#x, flags %#x",
				tc.degrees, got, dr.LightPenumbra, dr.Field_42, dr.LightFlags)
		}
	}
	for _, tc := range []struct {
		input, stored, fixed uint32
		radius               int
	}{
		{0xff7fffff, 0xff7fffff, 0x00000000, 0},
		{0xd7000000, 0xd7000000, 0x00000400, 0},
		{0xd6800000, 0xd6800000, 0x00000200, 0},
		{0xd1800000, 0xd1800000, 0x00000001, 0},
		{0xc2c80000, 0xc2c80000, 0xff9c0001, 0},
		{0xc0000000, 0xc0000000, 0xfffe0001, 0},
		{0xbfa00000, 0xbfa00000, 0xfffec001, 0},
		{0xbf000000, 0xbf000000, 0xffff8001, 0},
		{0xbe000000, 0xbe000000, 0xffffe001, 0},
		{0x80000000, 0x80000000, 0x00000000, 0},
		{0x00000000, 0x00000000, 0x00000000, 0},
		{0x3f000000, 0x3f000000, 0x00008000, 0},
		{0x3f800000, 0x3f800000, 0x00010000, 0},
		{0x3fa00000, 0x3fa00000, 0x00014000, 1},
		{0x40400000, 0x40400000, 0x00030000, 4},
		{0x41200000, 0x41200000, 0x000a0000, 19},
		{0x41f80000, 0x41f80000, 0x001f0000, 89},
		{0x42000000, 0x42000000, 0x00200000, 92},
		{0x4227d4fe, 0x4227d4fe, 0x0029f540, 121},
		{0x427bf5c2, 0x427bf5c2, 0x003efd71, 181},
		{0x427c0000, 0x427c0000, 0x003f0000, 181},
		{0x427c0106, 0x427c0000, 0x003f0000, 181},
		{0x42a20000, 0x427c0000, 0x003f0000, 181},
		{0x7f800000, 0x427c0000, 0x003f0000, 181},
		{0xff800000, 0xff800000, 0x00000000, 0},
		{0x3f800001, 0x3f800001, 0x00010000, 1},
		{0x427bffff, 0x427bffff, 0x003f0000, 181},
		{0x7fc30001, 0x7fc30001, 0x00000000, 0},
	} {
		*dr = client.Drawable{LightDir: 0xbabe, LightFlags: 0x4321, LightPenumbra: 97}
		var raw [17]byte
		block := raw[1:]
		binary.LittleEndian.PutUint32(block[:4], 0x4f020301)
		binary.LittleEndian.PutUint32(block[12:], 0x10203040)
		returned, packed := drawableLightIntensityCABI(dr, block, math.Float32frombits(tc.input))
		if returned != tc.radius || math.Float32bits(dr.LightIntensity) != tc.stored ||
			dr.LightIntensityU16 != tc.fixed || dr.LightIntensityRad != uint32(tc.radius) ||
			dr.LightFlags != 0x4321 || dr.LightDir != 0xbabe || dr.LightPenumbra != 97 {
			t.Fatalf("drawable intensity %#08x: returned %d, stored %#08x, fixed %#08x, radius %d",
				tc.input, returned, math.Float32bits(dr.LightIntensity), dr.LightIntensityU16, dr.LightIntensityRad)
		}
		if packed != tc.radius || binary.LittleEndian.Uint32(block[:4]) != 0x4f020301 ||
			binary.LittleEndian.Uint32(block[4:8]) != tc.stored || binary.LittleEndian.Uint32(block[8:12]) != uint32(tc.radius) ||
			binary.LittleEndian.Uint32(block[12:]) != 0x10203040 {
			t.Fatalf("packed intensity %#08x: returned %d, block %x", tc.input, packed, block)
		}
	}
}

func TestDrawableLightPackedColorPreservesAddressAndOtherFields(t *testing.T) {
	var raw [30]byte
	for i := range raw {
		raw[i] = byte(i + 1)
	}
	want := raw
	binary.LittleEndian.PutUint32(want[1:5], 2)
	binary.LittleEndian.PutUint32(want[17:21], 0xfffeffff)
	binary.LittleEndian.PutUint32(want[21:25], 0x00000102)
	binary.LittleEndian.PutUint32(want[25:29], 0x80000000)
	color := noxrender.RGB{R: -65537, G: 258, B: math.MinInt32}
	if pointer := drawablePackedLightColorCABI(&raw[1], color); pointer != unsafe.Pointer(&raw[1]) || raw != want {
		t.Fatalf("packed color returned %p for %p, bytes %x, want %x", pointer, &raw[1], raw, want)
	}
}

func TestDrawableLightRadiusOriginalIntegerBoundaries(t *testing.T) {
	// These are the original PE32 results where binary32 intermediates
	// used to round across an integer radius.
	for _, tc := range []struct {
		input  uint32
		radius int
	}{
		{0x3fae53e0, 1},
		{0x3ffdfa08, 2},
		{0x40e3403b, 12},
		{0x40f16e91, 13},
		{0x40ff580b, 14},
		{0x4113d282, 17},
		{0x4120bb0a, 19},
		{0x413f7bc1, 24},
		{0x4145684e, 25},
		{0x4156c94f, 28},
		{0x416d2326, 32},
		{0x417297d8, 33},
		{0x41780000, 34},
		{0x4183f8ba, 37},
		{0x41892d85, 39},
		{0x4195da1c, 44},
		{0x4198552d, 45},
		{0x419d3e74, 47},
		{0x41b730a6, 58},
		{0x41c9182d, 66},
		{0x41cb486f, 67},
		{0x41cfa114, 69},
		{0x41d61313, 72},
		{0x41e0a160, 77},
		{0x41e2b70a, 78},
		{0x41e8eb1a, 81},
		{0x41ed035f, 83},
		{0x41f11396, 85},
		{0x41ff1734, 91},
		{0x4203b454, 94},
		{0x42067a25, 96},
		{0x420aa2df, 99},
		{0x420ecb99, 102},
		{0x4215ba24, 107},
		{0x4219e2de, 110},
		{0x421e0b98, 113},
		{0x4224fa23, 118},
		{0x4227bff4, 120},
		{0x422922dd, 121},
		{0x422be8ae, 123},
		{0x42343a22, 129},
		{0x423862dc, 132},
		{0x423b28ad, 134},
		{0x423f5167, 137},
		{0x42421738, 139},
		{0x42437a21, 140},
		{0x424a68ac, 145},
		{0x424d2e7d, 147},
		{0x424e9166, 148},
		{0x42515737, 150},
		{0x4252ba20, 151},
		{0x42541d08, 152},
		{0x42557ff1, 153},
		{0x4259a8ab, 156},
		{0x425c6e7c, 158},
		{0x425dd165, 159},
		{0x42609736, 161},
		{0x4268e8aa, 167},
		{0x426bae7b, 169},
		{0x426d1164, 170},
		{0x426fd735, 172},
		{0x42729d06, 174},
		{0x4273ffef, 175},
		{0x4276c5c0, 177},
		{0x427828a9, 178},
	} {
		if got := client.LightRadius(math.Float32frombits(tc.input)); got != tc.radius {
			t.Errorf("intensity %#08x: radius = %d, want %d", tc.input, got, tc.radius)
		}
	}
}
