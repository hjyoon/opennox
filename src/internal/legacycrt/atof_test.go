package legacycrt

import (
	"math"
	"testing"
)

func TestAtofOriginalRounding(t *testing.T) {
	// Captured from GAME.EXE 005668EE with the startup x87 control word 0x0e7f.
	for _, tc := range []struct {
		text string
		f64  uint64
		f32  uint32
	}{
		{"-.", 0x0000000000000000, 0x00000000},
		{"-.00D99999", 0x8000000000000000, 0x80000000},
		{"\v1", 0x3ff0000000000000, 0x3f800000},
		{"\f1", 0x3ff0000000000000, 0x3f800000},
		{"1\x002", 0x3ff0000000000000, 0x3f800000},
		{"1.2D3", 0x4092c00000000000, 0x44960000},
		{"1e -2", 0x3ff0000000000000, 0x3f800000},
		{"0x1.8p+1", 0x0000000000000000, 0x00000000},
		{"1.1", 0x3ff199999999999a, 0x3f8ccccc},
		{"-0.1", 0xbfb999999999999a, 0xbdcccccc},
		{"1.4e-45", 0x369ff868bf4d956a, 0x00000000},
		{"3.4028234663852886e38", 0x47efffffe0000000, 0x7f7fffff},
		{"1e39", 0x48078287f49c4a1d, 0x7f7fffff},
		{"1E5201", 0x7ff0000000000000, 0x7f800000},
		{"-1E-9999999999", 0x8000000000000000, 0x80000000},
		{"-454797.10d15 suffix", 0xc438a793e0d4e9ce, 0xe1c53c9f},
		{"-2063.53018711D-315", 0x800000613eaa0d56, 0x80000000},
		{"3.0223145490365727689932800000000000000000000000000E+23", 0x44d0000000000000, 0x66800000},
		{"3.0948500982134505154577086545920000000000000000000E+26", 0x4570000000000000, 0x6b800000},
		{"9999999999999999999999990e-24", 0x4024000000000000, 0x41200000},
		{"1999999999999999999999999e-24", 0x4000000000000000, 0x40000000},
	} {
		if got := math.Float64bits(Atof(tc.text)); got != tc.f64 {
			t.Errorf("Atof(%q) = %#016x, want %#016x", tc.text, got, tc.f64)
		}
		if got := math.Float32bits(AtofFloat32(tc.text)); got != tc.f32 {
			t.Errorf("AtofFloat32(%q) = %#08x, want %#08x", tc.text, got, tc.f32)
		}
	}
}
