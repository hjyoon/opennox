package opennox

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Independent operation-by-operation 53-bit/chop reference, with exact
// rational sqrt bounds and integer binary32 height spills. It does not
// use the production distance, TwoSum, FMA or Nextafter corrections.
func aiEnemyDistanceReferenceSurface546E8E(unit, target *server.Object) *big.Float {
	dx := aiLocationReferenceChop53_546B63(0).Sub(aiLocationReferenceChop53_546B63(float64(unit.PosVec.X)), aiLocationReferenceChop53_546B63(float64(target.PosVec.X)))
	dy := aiLocationReferenceChop53_546B63(0).Sub(aiLocationReferenceChop53_546B63(float64(unit.PosVec.Y)), aiLocationReferenceChop53_546B63(float64(target.PosVec.Y)))
	ySquare := aiLocationReferenceChop53_546B63(0).Mul(dy, dy)
	xSquare := aiLocationReferenceChop53_546B63(0).Mul(dx, dx)
	sum := aiLocationReferenceChop53_546B63(0).Add(ySquare, xSquare)
	result := aiLocationReferenceSqrtChop53_546B63(sum)
	result.Sub(result, aiLocationReferenceChop53_546B63(aiObjectDistanceReferenceExtent546B13(&unit.Shape)))
	result.Sub(result, aiLocationReferenceChop53_546B63(aiObjectDistanceReferenceExtent546B13(&target.Shape)))
	minimum := aiLocationReferenceChop53_546B63(float64(math.Float32frombits(0x3c23d70a)))
	if result.Cmp(minimum) < 0 {
		return minimum
	}
	return result
}

func TestAIEnemyDistanceDependency546E8EIndependentComparisonModel(t *testing.T) {
	for _, shapes := range [][2]server.ShapeKind{{0, 0}, {2, 2}, {3, 3}, {2, 3}, {3, 2}, {0x60000002, 3}} {
		t.Run(fmt.Sprintf("%08x-%08x", shapes[0], shapes[1]), func(t *testing.T) {
			word := uint32(0x546e8e01)
			next := func() float32 {
				word = 1664525*word + 1013904223
				bits := word
				if bits&0x7f800000 == 0x7f800000 {
					bits ^= 0x00800000
				}
				return math.Float32frombits(bits)
			}
			for i := 0; i < 128; i++ {
				unit := server.Object{PosVec: types.Pointf{X: next(), Y: next()}}
				target := server.Object{PosVec: types.Pointf{X: next(), Y: next()}}
				for index, obj := range []*server.Object{&unit, &target} {
					obj.Shape.Kind = shapes[index]
					obj.Shape.Circle.R, obj.Shape.Box.W, obj.Shape.Box.H = next(), next(), next()
				}
				reference := aiEnemyDistanceReferenceSurface546E8E(&unit, &target)
				boundary, _ := reference.Float32()
				for _, radius := range []float32{boundary, math.Nextafter32(boundary, float32(math.Inf(-1))), math.Nextafter32(boundary, float32(math.Inf(1)))} {
					comparison := reference.Cmp(aiLocationReferenceChop53_546B63(float64(radius)))
					for _, closer := range []bool{false, true} {
						update := server.MonsterUpdateData{CurrentEnemy: &target, Field97: 71, Field101: 72}
						high := uintptr(0x12345678)
						high <<= 32
						slot := server.AIStackItem{Action: 66, Args: [4]uintptr{high | uintptr(math.Float32bits(radius)), high | 7, uintptr(unsafe.Pointer(&unit)), high | 9}, Field5: 17}
						if closer {
							slot.Action = 63
						}
						beforeSlot, beforeUpdate := slot, update
						unitWords, targetWords := aiObjectDistanceInputWords546B13(&unit), aiObjectDistanceInputWords546B13(&target)
						want := comparison >= 0
						if closer {
							want = comparison <= 0
						}
						if got := aiDependencyEnemyDistance546E8E(&unit, &update, &slot, closer); got != want {
							t.Fatalf("comparison model set=%d closer=%t radius=%08x got=%t want=%t", i, closer, math.Float32bits(radius), got, want)
						}
						if slot != beforeSlot || update != beforeUpdate || aiObjectDistanceInputWords546B13(&unit) != unitWords || aiObjectDistanceInputWords546B13(&target) != targetWords {
							t.Fatal("enemy comparison changed slot/update/object input")
						}
					}
				}
			}
		})
	}
}

