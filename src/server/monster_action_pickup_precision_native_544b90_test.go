package server

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// GAME.EXE 00544BB4..00544BCA retains FSUB, Y-square, X-square and
// FADDP at gameplay precision 53 / ToZero before FCOMP 5625. There is
// no float32 spill. This reference uses neither production arithmetic
// helpers, Nextafter corrections nor a modified thread FPU environment.
func monsterPickupArithmeticReference544B90(a, b float64, multiply bool) float64 {
	if math.IsNaN(a) || math.IsNaN(b) || math.IsInf(a, 0) || math.IsInf(b, 0) {
		if multiply {
			return a * b
		}
		return a + b
	}
	input := func(value float64) *big.Float {
		return new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(value)
	}
	result := new(big.Float).SetPrec(53).SetMode(big.ToZero)
	if multiply {
		result.Mul(input(a), input(b))
	} else {
		result.Add(input(a), input(b))
	}
	value, accuracy := result.Float64()
	if accuracy != big.Exact {
		panic("bounded pickup intermediates must be exactly binary64")
	}
	return value
}

func monsterPickupRangeReference544B90(unit, target types.Pointf) bool {
	dx := monsterPickupArithmeticReference544B90(float64(target.X), -float64(unit.X), false)
	dy := monsterPickupArithmeticReference544B90(float64(target.Y), -float64(unit.Y), false)
	ySquared := monsterPickupArithmeticReference544B90(dy, dy, true)
	xSquared := monsterPickupArithmeticReference544B90(dx, dx, true)
	distance := monsterPickupArithmeticReference544B90(ySquared, xSquared, false)
	// 00544BD4 tests C0 alone: unordered also reaches CanInteract.
	return !(distance >= 5625)
}

func monsterPickupAssertPrecisionNative544B90(t *testing.T, position, targetPosition types.Pointf, visible bool) {
	t.Helper()
	s, unit, update := monsterHealthRetreatNativeFixture547807(t)
	food, freeFood := alloc.New(Object{})
	t.Cleanup(freeFood)
	food.ObjClass, food.ObjSubClass, food.PosVec = object.ClassFood, 0x10, targetPosition
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(food)) <= math.MaxUint32 {
		t.Fatalf("native food address below 4 GiB: %p", food)
	}
	unit.PosVec, unit.NewPos = position, position
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_PICKUP_OBJECT)}
	update.AIStack[0].SetArgs(food)
	head := update.AIStackHead()
	unitBefore := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
	updateBefore := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))
	foodBefore := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(food), unsafe.Sizeof(*food))
	healthBefore := *unit.HealthData
	logic, other, frame, fps := s.Rand.Logic.Index(), s.Rand.Other.Index(), s.Frame(), s.TickRate()
	var events []string
	result := monsterActionPickupObject544B90(unit, monsterActionPickupObjectHooks544B90{
		canInteract: func(owner, item *Object, flags int) bool {
			if owner != unit || item != food || flags != 0 || head.ArgObj(0) != food {
				t.Fatal("original visibility identity/flags or native slot changed")
			}
			events = append(events, "interact")
			return visible
		},
		placeInventory: func(owner, item *Object, a, b int) bool {
			if owner != unit || item != food || a != 1 || b != 1 {
				t.Fatal("original inventory identity/flags changed")
			}
			events = append(events, "place")
			return false // The original ignores this return value.
		},
		useByNetCode: func(owner, item *Object) int32 {
			if owner != unit || item != food {
				t.Fatal("original use lost the full native target")
			}
			events = append(events, "use")
			return -37
		},
		pop: func() int { events = append(events, "pop"); return -129 },
	})
	want := []string{"pop"}
	if monsterPickupRangeReference544B90(position, targetPosition) {
		want = []string{"interact", "pop"}
		if visible {
			want = []string{"interact", "place", "use", "pop"}
		}
	}
	if result != -129 || !reflect.DeepEqual(events, want) {
		t.Errorf("original pickup admission/prefix differs: unit=%08x/%08x target=%08x/%08x visible=%t result=%d events=%v want=-129/%v",
			math.Float32bits(position.X), math.Float32bits(position.Y), math.Float32bits(targetPosition.X), math.Float32bits(targetPosition.Y), visible, result, events, want)
	}
	if !bytes.Equal(unitBefore, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
		!bytes.Equal(updateBefore, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) ||
		!bytes.Equal(foodBefore, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(food), unsafe.Sizeof(*food))) ||
		*unit.HealthData != healthBefore || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other || s.Frame() != frame || s.TickRate() != fps {
		t.Fatal("range calculation changed native object, food, stack, health, frame or RNG")
	}
}

