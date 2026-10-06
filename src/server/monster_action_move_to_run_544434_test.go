package server

import (
	"math"
	"math/big"
	"testing"

	"github.com/opennox/libs/types"
)

// This reference follows the original FLD/FSUB/FMUL/FADD/FST/FCOMPP
// sequence with independent arbitrary-precision arithmetic. It does not use
// production TwoSum, FMA, Nextafter, or the former square-root distance.
func monsterMoveToRunReference544434(unit, target types.Pointf, followRange float32) (distance, near, far float64) {
	chop := func(value float64) *big.Float {
		return new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(value)
	}
	exact64 := func(value *big.Float) float64 {
		out, accuracy := value.Float64()
		if accuracy != big.Exact {
			panic("finite 53-bit MOVE_TO reference outside binary64 exponent range")
		}
		return out
	}
	product := chop(0).Mul(chop(float64(followRange)), chop(3))
	far = exact64(chop(0).Add(product, chop(30)))
	limit := chop(float64(math.MaxFloat32))
	if product.Cmp(limit) > 0 {
		near = float64(math.MaxFloat32)
	} else if product.Cmp(chop(-float64(math.MaxFloat32))) < 0 {
		near = -float64(math.MaxFloat32)
	} else {
		spill := new(big.Float).SetPrec(24).SetMode(big.ToZero).Set(product)
		out, accuracy := spill.Float32()
		if accuracy != big.Exact {
			panic("Field329*3 binary32 spill reference not exact")
		}
		near = float64(out)
	}
	dx := chop(0).Sub(chop(float64(target.X)), chop(float64(unit.X)))
	dy := chop(0).Sub(chop(float64(target.Y)), chop(float64(unit.Y)))
	ySquare := chop(0).Mul(dy, dy)
	xSquare := chop(0).Mul(dx, dx)
	distance = exact64(chop(0).Add(ySquare, xSquare))
	near = exact64(chop(0).Mul(chop(near), chop(near)))
	far = exact64(chop(0).Mul(chop(far), chop(far)))
	return
}

func TestMonsterMoveToRun544434IndependentChopModel(t *testing.T) {
	for _, domain := range []string{"finite-binary32", "map-coordinates", "subnormal-coordinates", "adjacent-run-bands"} {
		t.Run(domain, func(t *testing.T) {
			word := uint32(0x54443401)
			next := func() float32 {
				word = 1664525*word + 1013904223
				bits := word
				switch domain {
				case "map-coordinates":
					bits = word&0x807fffff | (uint32(130)+(word>>23)%10)<<23
				case "subnormal-coordinates":
					bits = word & 0x807fffff
				default:
					if bits&0x7f800000 == 0x7f800000 {
						bits ^= 0x00800000
					}
				}
				return math.Float32frombits(bits)
			}
			for i := 0; i < 8192; i++ {
				unit, target := types.Pointf{X: next(), Y: next()}, types.Pointf{X: next(), Y: next()}
				followRange := next()
				if domain == "adjacent-run-bands" {
					unit = types.Pointf{}
					followRange = math.Float32frombits(word&0x007fffff | (130+(word>>23)%5)<<23)
					x := float32(float64(followRange) * 3)
					if i&1 != 0 {
						x = float32(float64(followRange)*3 + 30)
					}
					target = types.Pointf{X: math.Float32frombits(math.Float32bits(x) ^ 1), Y: math.Float32frombits(word&0x007fffff | 0x39000000)}
				}
				wantDistance, wantNear, wantFar := monsterMoveToRunReference544434(unit, target, followRange)
				product := float64(followRange) * 3
				gotNear := monsterMoveToRunSquareChop53_544434(monsterMoveToRunSpill544440(product))
				gotFar := monsterMoveToRunSquareChop53_544434(monsterMoveToRunAddChop53_544434(product, 30))
				dx := monsterMoveToRunAddChop53_544434(float64(target.X), -float64(unit.X))
				dy := monsterMoveToRunAddChop53_544434(float64(target.Y), -float64(unit.Y))
				gotDistance := monsterMoveToRunAddChop53_544434(monsterMoveToRunSquareChop53_544434(dy), monsterMoveToRunSquareChop53_544434(dx))
				if math.Float64bits(gotDistance) != math.Float64bits(wantDistance) || math.Float64bits(gotNear) != math.Float64bits(wantNear) || math.Float64bits(gotFar) != math.Float64bits(wantFar) {
					t.Fatalf("retained arithmetic at %d differs: distance=%016x/%016x near=%016x/%016x far=%016x/%016x inputs=%v,%v,%g", i, math.Float64bits(gotDistance), math.Float64bits(wantDistance), math.Float64bits(gotNear), math.Float64bits(wantNear), math.Float64bits(gotFar), math.Float64bits(wantFar), unit, target, followRange)
				}
				want := 0
				if wantDistance < wantNear {
					want = -1
				} else if wantDistance > wantFar {
					want = 1
				}
				if got := monsterMoveToRunBand544434(unit, target, followRange); got != want {
					t.Fatalf("original run band at %d differs: got=%d want=%d distance=%016x near=%016x far=%016x", i, got, want, math.Float64bits(wantDistance), math.Float64bits(wantNear), math.Float64bits(wantFar))
				}
			}
		})
	}
}

func TestMonsterMoveToRunSpill544440OriginalChopBits(t *testing.T) {
	for _, tc := range []struct {
		name            string
		rangeWord, want uint32
	}{
		{"positive-zero", 0, 0}, {"negative-zero", 0x80000000, 0x80000000},
		{"positive-smallest-subnormal", 1, 3}, {"negative-smallest-subnormal", 0x80000001, 0x80000003},
		{"positive-rounded-product", 0x41200001, 0x41f00001}, {"negative-rounded-product", 0xc1200001, 0xc1f00001},
		{"positive-overflow", 0x7f7fffff, 0x7f7fffff}, {"negative-overflow", 0xff7fffff, 0xff7fffff},
		{"positive-infinity", 0x7f800000, 0x7f800000}, {"negative-infinity", 0xff800000, 0xff800000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := math.Float32bits(float32(monsterMoveToRunSpill544440(float64(math.Float32frombits(tc.rangeWord)) * 3)))
			if got != tc.want {
				t.Fatalf("original FST/chop word=%08x want=%08x", got, tc.want)
			}
		})
	}
}
