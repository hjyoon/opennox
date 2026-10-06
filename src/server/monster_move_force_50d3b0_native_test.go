package server

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func monsterMoveForceRef53_50D3B0() *big.Float {
	return new(big.Float).SetPrec(53).SetMode(big.ToZero)
}

// Integer truncation encodes an independent binary32 FST under ToZero,
// including subnormal underflow, signed zero, and finite overflow. No
// production Nextafter/residual helper participates in this reference.
func monsterMoveForceRefSpill50D3B0(value *big.Float) float32 {
	sign := uint32(0)
	if value.Signbit() {
		sign = 0x80000000
	}
	if value.Sign() == 0 {
		return math.Float32frombits(sign)
	}
	if value.IsInf() {
		return math.Float32frombits(sign | 0x7f800000)
	}
	abs := monsterMoveForceRef53_50D3B0().Abs(value)
	exponent := abs.MantExp(nil) - 1
	if exponent > 127 {
		return math.Float32frombits(sign | 0x7f7fffff)
	}
	shift := 23 - exponent
	if exponent < -126 {
		shift = 149
	}
	scaled := monsterMoveForceRef53_50D3B0().SetMantExp(abs, shift)
	mantissa, _ := scaled.Int(nil)
	word := uint32(mantissa.Uint64())
	if exponent >= -126 {
		word = uint32(exponent+127)<<23 | (word & 0x7fffff)
	}
	return math.Float32frombits(sign | word)
}

// FSQRT's chopped 53-bit mantissa is floor(sqrt(exact rational * 2^scale)).
// Int.Sqrt avoids math.Sqrt and the production square-residual correction.
func monsterMoveForceRefSqrt50D3B0(value *big.Float) *big.Float {
	if value.Sign() == 0 {
		return monsterMoveForceRef53_50D3B0().Set(value)
	}
	exponent := value.MantExp(nil) - 1
	rootExponent := exponent / 2
	if exponent < 0 && exponent%2 != 0 {
		rootExponent--
	}
	rational, _ := value.Rat(nil)
	numerator := new(big.Int).Set(rational.Num())
	denominator := new(big.Int).Set(rational.Denom())
	shift := 104 - 2*rootExponent
	if shift >= 0 {
		numerator.Lsh(numerator, uint(shift))
	} else {
		denominator.Lsh(denominator, uint(-shift))
	}
	mantissa := new(big.Int).Sqrt(new(big.Int).Quo(numerator, denominator))
	root := monsterMoveForceRef53_50D3B0().SetInt(mantissa)
	return root.SetMantExp(root, rootExponent-52)
}

// Literal FLD/FSUB/FST, retained-Y times spilled-Y, X-square, FADD/FSQRT,
// bias/FST, segment FST, speed FMUL and force FMUL/FDIV/FST sequence from
// GAME.EXE 0050D4E6..0050D590. All inputs here are finite binary32 records.
func monsterMoveForceReference50D3B0(pos, target, from, to types.Pointf, speed, multiplier float32, running bool) (force, segment types.Pointf) {
	input := func(v float32) *big.Float { return monsterMoveForceRef53_50D3B0().SetFloat64(float64(v)) }
	sub := func(a, b float32) *big.Float { return monsterMoveForceRef53_50D3B0().Sub(input(a), input(b)) }
	dx := monsterMoveForceRefSpill50D3B0(sub(target.X, pos.X))
	dyRetained := sub(target.Y, pos.Y)
	dy := monsterMoveForceRefSpill50D3B0(dyRetained)
	ySquare := monsterMoveForceRef53_50D3B0().Mul(dyRetained, input(dy))
	xSquare := monsterMoveForceRef53_50D3B0().Mul(input(dx), input(dx))
	square := monsterMoveForceRef53_50D3B0().Add(ySquare, xSquare)
	root := monsterMoveForceRefSqrt50D3B0(square)
	// Original double at 00583C80, not the production normalization constant.
	bias := monsterMoveForceRef53_50D3B0().SetFloat64(math.Float64frombits(0x3f847ae140000000))
	distance := input(monsterMoveForceRefSpill50D3B0(monsterMoveForceRef53_50D3B0().Add(root, bias)))
	segment = types.Ptf(monsterMoveForceRefSpill50D3B0(sub(to.X, from.X)), monsterMoveForceRefSpill50D3B0(sub(to.Y, from.Y)))
	velocity := input(speed)
	if running {
		velocity = monsterMoveForceRef53_50D3B0().Mul(input(multiplier), velocity)
	}
	xProduct := monsterMoveForceRef53_50D3B0().Mul(velocity, input(dx))
	yProduct := monsterMoveForceRef53_50D3B0().Mul(velocity, input(dy))
	force = types.Ptf(
		monsterMoveForceRefSpill50D3B0(monsterMoveForceRef53_50D3B0().Quo(xProduct, distance)),
		monsterMoveForceRefSpill50D3B0(monsterMoveForceRef53_50D3B0().Quo(yProduct, distance)),
	)
	return force, segment
}

