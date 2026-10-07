package server

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/opennox/libs/prand"
)

// Independent instruction-order model of GAME.EXE 00416030 under the
// gameplay control word (004031F2 precision 53, 0043E2C1 round toward zero).
// It uses neither production rounding helpers nor a decimal scale literal.
func logicRandomFloatReference53_416030(a, b float64, operation byte) float64 {
	x := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(a)
	y := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(b)
	switch operation {
	case '-':
		x.Sub(x, y)
	case '*':
		x.Mul(x, y)
	default:
		x.Add(x, y)
	}
	value, _ := x.Float64()
	return value
}

func logicRandomFloatReferenceValue416030(min, max float32, word int) float64 {
	delta := logicRandomFloatReference53_416030(float64(max), float64(min), '-')
	scaled := logicRandomFloatReference53_416030(float64(word), float64(math.Float32frombits(0x38000100)), '*')
	product := logicRandomFloatReference53_416030(delta, scaled, '*')
	return logicRandomFloatReference53_416030(product, float64(min), '+')
}

func TestLogicRandomFloat416030OriginalGameplayPrecision(t *testing.T) {
	for _, bounds := range []struct {
		name     string
		min, max float32
	}{
		{"pi", math.Float32frombits(0xc0490fdb), math.Float32frombits(0x40490fdb)},
		{"unit", 0, 1},
		{"negative-unit", -1, 0},
		{"reverse-unit", 1, -1},
		{"motion", -2, 3},
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
			random, table := prand.New(0), prand.New(0)
			for index := 0; index < 4096; index++ {
				word := table.Int(0, 0x7fff)
				want := logicRandomFloatReferenceValue416030(bounds.min, bounds.max, word)
				got := logicRandomFloat416030(random, bounds.min, bounds.max)
				if math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("table[%d]=%d: retained result=%016x want original Chop53=%016x", index, word, math.Float64bits(got), math.Float64bits(want))
				}
				if random.Index() != (index+1)%4096 {
					t.Fatalf("index after word %d = %d", index, random.Index())
				}
			}
		})
	}
}

func TestLogicRandomFloat416030OriginalGameplayNoStep(t *testing.T) {
	for _, bounds := range [][2]uint32{
		{0, 0}, {0x80000000, 0x80000000}, {0x80000000, 0}, {0, 0x80000000},
		{0x3f800000, 0x3f800000}, {0x7f800000, 0x7f800000}, {0xff800000, 0xff800000},
		{0x7fc12345, 0x40000000}, {0x7f812345, 0x40000000},
		{0x3f800000, 0xffc12345}, {0x7fc12345, 0xffc12345},
	} {
		for _, index := range []int{17, 4095} {
			t.Run(fmt.Sprintf("%08x-%08x/at-%d", bounds[0], bounds[1], index), func(t *testing.T) {
				random := prand.New(index)
				min, max := math.Float32frombits(bounds[0]), math.Float32frombits(bounds[1])
				got := logicRandomFloat416030(random, min, max)
				if math.IsNaN(float64(max)) {
					if !math.IsNaN(got) {
						t.Fatalf("unordered max produced %v", got)
					}
				} else if math.Float64bits(got) != math.Float64bits(float64(max)) {
					t.Fatalf("no-step result=%016x want max=%016x", math.Float64bits(got), math.Float64bits(float64(max)))
				}
				if random.Index() != index {
					t.Fatalf("no-step path advanced %d to %d", index, random.Index())
				}
			})
		}
	}
}
