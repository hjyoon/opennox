package server

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// GAME.EXE 00547CA2, 00547CE4..00547D01 retains each FMUL/FADD at
// precision 53, then FSTS/FSTPS stores binary32 under gameplay ToZero.
// big.Float and the independent integer spill encoder deliberately share
// neither production arithmetic nor production spill/Nextafter helpers.
func monsterDodgeArithmeticReference547C50(a, b float64, multiply bool) float64 {
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
		panic("bounded original dodge intermediates must be exactly binary64")
	}
	return value
}

func monsterDodgeDestinationReference547C50(sample float64, speed float32, side int, direction byte, position types.Pointf) types.Pointf {
	raw := monsterDodgeArithmeticReference547C50(sample, float64(speed), true)
	distance := math.Float32frombits(monsterThreatSpillReference5475D5(raw))
	// 00547CAF compares the unspilled register against 15, not the FSTS.
	if raw > 15 {
		distance = 15
	}
	if side < 50 {
		distance = -distance
	}
	cos, sin := SinCosDir(direction)
	coordinate := func(component, location float32) float32 {
		product := monsterDodgeArithmeticReference547C50(float64(distance), float64(component), true)
		sum := monsterDodgeArithmeticReference547C50(product, float64(location), false)
		return math.Float32frombits(monsterThreatSpillReference5475D5(sum))
	}
	return types.Ptf(coordinate(-sin, position.X), coordinate(cos, position.Y))
}

func monsterDodgePointEqual547C50(a, b types.Pointf) bool {
	equal := func(a, b float32) bool {
		return math.Float32bits(a) == math.Float32bits(b) || math.IsNaN(float64(a)) && math.IsNaN(float64(b))
	}
	return equal(a.X, b.X) && equal(a.Y, b.Y)
}

func monsterDodgeAssertNative547C50(t *testing.T, sample float64, speed float32, side int, direction byte, position types.Pointf) {
	t.Helper()
	s, unit, update := monsterHealthRetreatNativeFixture547807(t)
	unit.PosVec, unit.NewPos, unit.SpeedCur, unit.Direction1 = position, position, speed, Dir16(direction)
	want := monsterDodgeDestinationReference547C50(sample, speed, side, direction, position)
	before := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
	healthBefore, base := *unit.HealthData, update.AIStack[0]
	logic, other, frame, fps := s.Rand.Logic.Index(), s.Rand.Other.Index(), s.Frame(), s.TickRate()
	var order []string
	runtime := MonsterMainRuntime547210{
		RandomFloat: func(min, max float32) float64 {
			if min != 2 || max != 3 {
				t.Fatal("original float RNG bounds changed")
			}
			order = append(order, "float")
			return sample
		},
		RandomInt: func(min, max int) int {
			if min != 0 || max != 100 {
				t.Fatal("original integer RNG bounds changed")
			}
			order = append(order, "integer")
			return side
		},
		TraceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
			order = append(order, "ray")
			if !monsterDodgePointEqual547C50(from, position) || !monsterDodgePointEqual547C50(to, want) || flags != MapTraceFlag1 {
				t.Errorf("ray from=%v destination=%08x/%08x flags=%x, original destination=%08x/%08x",
					from, math.Float32bits(to.X), math.Float32bits(to.Y), flags, math.Float32bits(want.X), math.Float32bits(want.Y))
			}
			return true
		},
		TraceObstacles: func(got *Object, from, to types.Pointf) bool {
			order = append(order, "obstacles")
			if got != unit || !monsterDodgePointEqual547C50(from, position) || !monsterDodgePointEqual547C50(to, want) {
				t.Error("obstacle probe lost the cached origin, native unit or original destination")
			}
			return true
		},
		TileAt: func(to types.Pointf) int {
			order = append(order, "tile")
			if !monsterDodgePointEqual547C50(to, want) {
				t.Error("tile probe lost the original destination")
			}
			return 0
		},
	}
	if !s.monsterMainCheckDodgeables547C50(unit, runtime) ||
		!reflect.DeepEqual(order, []string{"float", "integer", "ray", "obstacles", "tile"}) {
		t.Fatalf("original first successful probe order changed: %v", order)
	}
	if update.AIStackInd != 2 || update.AIStack[0] != base || update.AIStack[1].Type() != ai.DEPENDENCY_TIME ||
		update.AIStack[1].ArgU32(0) != frame+fps || update.AIStackHead().Type() != ai.ACTION_DODGE ||
		!monsterDodgePointEqual547C50(update.AIStackHead().ArgPos(0), want) || update.AIStackHead().ArgU32(2) != 0 {
		t.Errorf("original chopped native DODGE stack changed: %+v, want destination=%08x/%08x",
			update.GetAIStack(), math.Float32bits(want.X), math.Float32bits(want.Y))
	}
	if !bytes.Equal(before, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
		*unit.HealthData != healthBefore || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other || s.Frame() != frame || s.TickRate() != fps {
		t.Fatal("precision correction changed native object, health, clock or a game RNG")
	}
}

