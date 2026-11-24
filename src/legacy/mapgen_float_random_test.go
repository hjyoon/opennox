package legacy

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestMapgenFloatRandomMatchesNativeArithmetic(t *testing.T) {
	random := rand.New(rand.NewPCG(0x526bc0, 0x526ac0))
	for i := range 4096 {
		draw := random.Uint32() & 0x7fff
		// Separate exponents exercise cancellation and float32 width rounding.
		low := math.Ldexp(float64(float32(int32(random.Uint32()))), random.IntN(240)-200)
		high := math.Ldexp(float64(float32(int32(random.Uint32()))), random.IntN(240)-200)
		got := mapgenFloatRandomFromDraw(draw, float32(low), float32(high))
		want := mapgenFloatRandomReference(draw, float32(low), float32(high))
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("row %d, draw %d, bounds %g %g: float64 bits = %016x, want %016x",
				i, draw, float32(low), float32(high), math.Float64bits(got), math.Float64bits(want))
		}
	}
}
