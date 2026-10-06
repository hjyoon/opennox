package opennox

import (
	"math"
	"math/big"
	"testing"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestAILocationDependency546B63RetainedChopBits(t *testing.T) {
	for _, tc := range []struct {
		name               string
		unitX, unitY, x, y uint32
		want               uint64
	}{
		{"above-one", 0, 0, 0x3f800000, 0x39000000, 0x3ff0000001ffffff},
		{"below-one", 0, 0, 0x3f7fffff, 0x39a7c5ac, 0x3feffffffb7cdfd2},
		{"subnormal-diagonal", 0, 0, 1, 1, 0x36a6a09e667f3bcc},
		{"exact-three-four-five", 0, 0, 0x40400000, 0x40800000, 0x4014000000000000},
		{"positive-subtraction-chop", 1, 0, 0x3f800000, 0, 0x3feffffffffffffe},
		{"negative-subtraction-chop", 0x80000001, 0, 0xbf800000, 0, 0x3feffffffffffffe},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dx := aiDependencyAddChop53_546B63(float64(math.Float32frombits(tc.x)), -float64(math.Float32frombits(tc.unitX)))
			dy := aiDependencyAddChop53_546B63(float64(math.Float32frombits(tc.y)), -float64(math.Float32frombits(tc.unitY)))
			ySquare := aiDependencySquareChop53_546B63(dy)
			xSquare := aiDependencySquareChop53_546B63(dx)
			got := aiDependencySqrtChop53_546B63(aiDependencyAddChop53_546B63(ySquare, xSquare))
			if bits := math.Float64bits(got); bits != tc.want {
				t.Fatalf("retained x87 chop distance bits=%016x want=%016x", bits, tc.want)
			}
		})
	}
}

func aiLocationReferenceChop53_546B63(value float64) *big.Float {
	return new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(value)
}

func aiLocationReferenceSqrtChop53_546B63(square *big.Float) *big.Float {
	// Search the ordered positive binary64 words using exact rational
	// squares. This proves directed rounding even for a perfect square;
	// an approximate inverse-square-root reference can undershoot it.
	value, _ := square.Rat(nil)
	if value == nil || value.Sign() < 0 {
		panic("reference square must be finite and nonnegative")
	}
	low, high := uint64(0), math.Float64bits(math.MaxFloat64)
	candidate, candidateSquare := new(big.Rat), new(big.Rat)
	for low < high {
		mid := low + (high-low+1)/2
		candidate.SetFloat64(math.Float64frombits(mid))
		candidateSquare.Mul(candidate, candidate)
		if candidateSquare.Cmp(value) <= 0 {
			low = mid
		} else {
			high = mid - 1
		}
	}
	return aiLocationReferenceChop53_546B63(math.Float64frombits(low))
}

func TestAILocationDependency546B63ExactReferenceSqrtBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name   string
		square float64
		want   uint64
	}{
		{"zero", 0, 0},
		{"perfect-square", 25, 0x4014000000000000},
		{"irrational-root", 2, 0x3ff6a09e667f3bcc},
		{"below-one", math.Float64frombits(0x3feffffffffffffe), 0x3feffffffffffffe},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value := aiLocationReferenceSqrtChop53_546B63(aiLocationReferenceChop53_546B63(tc.square))
			got, accuracy := value.Float64()
			if accuracy != big.Exact || math.Float64bits(got) != tc.want {
				t.Fatalf("exact reference root=%016x accuracy=%v want=%016x", math.Float64bits(got), accuracy, tc.want)
			}
		})
	}
}

func TestAILocationDependency546B63IndependentChopModel(t *testing.T) {
	// The reference uses independent arbitrary-precision, directed rounding
	// and exact rational square-root bounds; it does not call production
	// TwoSum/FMA/Nextafter operations. Values
	// originate as finite binary32 words, the actual location input domain.
	for _, operation := range []string{"subtraction", "square", "sum", "sqrt"} {
		t.Run(operation, func(t *testing.T) {
			word := uint32(0x546b6301)
			next := func() float64 {
				word = 1664525*word + 1013904223
				bits := word
				if bits&0x7f800000 == 0x7f800000 {
					bits ^= 0x00800000
				}
				return float64(math.Float32frombits(bits))
			}
			for i := 0; i < 512; i++ {
				x, unitX, y, unitY := next(), next(), next(), next()
				dx := aiLocationReferenceChop53_546B63(0).Sub(aiLocationReferenceChop53_546B63(x), aiLocationReferenceChop53_546B63(unitX))
				dy := aiLocationReferenceChop53_546B63(0).Sub(aiLocationReferenceChop53_546B63(y), aiLocationReferenceChop53_546B63(unitY))
				ySquare := aiLocationReferenceChop53_546B63(0).Mul(dy, dy)
				xSquare := aiLocationReferenceChop53_546B63(0).Mul(dx, dx)
				sum := aiLocationReferenceChop53_546B63(0).Add(ySquare, xSquare)
				root := aiLocationReferenceSqrtChop53_546B63(sum)
				gotX := aiDependencyAddChop53_546B63(x, -unitX)
				gotY := aiDependencyAddChop53_546B63(y, -unitY)
				gotYSquare := aiDependencySquareChop53_546B63(gotY)
				gotXSquare := aiDependencySquareChop53_546B63(gotX)
				gotSum := aiDependencyAddChop53_546B63(gotYSquare, gotXSquare)
				var got float64
				var reference *big.Float
				switch operation {
				case "subtraction":
					got, reference = gotX, dx
				case "square":
					got, reference = gotXSquare, xSquare
				case "sum":
					got, reference = gotSum, sum
				case "sqrt":
					got, reference = aiDependencySqrtChop53_546B63(gotSum), root
				}
				want, accuracy := reference.Float64()
				if accuracy != big.Exact {
					t.Fatal("53-bit reference unexpectedly outside binary64 exponent range")
				}
				if math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("%s at finite coordinate set %d differs: got=%016x want=%016x inputs=%x,%x,%x,%x", operation, i, math.Float64bits(got), math.Float64bits(want), math.Float64bits(x), math.Float64bits(unitX), math.Float64bits(y), math.Float64bits(unitY))
				}
			}
		})
	}
}

func TestAILocationDependency546B63MissingInputFaultPrefix(t *testing.T) {
	for _, closer := range []bool{false, true} {
		label := "far"
		if closer {
			label = "close"
		}
		for _, missing := range []string{"unit", "slot", "both"} {
			t.Run(label+"/nil-"+missing, func(t *testing.T) {
				unit := &server.Object{PosVec: types.Pointf{X: 1, Y: 2}}
				slot := &server.AIStackItem{Action: 51, Args: [4]uintptr{0x7fc12345, 0x12345678, 0x80000000, 0xff800001}, Field5: 17}
				before := *slot
				selectedSlot := slot
				if missing == "unit" || missing == "both" {
					unit = nil
				}
				if missing == "slot" || missing == "both" {
					selectedSlot = nil
				}
				var fault any
				func() {
					defer func() { fault = recover() }()
					aiDependencyLocation546B63(unit, selectedSlot, closer)
				}()
				if fault == nil || *slot != before {
					t.Fatalf("missing input fault=%v or mutated entry-cached action slot", fault)
				}
			})
		}
	}
}