func TestMonsterMainDodgePrecision547C50NativeBoundary(t *testing.T) {
	for _, tc := range []struct {
		name      string
		sample    float64
		speed     float32
		side      int
		direction byte
		position  types.Pointf
	}{
		{"distance-spill-right", 2.0000002, 4, 50, 0, types.Ptf(0, 0)},
		{"distance-spill-left", 2.0000002, 4, 49, 0, types.Ptf(0, 0)},
		{"distance-spill-negative-speed", 2.0000002, -4, 50, 0, types.Ptf(0, 0)},
		{"destination-spill-positive", 2, 4, 50, 0, types.Ptf(0, math.Float32frombits(0x3f800001))},
		{"destination-spill-negative", 2, 4, 49, 0, types.Ptf(0, math.Float32frombits(0xbf800001))},
		{"retained-add-positive-tiny-negative", 2, 4, 50, 0, types.Ptf(0, -math.SmallestNonzeroFloat32)},
		{"retained-add-negative-tiny-positive", 2, 4, 49, 0, types.Ptf(0, math.SmallestNonzeroFloat32)},
		{"subnormal-spill-right", 2.75, math.SmallestNonzeroFloat32, 50, 0, types.Ptf(0, 0)},
		{"subnormal-spill-left", 2.75, math.SmallestNonzeroFloat32, 49, 0, types.Ptf(0, 0)},
		{"finite-overflow-chop", 3, -math.MaxFloat32, 50, 0, types.Ptf(0, -math.MaxFloat32)},
		{"cap-retained-below", math.Nextafter(2.5, 0), 6, 50, 0, types.Ptf(0, 0)},
		{"cap-equal-control", 2.5, 6, 50, 0, types.Ptf(0, 0)},
		{"cap-retained-above-control", math.Nextafter(2.5, math.Inf(1)), 6, 49, 0, types.Ptf(0, 0)},
		{"positive-zero-control", 2, 0, 50, 0, types.Ptf(0, 0)},
		{"negative-zero-control", 2, math.Float32frombits(0x80000000), 49, 0, types.Ptf(0, 0)},
		{"unordered-speed-control", 2, float32(math.NaN()), 49, 19, types.Ptf(13, 27)},
		{"infinite-speed-capped-control", 2, float32(math.Inf(1)), 50, 17, types.Ptf(0, 0)},
		{"infinite-negative-speed-control", 2, float32(math.Inf(-1)), 49, 17, types.Ptf(0, 0)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			monsterDodgeAssertNative547C50(t, tc.sample, tc.speed, tc.side, tc.direction, tc.position)
		})
	}
}

func TestMonsterMainDodgePrecision547C50NativeAllDirections(t *testing.T) {
	for direction := 0; direction < 256; direction++ {
		t.Run(fmt.Sprintf("%03d", direction), func(t *testing.T) {
			// Both side choices, non-exact distance spills and live position
			// additions exercise the original table at every valid index.
			for _, side := range []int{49, 50} {
				monsterDodgeAssertNative547C50(t, 2.23456789, math.Float32frombits(0x3fc00001), side, byte(direction),
					types.Ptf(math.Float32frombits(0x42c80001), math.Float32frombits(0xc3480001)))
			}
		})
	}
}