func TestAIEnemyDistanceDependency546E8EMissingInputFaultPrefix(t *testing.T) {
	s, unit, target, health := aiAliveDependencyNative546A70(t)
	unit.PosVec, target.PosVec = types.Pointf{X: 1, Y: 2}, types.Pointf{X: 3, Y: 4}
	update := unit.UpdateDataMonster()
	for _, closer := range []bool{false, true} {
		label := "far"
		if closer {
			label = "close"
		}
		for _, tc := range []struct {
			name                                             string
			missingUnit, missingUpdate, missingSlot, noEnemy bool
			wantFault                                        bool
		}{
			{name: "nil-update", missingUpdate: true, wantFault: true},
			{name: "nil-update-and-unit", missingUpdate: true, missingUnit: true, wantFault: true},
			{name: "nil-update-and-slot", missingUpdate: true, missingSlot: true, wantFault: true},
			{name: "nil-all-inputs", missingUpdate: true, missingUnit: true, missingSlot: true, wantFault: true},
			{name: "nil-slot-with-enemy", missingSlot: true, wantFault: true},
			{name: "nil-unit-with-enemy", missingUnit: true, wantFault: true},
			{name: "nil-unit-and-slot-with-enemy", missingUnit: true, missingSlot: true, wantFault: true},
			{name: "nil-enemy-and-unit", noEnemy: true, missingUnit: true},
			{name: "nil-enemy-and-slot", noEnemy: true, missingSlot: true},
			{name: "nil-enemy-unit-and-slot", noEnemy: true, missingUnit: true, missingSlot: true},
			{name: "nil-enemy", noEnemy: true},
		} {
			t.Run(label+"/"+tc.name, func(t *testing.T) {
				update.CurrentEnemy = target
				if tc.noEnemy {
					update.CurrentEnemy = nil
				}
				high := uintptr(0x12345678)
				high <<= 32
				slot := &server.AIStackItem{Action: uint32(ai.DEPENDENCY_ENEMY_FARTHER_THAN), Args: [4]uintptr{high | 0x7fc12345, high | 7, uintptr(unsafe.Pointer(unit)), high | 9}, Field5: 17}
				if closer {
					slot.Action = uint32(ai.DEPENDENCY_ENEMY_CLOSER_THAN)
				}
				selectedUnit, selectedUpdate, selectedSlot := unit, update, slot
				if tc.missingUnit {
					selectedUnit = nil
				}
				if tc.missingUpdate {
					selectedUpdate = nil
				}
				if tc.missingSlot {
					selectedSlot = nil
				}
				beforeSlot, beforeUpdate, beforeHealth := *slot, *update, *health
				unitWords, targetWords := aiObjectDistanceInputWords546B13(unit), aiObjectDistanceInputWords546B13(target)
				s.AI.StackChanged = false
				var fault any
				var got bool
				func() {
					defer func() { fault = recover() }()
					got = aiDependencyEnemyDistance546E8E(selectedUnit, selectedUpdate, selectedSlot, closer)
				}()
				want := tc.noEnemy && !closer && !tc.wantFault
				if (fault != nil) != tc.wantFault || got != want {
					t.Fatalf("enemy prefix fault=%v wantFault=%t result=%t want=%t", fault, tc.wantFault, got, want)
				}
				if *slot != beforeSlot || *update != beforeUpdate || *health != beforeHealth || target.HealthData != health || s.AI.StackChanged || aiObjectDistanceInputWords546B13(unit) != unitWords || aiObjectDistanceInputWords546B13(target) != targetWords {
					t.Fatal("enemy prefix mutated native object/health/update/slot")
				}
			})
		}
	}
}

func TestAIEnemyDistanceDependency546E8ECachedUpdateAndLiveEnemy(t *testing.T) {
	_, unit, target, _ := aiAliveDependencyNative546A70(t)
	unit.PosVec, target.PosVec = types.Pointf{}, types.Pointf{X: 3}
	cached := unit.UpdateDataMonster()
	other, freeOther := alloc.New(server.MonsterUpdateData{CurrentEnemy: unit, Field97: 91, Field101: 92})
	t.Cleanup(freeOther)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(other)) <= math.MaxUint32 {
		t.Fatal("replacement native update below 4 GiB")
	}
	beforePointer := unit.UpdateData
	unit.UpdateData = unsafe.Pointer(other)
	defer func() { unit.UpdateData = beforePointer }()
	for _, closer := range []bool{false, true} {
		label := "far"
		if closer {
			label = "close"
		}
		for _, tc := range []struct {
			name  string
			enemy *server.Object
			far   bool
		}{
			{"target", target, true}, {"unit", unit, false}, {"nil", nil, true},
		} {
			t.Run(label+"/"+tc.name, func(t *testing.T) {
				cached.CurrentEnemy = tc.enemy
				slot := server.AIStackItem{Action: 66, Args: [4]uintptr{0x40000000, 7, uintptr(unsafe.Pointer(unit)), 9}, Field5: 17}
				beforeSlot, beforeCached, beforeOther := slot, *cached, *other
				want := tc.far
				if closer {
					want = tc.enemy != nil && !tc.far
				}
				if got := aiDependencyEnemyDistance546E8E(unit, cached, &slot, closer); got != want {
					t.Fatalf("entry-cached update/live enemy result=%t want=%t", got, want)
				}
				if slot != beforeSlot || *cached != beforeCached || *other != beforeOther || unit.UpdateData != unsafe.Pointer(other) {
					t.Fatal("enemy calculation changed cached/replacement update identity or slot")
				}
			})
		}
	}
}
