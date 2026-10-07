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
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Independently interpret the arithmetic of sealed GAME.EXE 0057C8A0.
// No production helper, FMA/Nextafter correction, thread FPU change or
// recorded result is an input. Each x87 instruction rounds to 53 bits,
// ToZero; the FST/FSTP boundaries are separate 24-bit stores.
func pointOnLineReferenceOp57C8A0(a, b float64, operation byte) float64 {
	if math.IsNaN(a) || math.IsNaN(b) || math.IsInf(a, 0) || math.IsInf(b, 0) || operation == '/' && b == 0 {
		switch operation {
		case '*':
			return a * b
		case '/':
			return a / b
		default:
			return a + b
		}
	}
	input := func(v float64) *big.Float { return new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(v) }
	result := new(big.Float).SetPrec(53).SetMode(big.ToZero)
	switch operation {
	case '*':
		result.Mul(input(a), input(b))
	case '/':
		result.Quo(input(a), input(b))
	default:
		result.Add(input(a), input(b))
	}
	v, accuracy := result.Float64()
	if accuracy != big.Exact {
		panic("bounded binary32-coordinate x87 intermediates must fit binary64")
	}
	return v
}

func pointOnLineReferenceSpill57C8A0(v float64) float32 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v == 0 {
		return float32(v)
	}
	negative := math.Signbit(v)
	value := new(big.Float).SetPrec(256).SetFloat64(math.Abs(v))
	var out float32
	if value.Cmp(new(big.Float).SetFloat64(math.MaxFloat32)) >= 0 {
		out = math.MaxFloat32 // finite overflow under chop saturates
	} else if value.MantExp(nil) <= -126 {
		// A subnormal store truncates on the fixed 2^-149 grid, not
		// on the unbounded-exponent grid of big.Float's precision 24.
		units, _ := new(big.Float).SetPrec(256).SetMantExp(value, 149).Int(nil)
		out = float32(units.Uint64()) * math.SmallestNonzeroFloat32
	} else {
		rounded := new(big.Float).SetPrec(24).SetMode(big.ToZero).Set(value)
		var accuracy big.Accuracy
		out, accuracy = rounded.Float32()
		if accuracy != big.Exact {
			panic("normal chopped 24-bit store must be exactly binary32")
		}
	}
	if negative {
		out = -out
	}
	return out
}

func pointOnLineReference57C8A0(from, to, center types.Pointf) (out types.Pointf, ok bool) {
	add := func(a, b float64) float64 { return pointOnLineReferenceOp57C8A0(a, b, '+') }
	mul := func(a, b float64) float64 { return pointOnLineReferenceOp57C8A0(a, b, '*') }
	div := func(a, b float64) float64 { return pointOnLineReferenceOp57C8A0(a, b, '/') }
	spill := func(v float64) float64 { return float64(pointOnLineReferenceSpill57C8A0(v)) }
	x := add(float64(to.X), -float64(from.X))
	y := add(float64(to.Y), -float64(from.Y))
	storedY := spill(y)                              // FST 0057C8BA; the register is not popped
	length := spill(add(mul(y, storedY), mul(x, x))) // FSTP 0057C8C8
	cx := add(float64(center.X), -float64(from.X))
	cy := add(float64(center.Y), -float64(from.Y))
	dot := add(mul(cy, storedY), mul(cx, x))
	storedDot := spill(dot) // FST 0057C8E0; retain dot for X
	xQuotient := div(mul(dot, x), length)
	yQuotient := spill(div(mul(storedDot, storedY), length)) // FSTP 0057C8F6
	out.X = pointOnLineReferenceSpill57C8A0(add(xQuotient, float64(from.X)))
	out.Y = pointOnLineReferenceSpill57C8A0(add(yQuotient, float64(from.Y)))
	low, high := from, to
	if !(from.X < to.X) && !math.IsNaN(float64(from.X)) && !math.IsNaN(float64(to.X)) {
		low.X, high.X = to.X, from.X
	}
	if !(from.Y < to.Y) && !math.IsNaN(float64(from.Y)) && !math.IsNaN(float64(to.Y)) {
		low.Y, high.Y = to.Y, from.Y
	}
	// FCOMP C0|C3 admits unordered in the X and final Y comparisons;
	// the preceding Y comparison tests C0 alone and rejects unordered.
	return out, !(low.X > out.X) && !(out.X > high.X) && out.Y >= low.Y && !(out.Y > high.Y)
}

