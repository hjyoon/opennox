package legacycrt

import (
	"math"
	"strings"
	"testing"
)

func TestScanFloat32Original(t *testing.T) {
	// Observed at GAME.EXE 00536550 with the "%f %f" format.
	for _, tc := range []struct {
		text, rest string
		bits       uint32
		ok         bool
	}{
		{"1.2D3 4.1", "D3 4.1", 0x3f99999a, true},
		{"1e- 5", " 5", 0x3f800000, true},
		{"0x1.8p+1 2", "x1.8p+1 2", 0x00000000, true},
		{"  -infinity 3", "-infinity 3", 0x00000000, false},
		{" +1.4e-45 1.1", " 1.1", 0x00000001, true},
		{"1e39", "", 0x7f800000, true},
		{"5.877471754111437539843682686111228389093327783860e-39", "", 0x00000000, true},
		{strings.Repeat("0", 350) + "1 2.5", "01 2.5", 0x00000000, true},
	} {
		value, rest, ok := ScanFloat32(tc.text)
		if got := math.Float32bits(value); got != tc.bits || rest != tc.rest || ok != tc.ok {
			t.Errorf("ScanFloat32(%q) = %#08x, %q, %t, want %#08x, %q, %t", tc.text, got, rest, ok, tc.bits, tc.rest, tc.ok)
		}
	}
}
