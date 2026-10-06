package server

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Independent finite-input register model for GAME.EXE 004E6C08..004E6CD3:
// each FSUB/FMUL/FADD is precision 53/ToZero. Integer sqrt truncates the
// retained significand without math.Sqrt or any production residual helper.
func objectDistanceReferenceRetained4E6C00(a, b *Object) float64 {
	chop := func() *big.Float { return new(big.Float).SetPrec(53).SetMode(big.ToZero) }
	input := func(v float32) *big.Float { return chop().SetFloat64(float64(v)) }
	dx := chop().Sub(input(a.PosVec.X), input(b.PosVec.X))
	dy := chop().Sub(input(a.PosVec.Y), input(b.PosVec.Y))
	square := chop().Add(chop().Mul(dy, dy), chop().Mul(dx, dx))
	distance := chop().Set(square)
	if square.Sign() != 0 {
		exponent := square.MantExp(nil) - 1
		rootExponent := exponent / 2
		if exponent < 0 && exponent%2 != 0 {
			rootExponent--
		}
		rational, _ := square.Rat(nil)
		numerator, denominator := new(big.Int).Set(rational.Num()), new(big.Int).Set(rational.Denom())
		shift := 104 - 2*rootExponent
		if shift >= 0 {
			numerator.Lsh(numerator, uint(shift))
		} else {
			denominator.Lsh(denominator, uint(-shift))
		}
		mantissa := new(big.Int).Sqrt(new(big.Int).Quo(numerator, denominator))
		distance.SetInt(mantissa)
		distance.SetMantExp(distance, rootExponent-52)
	}
	for _, shape := range []*Shape{&a.Shape, &b.Shape} {
		var extent float64
		switch shape.Kind {
		case ShapeKindCircle:
			extent = float64(shape.Circle.R)
		case ShapeKindBox:
			width := float64(shape.Box.W) * 0.5
			height := objectDistanceHeightReference4E6C54(math.Float32bits(shape.Box.H))
			extent = height
			if width > height {
				extent = width
			}
		default:
			continue
		}
		distance.Sub(distance, chop().SetFloat64(extent))
	}
	minimum := input(math.Float32frombits(0x3c23d70a))
	if distance.Cmp(minimum) < 0 {
		distance.Set(minimum)
	}
	result, accuracy := distance.Float64()
	if accuracy != big.Exact {
		panic("independent retained distance is not representable in binary64")
	}
	return result
}

func objectDistanceNativePair4E6C00(t *testing.T) (*Object, *Object) {
	t.Helper()
	a, freeA := alloc.New(Object{})
	b, freeB := alloc.New(Object{})
	t.Cleanup(freeA)
	t.Cleanup(freeB)
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(a)) <= math.MaxUint32 || uintptr(unsafe.Pointer(b)) <= math.MaxUint32) {
		t.Fatalf("native object addresses must exceed 4 GiB: %p %p", a, b)
	}
	return a, b
}

func objectDistanceAssertNative4E6C00(t *testing.T, a, b *Object, want float64) {
	t.Helper()
	aBefore := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(a)), int(unsafe.Sizeof(*a))))
	bBefore := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(b)), int(unsafe.Sizeof(*b))))
	if got := ObjectDistance4E6C00(a, b); math.Float64bits(got) != math.Float64bits(want) {
		t.Errorf("original retained distance=%016x want=%016x a=%v b=%v", math.Float64bits(got), math.Float64bits(want), a.PosVec, b.PosVec)
	}
	if !bytes.Equal(aBefore, unsafe.Slice((*byte)(unsafe.Pointer(a)), len(aBefore))) || !bytes.Equal(bBefore, unsafe.Slice((*byte)(unsafe.Pointer(b)), len(bBefore))) {
		t.Fatal("distance calculation changed a native object record")
	}
}