func monsterMoveForceAssertNative50D3B0(t *testing.T, unit *Object, update *MonsterUpdateData, selected int) {
	t.Helper()
	beforeUnit, beforeUpdate := *unit, *update
	from, to := update.Path[0], update.Path[1]
	if selected > 0 {
		from, to = update.Path[selected-1], update.Path[selected]
	}
	multiplier := float32(1)
	if update.MonsterDef != nil {
		multiplier = update.MonsterDef.RunMultiplier96
	}
	wantForce, segment := monsterMoveForceReference50D3B0(unit.PosVec, update.Path[selected], from, to, unit.SpeedCur, multiplier, update.StatusFlags.Has(object.MonStatusRunning))
	beforeUnit.ForceVec = wantForce
	beforeUnit.Direction1, beforeUnit.Direction2 = DirFromVec(segment), DirFromVec(segment)
	beforeUpdate.Field67 = uint32(selected)
	*memmap.PtrUint32(0x5D4594, 2386204) = 0xfedcba98
	calls := 0
	complete := monsterCreatureActuallyMove50D3B0(unit, func(origin, endpoint types.Pointf, flags MapTraceFlags) bool {
		index := calls
		calls++
		if origin != beforeUnit.PosVec || endpoint != update.Path[index] || flags != 132 || memmap.Uint32(0x5D4594, 2386204) != 0xfedcba98 {
			t.Fatal("force suffix changed trace arguments or prefix writes")
		}
		return index == selected
	})
	if complete || calls != 2 || memmap.Uint32(0x5D4594, 2386204) != uint32(selected) ||
		!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
		!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUpdate), unsafe.Sizeof(beforeUpdate)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
		t.Fatalf("force raw bits=(%08x,%08x) want=(%08x,%08x), direction=%d/%d want=%d, index=%d, complete=%t calls=%d pos=%v target=%v speed=%g multiplier=%g",
			math.Float32bits(unit.ForceVec.X), math.Float32bits(unit.ForceVec.Y), math.Float32bits(wantForce.X), math.Float32bits(wantForce.Y), unit.Direction1, unit.Direction2, beforeUnit.Direction1, update.Field67, complete, calls, beforeUnit.PosVec, beforeUpdate.Path[selected], beforeUnit.SpeedCur, multiplier)
	}
}