func pointOnLineSame57C8A0(got, want float32) bool {
	return math.Float32bits(got) == math.Float32bits(want) || math.IsNaN(float64(got)) && math.IsNaN(float64(want))
}

func pointOnLineAssert57C8A0(t *testing.T, from, to, center types.Pointf) {
	t.Helper()
	want, admitted := pointOnLineReference57C8A0(from, to, center)
	got, ok := PointOnTheLine(from, to, center)
	if ok != admitted || !pointOnLineSame57C8A0(got.X, want.X) || !pointOnLineSame57C8A0(got.Y, want.Y) {
		t.Errorf("original projection differs: from=%v to=%v center=%v got=%08x/%08x/%t want=%08x/%08x/%t",
			from, to, center, math.Float32bits(got.X), math.Float32bits(got.Y), ok, math.Float32bits(want.X), math.Float32bits(want.Y), admitted)
	}
}

func TestPointOnTheLine57C8A0ReferenceCalibration(t *testing.T) {
	for _, tc := range []struct {
		name             string
		from, to, center types.Pointf
		x, y             uint32
		ok               bool
	}{
		{"horizontal-interior", types.Ptf(0, 0), types.Ptf(8, 0), types.Ptf(4, 2), 0x40800000, 0, true},
		{"vertical-interior", types.Ptf(0, 0), types.Ptf(0, 8), types.Ptf(2, 4), 0, 0x40800000, true},
		{"exact-start", types.Ptf(0, 0), types.Ptf(3, 4), types.Ptf(0, 0), 0, 0, true},
		{"exact-end", types.Ptf(0, 0), types.Ptf(3, 4), types.Ptf(3, 4), 0x40400000, 0x40800000, true},
		{"exact-midpoint", types.Ptf(0, 0), types.Ptf(6, 8), types.Ptf(3, 4), 0x40400000, 0x40800000, true},
		{"3-4-chopped-quotient", types.Ptf(0, 0), types.Ptf(3, 4), types.Ptf(1, 1), 0x3f570a3d, 0x3f8f5c28, true},
		{"before-start", types.Ptf(0, 0), types.Ptf(3, 4), types.Ptf(-3, -4), 0xc0400000, 0xc0800000, false},
		{"after-end", types.Ptf(0, 0), types.Ptf(3, 4), types.Ptf(6, 8), 0x40c00000, 0x41000000, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := pointOnLineReference57C8A0(tc.from, tc.to, tc.center)
			if math.Float32bits(got.X) != tc.x || math.Float32bits(got.Y) != tc.y || ok != tc.ok {
				t.Fatalf("independent reference calibration=%08x/%08x/%t want=%08x/%08x/%t", math.Float32bits(got.X), math.Float32bits(got.Y), ok, tc.x, tc.y, tc.ok)
			}
		})
	}
}

