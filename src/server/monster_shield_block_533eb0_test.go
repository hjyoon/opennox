package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func TestShieldBlockApproach533EB0OriginalPrecision(t *testing.T) {
	// Expected stores from GAME.EXE 00533EB0 at control 0x0e3f.
	facingEast := types.Ptf(1, 0)
	diagonal := math.Float32frombits(0x3f3504f3)
	facingSouthEast := types.Ptf(diagonal, diagonal)
	defender := types.Ptf(300, 300)
	for _, tc := range []struct {
		name                       string
		position, velocity, facing types.Pointf
		distanceBits               uint32
	}{
		{"head-on", types.Ptf(340, 300), types.Ptf(-4, 0), facingEast, 0x42200000},
		{"off-center", types.Ptf(340, 315.25), types.Ptf(-4, -1.5), facingEast, 0x422b3bd7},
		{"outside-lane", types.Ptf(340, 320), types.Ptf(-4, -2), facingEast, 0},
		{"inside-lane", types.Ptf(340, math.Float32frombits(math.Float32bits(320)-1)), types.Ptf(-4, -2), facingEast, 0x4232e2a8},
		{"rotated-defender", types.Ptf(320, 328), types.Ptf(-2, -3), facingSouthEast, 0x4209a31f},
		{"departing", types.Ptf(340, 300), types.Ptf(4, 0), facingEast, 0},
		{"stationary", types.Ptf(340, 300), types.Pointf{}, facingEast, 0},
		{"coincident", defender, types.Ptf(-4, 0), facingEast, 0},
		{"equal-aim", types.Ptf(340, 300), types.Ptf(-0.7, 0.7141428), facingEast, 0},
		{"outside-aim", types.Ptf(340, 300), types.Ptf(-0.6999999, 0.7141428), facingEast, 0},
		{"inside-aim", types.Ptf(340, 300), types.Ptf(-0.7000001, 0.7141428), facingEast, 0x42200000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			distance, incoming := ShieldBlockApproach533EB0(defender, tc.position, tc.velocity, tc.facing)
			if incoming != (tc.distanceBits != 0) || incoming && math.Float32bits(distance) != tc.distanceBits {
				t.Fatalf("incoming=%t distance bits=%08x, want %08x", incoming, math.Float32bits(distance), tc.distanceBits)
			}
		})
	}
}