func TestMonsterMoveForce50D3B0NativeOriginalArithmetic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		pos, target types.Pointf
		speed, run  float32
	}{
		{"three-four-five", types.Pointf{}, types.Ptf(30, 40), 2, 1.5},
		{"negative-x", types.Pointf{}, types.Ptf(-30, 40), 2, 1.5},
		{"negative-y", types.Pointf{}, types.Ptf(30, -40), 2, 1.5},
		{"negative-both", types.Pointf{}, types.Ptf(-30, -40), 2, 1.5},
		{"negative-speed", types.Pointf{}, types.Ptf(30, 40), -2, 1.5},
		{"negative-run", types.Pointf{}, types.Ptf(30, 40), 2, -1.5},
		{"zero-speed", types.Pointf{}, types.Ptf(30, 40), 0, 1.5},
		{"zero-run", types.Pointf{}, types.Ptf(30, 40), 2, 0},
		{"live-endpoint", types.Ptf(90, 0), types.Ptf(200, 0), 2, 1.5},
		{"nonexact-root", types.Pointf{}, types.Ptf(20, 20), 1.1, 1.2},
		{"retained-y", types.Ptf(-0.002, -0.001), types.Ptf(10, 10), 3.7, 1.17},
		{"retained-y-negative", types.Ptf(0.002, 0.001), types.Ptf(-10, -10), 3.7, 1.17},
		{"large-difference", types.Ptf(-math.MaxFloat32, math.MaxFloat32), types.Ptf(math.MaxFloat32, -math.MaxFloat32), 2, 1.5},
		{"large-speed", types.Pointf{}, types.Ptf(30, 40), math.MaxFloat32, math.MaxFloat32},
		{"subnormal-force", types.Pointf{}, types.Ptf(30, 40), math.Float32frombits(3), 1.5},
		{"negative-subnormal-force", types.Pointf{}, types.Ptf(-30, 40), math.Float32frombits(3), 1.5},
	} {
		for _, running := range []bool{false, true} {
			for _, selected := range []int{0, 1} {
				t.Run(fmt.Sprintf("%s/running-%t/selected-%d", tc.name, running, selected), func(t *testing.T) {
					unit, update := monsterMoveSelectionFixture50D3B0(t)
					def, freeDef := alloc.New(MonsterDef{})
					t.Cleanup(freeDef)
					*def = MonsterDef{RunMultiplier96: tc.run}
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(def)) <= math.MaxUint32 {
						t.Fatal("native monster definition below 4 GiB")
					}
					unit.PosVec, unit.SpeedCur = tc.pos, tc.speed
					update.Field2, update.MonsterDef = 2, def
					if running {
						update.StatusFlags |= object.MonStatusRunning
					}
					update.Path[selected] = tc.target
					update.Path[1-selected] = types.Ptf(-17.125, 13.875)
					beforeDef := *def
					monsterMoveForceAssertNative50D3B0(t, unit, update, selected)
					if *def != beforeDef {
						t.Fatal("force normalization modified monster definition")
					}
				})
			}
		}
	}
}

func TestMonsterMoveForce50D3B0NativeCloseFallback(t *testing.T) {
	for _, target := range []types.Pointf{{}, {X: math.Float32frombits(1), Y: math.Float32frombits(0x80000001)}, {X: 3, Y: 4}} {
		t.Run(fmt.Sprintf("%08x-%08x", math.Float32bits(target.X), math.Float32bits(target.Y)), func(t *testing.T) {
			unit, update := monsterMoveSelectionFixture50D3B0(t)
			update.Field2 = 2
			update.Path[0], update.Path[1] = target, types.Ptf(100, 0)
			monsterMoveForceAssertNative50D3B0(t, unit, update, 0)
		})
	}
}

func TestMonsterMoveForce50D3B0NativeIndependentFiniteMatrix(t *testing.T) {
	for _, domain := range []string{"finite-binary32", "map-coordinates", "small-force", "running-product"} {
		t.Run(domain, func(t *testing.T) {
			unit, update := monsterMoveSelectionFixture50D3B0(t)
			def, freeDef := alloc.New(MonsterDef{})
			t.Cleanup(freeDef)
			*def = MonsterDef{}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(def)) <= math.MaxUint32 {
				t.Fatal("finite matrix native definition below 4 GiB")
			}
			word := uint32(0x50d4e601)
			next := func(mapCoordinate bool) float32 {
				word = 1664525*word + 1013904223
				bits := word
				if mapCoordinate {
					bits = word&0x807fffff | (uint32(128)+(word>>23)%12)<<23
				}
				if bits&0x7f800000 == 0x7f800000 {
					bits ^= 0x00800000
				}
				return math.Float32frombits(bits)
			}
			for sample := 0; sample < 1024; sample++ {
				unit.PosVec = types.Ptf(next(domain != "finite-binary32"), next(domain != "finite-binary32"))
				unit.SpeedCur = next(domain == "map-coordinates")
				if domain == "small-force" {
					unit.SpeedCur = math.Float32frombits(word & 0x80ffffff)
				}
				*def = MonsterDef{RunMultiplier96: next(domain == "map-coordinates")}
				*update = MonsterUpdateData{Field2: 2, MonsterDef: def, Field70: 333, Field71: 444, StatusFlags: object.MonStatusFrustrated}
				if domain == "running-product" || sample%2 != 0 {
					update.StatusFlags |= object.MonStatusRunning
				}
				update.Path[0] = types.Ptf(next(domain != "finite-binary32"), next(domain != "finite-binary32"))
				update.Path[1] = types.Ptf(next(domain != "finite-binary32"), next(domain != "finite-binary32"))
				// Selected zero also normalizes the close fallback, without the
				// last-point completion branch. Slot one is rejected by trace.
				monsterMoveForceAssertNative50D3B0(t, unit, update, 0)
			}
		})
	}
}