func TestPointOnTheLine57C8A0OriginalSpills(t *testing.T) {
	for _, ray := range []struct {
		name     string
		from, to types.Pointf
	}{
		{"3-4", types.Ptf(0, 0), types.Ptf(3, 4)},
		{"ascending-world", types.Ptf(3300, 2212), types.Ptf(3428, 2308)},
		{"descending-world", types.Ptf(300, 412), types.Ptf(428, 300)},
		{"captured-food-arrival", types.Ptf(3426.8457, 2309.9163), types.Ptf(3428, 2308)},
		{"retained-delta", types.Ptf(math.SmallestNonzeroFloat32, -math.SmallestNonzeroFloat32), types.Ptf(3, 4)},
		{"fractional-deltas", types.Ptf(0.125, -0.375), types.Ptf(3.0625, 4.03125)},
		{"length-underflow", types.Ptf(0, 0), types.Ptf(0x1p-80, -0x1p-90)},
		{"finite-length-overflow", types.Ptf(-math.MaxFloat32, math.MaxFloat32), types.Ptf(math.MaxFloat32, -math.MaxFloat32)},
	} {
		for _, reverse := range []bool{false, true} {
			for _, fraction := range []float32{-0.125, 0, 0.125, 0.5, 0.875, 1, 1.125} {
				for _, offset := range []float32{0, 0x1p-12, -0x1p-12, math.SmallestNonzeroFloat32, -math.SmallestNonzeroFloat32} {
					t.Run(fmt.Sprintf("%s/reverse-%t/fraction-%g/offset-%g", ray.name, reverse, fraction, offset), func(t *testing.T) {
						from, to := ray.from, ray.to
						if reverse {
							from, to = to, from
						}
						// This constructs inputs, not an expected projection.
						center := types.Ptf(float32(float64(from.X)+(float64(to.X)-float64(from.X))*float64(fraction)),
							float32(float64(from.Y)+(float64(to.Y)-float64(from.Y))*float64(fraction))+offset)
						pointOnLineAssert57C8A0(t, from, to, center)
					})
				}
			}
		}
	}
}

func TestPointOnTheLine57C8A0SpecialValues(t *testing.T) {
	for _, value := range []float32{0, math.Float32frombits(0x80000000), math.SmallestNonzeroFloat32,
		-math.SmallestNonzeroFloat32, math.MaxFloat32, -math.MaxFloat32, float32(math.Inf(1)), float32(math.Inf(-1)),
		math.Float32frombits(0x7fc12345), math.Float32frombits(0xffc12345)} {
		for input := 0; input < 6; input++ {
			t.Run(fmt.Sprintf("%08x/input-%d", math.Float32bits(value), input), func(t *testing.T) {
				from, to, center := types.Ptf(0, 0), types.Ptf(3, 4), types.Ptf(1, 1)
				*[]*float32{&from.X, &from.Y, &to.X, &to.Y, &center.X, &center.Y}[input] = value
				pointOnLineAssert57C8A0(t, from, to, center)
			})
		}
	}
}