func TestMonsterMainDodgePrecision547C50NativeDefaultRNG(t *testing.T) {
	s, unit, update := monsterHealthRetreatNativeFixture547807(t)
	s.Rand.Logic = prand.New(0)
	reference := prand.New(0)
	other := s.Rand.Other.Index()
	const count = 4096
	for i := 0; i < count; i++ {
		unit.PosVec, unit.NewPos = types.Ptf(100.00001, -200.00002), types.Ptf(100.00001, -200.00002)
		unit.SpeedCur, unit.Direction1 = math.Float32frombits(0x3fc00001), Dir16(byte(i))
		update.AIStackInd = 0
		update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_FIGHT)}
		// 00416030 with bounds 2..3 is exact in binary64: original
		// sealed table DWORD * binary32 scale, then +2. One step each
		// for float and side; no substitution for the production RNG.
		word := reference.Int(0, 0x7fff)
		sample := float64(word)*float64(math.Float32frombits(0x38000100)) + 2
		side := reference.IntClamp(0, 100)
		want := monsterDodgeDestinationReference547C50(sample, unit.SpeedCur, side, byte(unit.Direction1), unit.PosVec)
		calls := 0
		if !s.monsterMainCheckDodgeables547C50(unit, MonsterMainRuntime547210{
			TraceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
				calls++
				if from != unit.PosVec || !monsterDodgePointEqual547C50(to, want) || flags != MapTraceFlag1 {
					t.Errorf("table case=%d side=%d destination=%08x/%08x want=%08x/%08x", i, side,
						math.Float32bits(to.X), math.Float32bits(to.Y), math.Float32bits(want.X), math.Float32bits(want.Y))
				}
				return true
			},
			TraceObstacles: func(got *Object, from, to types.Pointf) bool {
				return got == unit && from == unit.PosVec && monsterDodgePointEqual547C50(to, want)
			},
			TileAt: func(types.Pointf) int { return 0 },
		}) || calls != 1 || !monsterDodgePointEqual547C50(update.AIStackHead().ArgPos(0), want) ||
			s.Rand.Logic.Index() != reference.Index() || s.Rand.Other.Index() != other {
			t.Fatalf("table case=%d changed original destination/probe count/RNG order: calls=%d logic=%d want=%d", i, calls, s.Rand.Logic.Index(), reference.Index())
		}
	}
}

func TestMonsterMainDodgePrecision547C50NativeProgressBoundary(t *testing.T) {
	for _, base := range []ai.ActionType{ai.ACTION_FIGHT, ai.ACTION_RETREAT} {
		t.Run(base.String(), func(t *testing.T) {
			s, unit, update := monsterHealthRetreatNativeFixture547807(t)
			unit.PosVec, unit.NewPos, unit.SpeedCur, unit.Direction1 = types.Ptf(0, 0), types.Ptf(0, 0), 4, 0
			*unit.HealthData = HealthData{Cur: 100, Field2: 100, Max: 100}
			update.AIStack[0] = AIStackItem{Action: uint32(base)}
			update.AIStackInd = 1
			update.AIStack[1].Action = uint32(ai.ACTION_MOVE_TO)
			update.Field124 = 0
			floats, sides, rays, tiles := 0, 0, 0, 0
			if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
				RandomFloat: func(min, max float32) float64 {
					floats++
					if min != 2 || max != 3 {
						t.Fatal("dodge float bounds changed")
					}
					return 2.0000002
				},
				RandomInt: func(min, max int) int {
					sides++
					if min != 0 || max != 100 {
						t.Error("clear original dodge incorrectly reached WAIT-duration RNG")
					}
					if base == ai.ACTION_RETREAT && sides == 1 {
						return 0 // original 33-percent probe branch
					}
					return 50
				},
				TraceRay: func(_, to types.Pointf, _ MapTraceFlags) bool {
					rays++
					// A ray boundary at exactly eight admits the original
					// chopped destination, not a nearest-even overshoot.
					return to.Y <= 8
				},
				TraceObstacles: func(*Object, types.Pointf, types.Pointf) bool { return true },
				TileAt:         func(types.Pointf) int { tiles++; return 0 },
			}) || update.AIStackHead().Type() != ai.ACTION_DODGE || update.AIStackHead().ArgPos(0) != types.Ptf(0, 8) {
				t.Errorf("original progress tail lost first accepted DODGE: stack=%+v", update.GetAIStack())
			}
			wantSides := 1
			if base == ai.ACTION_RETREAT {
				wantSides++
			}
			if floats != 1 || sides != wantSides || rays != 1 || tiles != 1 {
				t.Errorf("precision changed subsequent probes/RNG: float=%d integer=%d ray=%d tile=%d", floats, sides, rays, tiles)
			}
		})
	}
}