func TestMonsterActionPickupPrecision544B90ReferenceCalibration(t *testing.T) {
	for _, tc := range []struct {
		name     string
		unit     types.Pointf
		target   types.Pointf
		eligible bool
	}{
		{"exact-75-excluded", types.Ptf(0, 0), types.Ptf(75, 0), false},
		{"smallest-positive-delta-inside", types.Ptf(math.SmallestNonzeroFloat32, 0), types.Ptf(75, 0), true},
		{"smallest-negative-delta-outside", types.Ptf(-math.SmallestNonzeroFloat32, 0), types.Ptf(75, 0), false},
		{"negative-target-inside", types.Ptf(-math.SmallestNonzeroFloat32, 0), types.Ptf(-75, 0), true},
		{"45-60-exact-excluded", types.Ptf(0, 0), types.Ptf(45, 60), false},
		{"45-60-tiny-inside", types.Ptf(math.SmallestNonzeroFloat32, 0), types.Ptf(45, 60), true},
		{"unordered-admitted", types.Ptf(0, 0), types.Ptf(float32(math.NaN()), 0), true},
		{"infinite-excluded", types.Ptf(0, 0), types.Ptf(float32(math.Inf(1)), 0), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := monsterPickupRangeReference544B90(tc.unit, tc.target); got != tc.eligible {
				t.Fatalf("independent reference admission=%t want=%t", got, tc.eligible)
			}
		})
	}
}

func TestMonsterActionPickupPrecision544B90NativeBoundary(t *testing.T) {
	for _, tc := range []struct {
		name   string
		unit   types.Pointf
		target types.Pointf
	}{
		{"exact-75", types.Ptf(0, 0), types.Ptf(75, 0)},
		{"just-inside", types.Ptf(0, 0), types.Ptf(math.Float32frombits(0x4295ffff), 0)},
		{"just-outside", types.Ptf(0, 0), types.Ptf(math.Float32frombits(0x42960001), 0)},
		{"signed-zero", types.Ptf(math.Float32frombits(0x80000000), 0), types.Ptf(0, math.Float32frombits(0x80000000))},
		{"nonzero-spill-boundary", types.Ptf(math.Float32frombits(0x35800000), 0), types.Ptf(75, 0)},
		{"45-60-exact", types.Ptf(0, 0), types.Ptf(45, 60)},
		{"45-60-tiny-X", types.Ptf(math.SmallestNonzeroFloat32, 0), types.Ptf(45, 60)},
		{"45-60-tiny-Y", types.Ptf(0, math.SmallestNonzeroFloat32), types.Ptf(45, 60)},
		{"largest-finite-opposite-signs", types.Ptf(-math.MaxFloat32, math.MaxFloat32), types.Ptf(math.MaxFloat32, -math.MaxFloat32)},
		{"positive-infinity", types.Ptf(0, 0), types.Ptf(float32(math.Inf(1)), 0)},
		{"negative-infinity", types.Ptf(0, 0), types.Ptf(0, float32(math.Inf(-1)))},
		{"infinity-minus-infinity", types.Ptf(float32(math.Inf(1)), 0), types.Ptf(float32(math.Inf(1)), 0)},
		{"NaN-X", types.Ptf(0, 0), types.Ptf(math.Float32frombits(0x7fc54321), 0)},
		{"NaN-Y", types.Ptf(0, 0), types.Ptf(0, math.Float32frombits(0x7fc54321))},
	} {
		for _, visible := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/visible-%t", tc.name, visible), func(t *testing.T) {
				monsterPickupAssertPrecisionNative544B90(t, tc.unit, tc.target, visible)
			})
		}
	}
}

func TestMonsterActionPickupPrecision544B90NativeTinyOffsets(t *testing.T) {
	// These exact binary32 offsets include differences below, at and above
	// half a binary64 ULP at 75, where nearest-even can erase the difference.
	for _, bits := range []uint32{1, 2, 0x007fffff, 0x00800000, 0x23800000, 0x26000000, 0x28000000, 0x28800000} {
		for _, negative := range []bool{false, true} {
			for _, axis := range []string{"X", "Y"} {
				for _, inside := range []bool{false, true} {
					t.Run(fmt.Sprintf("%08x/negative-%t/%s/inside-%t", bits, negative, axis, inside), func(t *testing.T) {
						offset, target := math.Float32frombits(bits), float32(75)
						if negative {
							offset, target = -offset, -target
						}
						if !inside {
							offset = -offset
						}
						position, targetPosition := types.Ptf(offset, 0), types.Ptf(target, 0)
						if axis == "Y" {
							position, targetPosition = types.Ptf(0, offset), types.Ptf(0, target)
						}
						if got := monsterPickupRangeReference544B90(position, targetPosition); got != inside {
							t.Fatalf("independent tiny-offset calibration=%t want=%t", got, inside)
						}
						monsterPickupAssertPrecisionNative544B90(t, position, targetPosition, true)
					})
				}
			}
		}
	}
}