// Real indexed vision/CanInteract/Refresh calls on C-owned native records.
// This is not a stock-world GUI run or proof of passive food consumption.
func TestPointOnTheLine57C8A0NativeVisionRefresh(t *testing.T) {
	for _, ray := range []struct {
		name     string
		from, to types.Pointf
	}{
		{"ascending", types.Ptf(3300, 2212), types.Ptf(3428, 2308)},
		{"descending", types.Ptf(300, 412), types.Ptf(428, 300)},
		{"captured-arrival", types.Ptf(3426.8457, 2309.9163), types.Ptf(3428, 2308)},
	} {
		for _, reverse := range []bool{false, true} {
			for _, fraction := range []float32{0.125, 0.5, 0.875} {
				for _, radius := range []float32{0, 0x1p-12, 1} {
					for _, shadow := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/reverse-%t/fraction-%g/radius-%g/shadow-%t", ray.name, reverse, fraction, radius, shadow), func(t *testing.T) {
							s, unit, update := moveToNativeRetryFixture5443F0(t, ai.ACTION_MOVE_TO)
							target, freeTarget := alloc.New(Object{})
							t.Cleanup(freeTarget)
							obstacle, freeObstacle := alloc.New(Object{})
							t.Cleanup(freeObstacle)
							s.Map.Init()
							s.Walls.byPos = make([]*Wall, wallsPerBucket*WallGridSize)
							t.Cleanup(s.Map.Free)
							from, to := ray.from, ray.to
							if reverse {
								from, to = to, from
							}
							center := types.Ptf(from.X+(to.X-from.X)*fraction, from.Y+(to.Y-from.Y)*fraction)
							unit.PosVec, unit.NewPos, unit.ObjFlags = from, from, object.FlagActive
							*target = Object{ObjClass: object.ClassFood, ObjFlags: object.FlagActive | object.FlagNoCollide, PosVec: to, NewPos: to}
							*obstacle = Object{ObjClass: object.ClassObstacle, ObjFlags: object.FlagActive, PosVec: center, NewPos: center}
							if shadow {
								obstacle.ObjFlags |= object.FlagShadow
							}
							obstacle.Shape.Kind = ShapeKindCircle
							obstacle.Shape.Circle.R, obstacle.Shape.Circle.R2 = radius, radius*radius
							for _, record := range []*Object{unit, target, obstacle} {
								if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(record)) <= math.MaxUint32 {
									t.Fatalf("native indexed record below 4 GiB: %p", record)
								}
								s.Map.AddObjectToIndex(record)
							}
							projection, admitted := pointOnLineReference57C8A0(from, to, center)
							dx := pointOnLineReferenceOp57C8A0(float64(projection.X), -float64(center.X), '+')
							dy := pointOnLineReferenceOp57C8A0(float64(projection.Y), -float64(center.Y), '+')
							distance := pointOnLineReferenceOp57C8A0(pointOnLineReferenceOp57C8A0(dy, dy, '*'), pointOnLineReferenceOp57C8A0(dx, dx, '*'), '+')
							want := !shadow || !admitted || distance > float64(radius*radius)
							if !s.MapTraceRayAt(from, to, nil, nil, 9) {
								t.Fatal("empty wall map unexpectedly blocks fixture")
							}
							logic, other, frame := s.Rand.Logic.Index(), s.Rand.Other.Index(), s.Frame()
							if got := s.MapTraceVision(unit, target); got != want {
								t.Errorf("real indexed vision=%t want=%t projection=%v center=%v squared-distance=%g", got, want, projection, center, distance)
							}
							if got := s.CanInteract(unit, target, 0); got != want {
								t.Errorf("real CanInteract=%t want=%t", got, want)
							}
							head := update.AIStackHead()
							head.SetArgs(types.Ptf(-37, -129), target)
							beforeUnit, beforeUpdate, beforeTarget, beforeObstacle := *unit, *update, *target, *obstacle
							if s.MonsterActionRefresh50A910(unit) != 0 {
								t.Fatal("real native target refresh failed")
							}
							if want {
								beforeUpdate.AIStack[0].SetArgs(to, target)
							} else {
								beforeUpdate.AIStack[0].Args[2] = 0
							}
							if *update != beforeUpdate {
								t.Errorf("native refresh retained/cleared target incorrectly: visible=%t head=%+v want=%+v", want, *head, beforeUpdate.AIStack[0])
							}
							// Map visits may advance Field62; no other record
							// field, health, stack tail, frame or RNG may change.
							for index, pair := range [][2]*Object{{&beforeUnit, unit}, {&beforeTarget, target}, {&beforeObstacle, obstacle}} {
								actual := *pair[1]
								actual.Field62 = pair[0].Field62
								if !bytes.Equal(moveToRawSnapshot5443F0(unsafe.Pointer(&actual), unsafe.Sizeof(actual)), moveToRawSnapshot5443F0(unsafe.Pointer(pair[0]), unsafe.Sizeof(*pair[0]))) {
									t.Errorf("indexed record %d changed outside visitation token", index)
								}
							}
							if s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other || s.Frame() != frame {
								t.Fatal("projection/visibility/refresh consumed RNG or changed frame")
							}
						})
					}
				}
			}
		}
	}
}
