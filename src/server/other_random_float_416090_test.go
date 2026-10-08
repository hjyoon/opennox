package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/prand"
)

// Use the existing independent big.Float/ToZero instruction reference for
// finite values. IEEE exceptional operations are checked by result class,
// not a platform-specific x87/ARM NaN payload.
func otherRandomFloatReferenceOp416090(a, b float64, operation byte) float64 {
	if !math.IsNaN(a) && !math.IsNaN(b) && !math.IsInf(a, 0) && !math.IsInf(b, 0) {
		return logicRandomFloatReference53_416030(a, b, operation)
	}
	switch operation {
	case '-':
		return a - b
	case '*':
		return a * b
	default:
		return a + b
	}
}

func otherRandomFloatReference416090(min, max float32, word int) float64 {
	delta := otherRandomFloatReferenceOp416090(float64(max), float64(min), '-')
	scaled := otherRandomFloatReferenceOp416090(float64(word), float64(math.Float32frombits(0x38000100)), '*')
	return otherRandomFloatReferenceOp416090(otherRandomFloatReferenceOp416090(delta, scaled, '*'), float64(min), '+')
}

func TestOtherRandomFloat416090OriginalGameplayPrecision(t *testing.T) {
	for _, bounds := range []struct {
		name     string
		min, max float32
	}{
		{"pi", math.Float32frombits(0xc0490fdb), math.Float32frombits(0x40490fdb)},
		{"light", 0, 100}, {"unit", 0, 1}, {"negative-unit", -1, 0},
		{"reverse-unit", 1, -1}, {"motion", -2, 3},
		{"dense", math.Float32frombits(0xc2012345), math.Float32frombits(0x433abcde)},
		{"positive-dense", math.Float32frombits(0x3f812345), math.Float32frombits(0x410bcdef)},
		{"negative-dense", math.Float32frombits(0xc1812345), math.Float32frombits(0xc00bcdef)},
		{"positive-wide-exponents", math.SmallestNonzeroFloat32, math.MaxFloat32},
		{"negative-wide-exponents", -math.MaxFloat32, -math.SmallestNonzeroFloat32},
		{"opposite-extremes", -math.MaxFloat32, math.MaxFloat32},
		{"max-neighbors", math.Float32frombits(0x7f7ffffe), math.MaxFloat32},
		{"subnormal", math.Float32frombits(7), math.Float32frombits(0x00012345)},
		{"opposite-subnormal", -math.Float32frombits(0x00012345), math.Float32frombits(7)},
		{"normal-subnormal", math.Float32frombits(0x007fffff), math.Float32frombits(0x00800001)},
		{"negative-zero-bound", math.Float32frombits(0x80000000), 12},
	} {
		t.Run(bounds.name, func(t *testing.T) {
			s := new(Server)
			s.Rand.Other, s.Rand.Logic = prand.New(0), prand.New(93)
			table := prand.New(0)
			for index := 0; index < 4096; index++ {
				word := table.Int(0, 0x7fff)
				want := otherRandomFloatReference416090(bounds.min, bounds.max, word)
				got := s.RandomFloat416090(bounds.min, bounds.max)
				if math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("table[%d]=%d: retained result=%016x want original Chop53=%016x", index, word, math.Float64bits(got), math.Float64bits(want))
				}
				if s.Rand.Other.Index() != (index+1)%4096 || s.Rand.Logic.Index() != 93 {
					t.Fatalf("stream order at word %d: other=%d logic=%d", index, s.Rand.Other.Index(), s.Rand.Logic.Index())
				}
			}
		})
	}
}

func TestOtherRandomFloat416090AlwaysSteps(t *testing.T) {
	for _, bounds := range []struct {
		name     string
		min, max uint32
	}{
		{"positive-zero", 0, 0}, {"negative-zero", 0x80000000, 0x80000000},
		{"minus-plus-zero", 0x80000000, 0}, {"plus-minus-zero", 0, 0x80000000},
		{"equal-finite", 0x3f800000, 0x3f800000},
		{"equal-positive-infinity", 0x7f800000, 0x7f800000},
		{"equal-negative-infinity", 0xff800000, 0xff800000},
		{"quiet-nan-min", 0x7fc12345, 0x40000000},
		{"signaling-nan-min", 0x7f812345, 0x40000000},
		{"quiet-nan-max", 0x3f800000, 0xffc12345},
		{"two-nans", 0x7fc12345, 0xffc12345},
		{"positive-infinity-max", 0x3f800000, 0x7f800000},
		{"negative-infinity-max", 0x3f800000, 0xff800000},
		{"opposite-infinities", 0xff800000, 0x7f800000},
	} {
		t.Run(bounds.name, func(t *testing.T) {
			s := new(Server)
			s.Rand.Other, s.Rand.Logic = prand.New(0), prand.New(93)
			table := prand.New(0)
			min, max := math.Float32frombits(bounds.min), math.Float32frombits(bounds.max)
			for index := 0; index < 4096; index++ {
				word := table.Int(0, 0x7fff)
				want := otherRandomFloatReference416090(min, max, word)
				got := s.RandomFloat416090(min, max)
				if math.IsNaN(want) {
					if !math.IsNaN(got) {
						t.Fatalf("table[%d]=%d: exceptional result=%v want NaN", index, word, got)
					}
				} else if math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("table[%d]=%d: result=%016x want=%016x", index, word, math.Float64bits(got), math.Float64bits(want))
				}
				if s.Rand.Other.Index() != (index+1)%4096 || s.Rand.Logic.Index() != 93 {
					t.Fatalf("unconditional step at %d: other=%d logic=%d", index, s.Rand.Other.Index(), s.Rand.Logic.Index())
				}
			}
		})
	}
}
