package server

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/prand"
)

// The script namespace must use the same sealed 00416030 arithmetic as the
// original native logic stream, not prand's approximate decimal-scale Float.
func TestNoxScriptNSRandomFloat416030OriginalPiSeal(t *testing.T) {
	world := &Server{}
	world.Rand.Logic, world.Rand.Other = prand.New(0), prand.New(137)
	script := world.NoxScriptNS()
	words := make([]uint32, 4096)
	for index := range words {
		words[index] = math.Float32bits(script.RandomFloat(
			math.Float32frombits(0xc0490fdb), math.Float32frombits(0x40490fdb),
		))
		if got := world.Rand.Logic.Index(); got != (index+1)%4096 {
			t.Fatalf("logic index after word %d = %d", index, got)
		}
		if got := world.Rand.Other.Index(); got != 137 {
			t.Fatalf("script random float consumed the other stream: %d", got)
		}
	}
	if got := logicRandomFloatWordsSHA256_416030(words); got != logicRandomFloatPiOutputSHA256416030 {
		t.Fatalf("script pi-output SHA-256 = %s, want original %s", got, logicRandomFloatPiOutputSHA256416030)
	}
}

func TestNoxScriptNSRandomFloat416030OriginalGameplayPrecision(t *testing.T) {
	for _, bounds := range []struct {
		name     string
		min, max float32
	}{
		{"unit", 0, 1},
		{"negative-unit", -1, 0},
		{"reverse", 1, -1},
		{"motion", -2, 3},
		{"dense", math.Float32frombits(0xc2012345), math.Float32frombits(0x433abcde)},
		{"wide-exponents", math.SmallestNonzeroFloat32, math.MaxFloat32},
		{"negative-wide-exponents", -math.MaxFloat32, -math.SmallestNonzeroFloat32},
		{"opposite-extremes", -math.MaxFloat32, math.MaxFloat32},
		{"max-neighbors", math.Float32frombits(0x7f7ffffe), math.MaxFloat32},
		{"subnormal", math.Float32frombits(7), math.Float32frombits(0x00012345)},
		{"negative-zero-bound", math.Float32frombits(0x80000000), 12},
	} {
		t.Run(bounds.name, func(t *testing.T) {
			world := &Server{}
			world.Rand.Logic, world.Rand.Other = prand.New(0), prand.New(137)
			script, table := world.NoxScriptNS(), prand.New(0)
			for index := 0; index < 4096; index++ {
				word := table.Int(0, 0x7fff)
				want := float32(logicRandomFloatReferenceValue416030(bounds.min, bounds.max, word))
				got := script.RandomFloat(bounds.min, bounds.max)
				if math.Float32bits(got) != math.Float32bits(want) {
					t.Fatalf("word[%d]=%d: script=%08x, want original=%08x", index, word, math.Float32bits(got), math.Float32bits(want))
				}
				if got := world.Rand.Logic.Index(); got != (index+1)%4096 {
					t.Fatalf("logic index after word %d = %d", index, got)
				}
				if got := world.Rand.Other.Index(); got != 137 {
					t.Fatalf("other stream index after word %d = %d", index, got)
				}
			}
		})
	}
}

func TestNoxScriptNSRandomFloat416030ZeroAndUnorderedNoStep(t *testing.T) {
	for _, bounds := range [][2]uint32{
		{0, 0}, {0x80000000, 0x80000000}, {0x80000000, 0}, {0, 0x80000000},
		{0x3f800000, 0x3f800000}, {0x7f800000, 0x7f800000}, {0xff800000, 0xff800000},
		{0x7fc12345, 0x40000000}, {0x7f812345, 0x40000000},
		{0x3f800000, 0xffc12345}, {0x7fc12345, 0xffc12345},
	} {
		for _, index := range []int{17, 4095} {
			t.Run(fmt.Sprintf("%08x-%08x/at-%d", bounds[0], bounds[1], index), func(t *testing.T) {
				world := &Server{}
				world.Rand.Logic, world.Rand.Other = prand.New(index), prand.New(137)
				min, max := math.Float32frombits(bounds[0]), math.Float32frombits(bounds[1])
				got := world.NoxScriptNS().RandomFloat(min, max)
				if math.IsNaN(float64(max)) {
					if !math.IsNaN(float64(got)) {
						t.Fatalf("unordered max returned %v", got)
					}
				} else if math.Float32bits(got) != math.Float32bits(max) {
					t.Fatalf("no-step result=%08x, want max=%08x", math.Float32bits(got), math.Float32bits(max))
				}
				if got := world.Rand.Logic.Index(); got != index {
					t.Fatalf("no-step path advanced logic from %d to %d", index, got)
				}
				if got := world.Rand.Other.Index(); got != 137 {
					t.Fatalf("no-step path advanced other stream to %d", got)
				}
			})
		}
	}
}