func TestMonsterMoveForce50D3B0ChopArithmetic(t *testing.T) {
	for _, operation := range []string{"multiply", "divide", "sqrt"} {
		t.Run(operation, func(t *testing.T) {
			word := uint64(0x50d57d50d581)
			next := func() float64 {
				word = 6364136223846793005*word + 1442695040888963407
				bits := word&0x800fffffffffffff | (uint64(723)+word%601)<<52
				return math.Float64frombits(bits)
			}
			for sample := 0; sample < 2048; sample++ {
				a, b := next(), next()
				x := monsterMoveForceRef53_50D3B0().SetFloat64(a)
				y := monsterMoveForceRef53_50D3B0().SetFloat64(b)
				var got float64
				var reference *big.Float
				switch operation {
				case "multiply":
					got = monsterMoveForceMulChop53_50D4FE(a, b)
					reference = monsterMoveForceRef53_50D3B0().Mul(x, y)
				case "divide":
					got = monsterMoveForceDivChop53_50D581(a, b)
					reference = monsterMoveForceRef53_50D3B0().Quo(x, y)
				case "sqrt":
					got = monsterMoveForceSqrtChop53_50D50C(math.Abs(a))
					reference = monsterMoveForceRefSqrt50D3B0(monsterMoveForceRef53_50D3B0().Abs(x))
				}
				want, accuracy := reference.Float64()
				if accuracy != big.Exact || math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("%s sample=%d a=%016x b=%016x got=%016x want=%016x accuracy=%v", operation, sample, math.Float64bits(a), math.Float64bits(b), math.Float64bits(got), math.Float64bits(want), accuracy)
				}
			}
		})
	}
}

func TestMonsterMoveForce50D3B0IndependentReferenceLiterals(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value float64
		word  uint32
	}{
		{"positive-zero", 0, 0},
		{"negative-zero", math.Copysign(0, -1), 0x80000000},
		{"positive-half-min", math.Ldexp(1, -150), 0},
		{"negative-half-min", -math.Ldexp(1, -150), 0x80000000},
		{"positive-subnormal", math.Ldexp(3, -150), 1},
		{"negative-subnormal", -math.Ldexp(3, -150), 0x80000001},
		{"positive-overflow", float64(math.MaxFloat32) * 2, 0x7f7fffff},
		{"negative-overflow", -float64(math.MaxFloat32) * 2, 0xff7fffff},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value := monsterMoveForceRef53_50D3B0().SetFloat64(tc.value)
			if got := math.Float32bits(monsterMoveForceRefSpill50D3B0(value)); got != tc.word {
				t.Fatalf("independent spill=%08x want=%08x", got, tc.word)
			}
		})
	}
	t.Run("sqrt-two", func(t *testing.T) {
		root := monsterMoveForceRefSqrt50D3B0(monsterMoveForceRef53_50D3B0().SetInt64(2))
		value, accuracy := root.Float64()
		if accuracy != big.Exact || math.Float64bits(value) != 0x3ff6a09e667f3bcc {
			t.Fatalf("independent sqrt two=%016x accuracy=%v", math.Float64bits(value), accuracy)
		}
	})
	t.Run("running-three-four-five", func(t *testing.T) {
		force, _ := monsterMoveForceReference50D3B0(types.Pointf{}, types.Ptf(30, 40), types.Pointf{}, types.Ptf(30, 40), 2, 1.5, true)
		if math.Float32bits(force.X) != 0x3fe65a9b || math.Float32bits(force.Y) != 0x401991bd {
			t.Fatalf("independent original running force=(%08x,%08x)", math.Float32bits(force.X), math.Float32bits(force.Y))
		}
	})
}