func TestObjectDistanceNative4E6C00RetainedPrecision(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b types.Pointf
	}{
		{"sqrt-two", types.Ptf(1, 1), types.Pointf{}},
		{"sqrt-five", types.Ptf(1, 2), types.Pointf{}},
		{"sqrt-thirteen", types.Ptf(2, 3), types.Pointf{}},
		{"retained-x-subtract", types.Ptf(4, 3), types.Ptf(float32(math.Ldexp(1, -52)), 0)},
		{"retained-y-subtract", types.Ptf(3, 4), types.Ptf(0, float32(math.Ldexp(1, -52)))},
		{"wide-positive-subtract", types.Ptf(math.MaxFloat32, 2), types.Ptf(-1, -3)},
		{"wide-negative-subtract", types.Ptf(-math.MaxFloat32, -2), types.Ptf(1, 3)},
		{"exact-three-four-five", types.Ptf(3, 4), types.Pointf{}},
		{"below-floor", types.Ptf(math.SmallestNonzeroFloat32, 0), types.Pointf{}},
	} {
		for _, kind := range []ShapeKind{ShapeKindCenter, ShapeKindCircle, ShapeKindBox, ShapeKind(0x60000002)} {
			t.Run(fmt.Sprintf("%s/kind-%08x", tc.name, uint32(kind)), func(t *testing.T) {
				a, b := objectDistanceNativePair4E6C00(t)
				a.PosVec, b.PosVec = tc.a, tc.b
				a.Shape, b.Shape = Shape{Kind: kind}, Shape{Kind: kind}
				a.Shape.Circle.R, b.Shape.Circle.R = 0.125, 0.25
				a.Shape.Box.W, a.Shape.Box.H = 0.25, 0.0625
				b.Shape.Box.W, b.Shape.Box.H = 0.125, 0.5
				objectDistanceAssertNative4E6C00(t, a, b, objectDistanceReferenceRetained4E6C00(a, b))
			})
		}
	}
	t.Run("reference-calibration", func(t *testing.T) {
		a, b := objectDistanceNativePair4E6C00(t)
		a.PosVec = types.Ptf(1, 1)
		if got := objectDistanceReferenceRetained4E6C00(a, b); math.Float64bits(got) != 0x3ff6a09e667f3bcc {
			t.Fatalf("independent chopped sqrt(2)=%016x", math.Float64bits(got))
		}
		a.PosVec, b.PosVec = types.Ptf(4, 3), types.Ptf(float32(math.Ldexp(1, -52)), 0)
		if !(objectDistanceReferenceRetained4E6C00(a, b) < 5) {
			t.Fatal("independent subtraction boundary lost retained register precision")
		}
	})
	t.Run("independent-finite-matrix", func(t *testing.T) {
		a, b := objectDistanceNativePair4E6C00(t)
		random := rand.New(rand.NewSource(0x4e6c00)) // test-local, never the game's RNG
		for sample := 0; sample < 4096; sample++ {
			next := func() float32 {
				word := random.Uint32()
				if word&0x7f800000 == 0x7f800000 {
					word ^= 0x00800000
				}
				return math.Float32frombits(word)
			}
			a.PosVec, b.PosVec = types.Ptf(next(), next()), types.Ptf(next(), next())
			for _, obj := range []*Object{a, b} {
				obj.Shape = Shape{Kind: ShapeKind(sample % 4)}
				obj.Shape.Circle.R = next()
				obj.Shape.Box.W, obj.Shape.Box.H = next(), next()
			}
			objectDistanceAssertNative4E6C00(t, a, b, objectDistanceReferenceRetained4E6C00(a, b))
			if t.Failed() {
				t.Fatalf("independent finite sample=%d", sample)
			}
		}
	})
}

func TestMonsterMainThreatFlee547210NativeDistanceChop(t *testing.T) {
	for _, kind := range []ShapeKind{ShapeKindCenter, ShapeKindCircle, ShapeKindBox} {
		t.Run(fmt.Sprintf("kind-%d", kind), func(t *testing.T) {
			s, unit, update, _ := newMonsterMainFleeTest547210(t)
			_, enemy := objectDistanceNativePair4E6C00(t)
			update.CurrentEnemy, update.PreferredEnemy = enemy, enemy
			unit.PosVec, enemy.PosVec = types.Ptf(4, 3), types.Ptf(float32(math.Ldexp(1, -52)), 0)
			unit.Shape, enemy.Shape = Shape{Kind: kind}, Shape{Kind: kind}
			unit.Shape.Circle.R, enemy.Shape.Circle.R = 1, 1
			unit.Shape.Box.W, unit.Shape.Box.H = 2, 1
			enemy.Shape.Box.W, enemy.Shape.Box.H = 1, 2
			update.FleeRange = 5
			if kind != ShapeKindCenter {
				update.FleeRange = 3
			}
			if !(objectDistanceReferenceRetained4E6C00(unit, enemy) < float64(update.FleeRange)) {
				t.Fatal("independent original boundary must schedule FLEE")
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unit.UpdateData) <= math.MaxUint32) {
				t.Fatalf("native FLEE unit/update below 4 GiB: %p %p", unit, unit.UpdateData)
			}
			unitBefore := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))))
			enemyBefore := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(enemy)), int(unsafe.Sizeof(*enemy))))
			rolls := 0
			if !s.monsterMainThreatFlee547210(unit, update, nil, MonsterMainRuntime547210{
				RandomInt: func(min, max int) int {
					rolls++
					if min != 0 || max != 1 {
						t.Fatal("ordinary FLEE RNG bounds changed")
					}
					return 0
				},
			}) || update.AIStackInd != 4 || update.AIStackHead().Type() != ai.ACTION_FLEE || rolls != 1 {
				t.Fatal("default native distance service rejected original FLEE boundary")
			}
			if update.AIStack[3].ArgU32(0) != math.Float32bits(update.FleeRange+30) || update.AIStack[4].ArgU32(0) != math.Float32bits(enemy.PosVec.X) || update.AIStack[4].ArgU32(1) != math.Float32bits(enemy.PosVec.Y) {
				t.Fatal("FLEE action arguments changed")
			}
			if !bytes.Equal(unitBefore, unsafe.Slice((*byte)(unsafe.Pointer(unit)), len(unitBefore))) || !bytes.Equal(enemyBefore, unsafe.Slice((*byte)(unsafe.Pointer(enemy)), len(enemyBefore))) {
				t.Fatal("FLEE scheduling changed native object records")
			}
		})
	}
}
