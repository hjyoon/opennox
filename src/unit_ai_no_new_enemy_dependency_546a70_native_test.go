package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestAINoNewEnemyDependency546A70NativeOriginalDistanceAndGroup(t *testing.T) {
	s, unit, old, health := aiAliveDependencyNative546A70(t)
	enemy, freeEnemy := alloc.New(server.Object{})
	t.Cleanup(freeEnemy)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(enemy)) <= math.MaxUint32 {
		t.Fatalf("native new enemy below 4 GiB: %p", enemy)
	}
	// GAME.EXE table DWORD 005470E4 selects 00546D1B. A nil Args0
	// fails before cached CurrentEnemy is read; a nil enemy then passes.
	// Old and new centre-distance squares retain 53-bit/chop results,
	// without a binary32 spill or square root. FCOMP tests only C0:
	// ordered new >= old passes, and less/unordered fails.
	for _, tc := range []struct {
		name                           string
		unitX, unitY, oldX, oldY, x, y uint32
		noOld, noEnemy, same, valid    bool
	}{
		{name: "equal-zero", valid: true},
		{name: "signed-zero", unitX: 0x80000000, oldY: 0x80000000, x: 0x80000000, valid: true},
		{name: "closer-integer", oldX: 0x40000000, x: 0x3f800000},
		{name: "farther-integer", oldX: 0x3f800000, x: 0x40000000, valid: true},
		{name: "equal-three-four-five", oldX: 0x40400000, oldY: 0x40800000, x: 0x40a00000, valid: true},
		{name: "translated-three-four-five", unitX: 0x41200000, unitY: 0xc1a00000, oldX: 0x41500000, oldY: 0xc1800000, x: 0x41700000, y: 0xc1a00000, valid: true},
		{name: "retained-old-diagonal-above-one", oldX: 0x3f800000, oldY: 0x39000000, x: 0x3f800000},
		{name: "retained-new-diagonal-above-one", oldX: 0x3f800000, x: 0x3f800000, y: 0x39000000, valid: true},
		{name: "retained-new-below-one", oldX: 0x3f800000, x: 0x3f7fffff, y: 0x39a7c5ac},
		{name: "retained-old-below-one", oldX: 0x3f7fffff, oldY: 0x39a7c5ac, x: 0x3f800000, valid: true},
		{name: "unspilled-old-x-difference", unitX: 0x4b800000, oldX: 0xbf800000},
		{name: "unspilled-old-y-difference", unitY: 0x4b800000, oldY: 0xbf800000},
		{name: "unspilled-negative-old-x-difference", unitX: 0xcb800000, oldX: 0x3f800000},
		{name: "chop-new-positive-difference", unitX: 0x3f800000, x: 1},
		{name: "chop-new-negative-difference", unitX: 0xbf800000, x: 0x80000001},
		{name: "equal-smallest-subnormal", oldX: 1, x: 1, valid: true},
		{name: "retained-subnormal-diagonal", oldX: 1, oldY: 1, x: 1},
		{name: "retained-largest-subnormal-diagonal", oldX: 0x007fffff, oldY: 0x007fffff, x: 0x007fffff},
		{name: "retained-smallest-normal-diagonal", oldX: 0x00800000, oldY: 0x00800000, x: 0x00800000},
		{name: "finite-squares-do-not-overflow", oldX: 0x7f7fffff, oldY: 0x7f7fffff, x: 0x7f7fffff},
		{name: "finite-differences-do-not-overflow", unitX: 0xff7fffff, oldX: 0x7f7fffff},
		{name: "same-native-enemy", oldX: 0x3f800000, oldY: 0x39000000, same: true, valid: true},
		{name: "same-native-enemy-unordered", oldY: 0x7fc12345, same: true},
		{name: "new-positive-infinity", oldX: 0x3f800000, x: 0x7f800000, valid: true},
		{name: "new-negative-infinity", oldX: 0x3f800000, x: 0xff800000, valid: true},
		{name: "old-infinity-new-finite", oldX: 0x7f800000, x: 0x3f800000},
		{name: "equal-infinite-squares", oldX: 0x7f800000, x: 0xff800000, valid: true},
		{name: "old-qnan", oldX: 0x7fc12345, x: 0x3f800000},
		{name: "old-negative-snan", oldY: 0xff800001, x: 0x3f800000},
		{name: "new-qnan", oldX: 0x3f800000, x: 0xffc12345},
		{name: "new-snan", oldX: 0x3f800000, y: 0x7f800001},
		{name: "unit-qnan", unitX: 0x7fc12345, oldX: 0x3f800000, x: 0x40000000},
		{name: "unit-snan", unitY: 0xff800001, oldX: 0x3f800000, x: 0x40000000},
		{name: "old-infinity-minus-infinity", unitX: 0x7f800000, oldX: 0x7f800000},
		{name: "new-infinity-minus-infinity", unitY: 0xff800000, y: 0xff800000},
		{name: "nil-old", noOld: true, x: 0x3f800000},
		{name: "nil-old-and-enemy", noOld: true, noEnemy: true},
		{name: "nil-old-before-unordered-inputs", noOld: true, unitX: 0x7fc12345, y: 0xff800001},
		{name: "nil-enemy", oldX: 0x3f800000, noEnemy: true, valid: true},
		{name: "nil-enemy-before-old-qnan", oldX: 0x7fc12345, noEnemy: true, valid: true},
		{name: "nil-enemy-before-unit-snan", unitY: 0xff800001, noEnemy: true, valid: true},
	} {
		for _, shape := range aiLocationDependencyShapes546A70 {
			t.Run(tc.name+"/"+shape, func(t *testing.T) {
				unit.PosVec = types.Pointf{X: math.Float32frombits(tc.unitX), Y: math.Float32frombits(tc.unitY)}
				old.PosVec = types.Pointf{X: math.Float32frombits(tc.oldX), Y: math.Float32frombits(tc.oldY)}
				enemy.PosVec = types.Pointf{X: math.Float32frombits(tc.x), Y: math.Float32frombits(tc.y)}
				for _, obj := range []*server.Object{unit, old, enemy} {
					// Original centre distance ignores all shapes, flags and HP.
					obj.Shape = server.Shape{Kind: server.ShapeKind(0x60000003)}
					obj.Shape.Circle.R = math.Float32frombits(0x7fc12345)
					obj.Shape.Box.W, obj.Shape.Box.H = math.Float32frombits(0xff800001), math.Float32frombits(0x7f800000)
					obj.ObjFlags = object.FlagDead | object.FlagNoUpdate | object.FlagDestroyed
				}
				old.ObjClass, enemy.ObjClass = object.ClassFood, object.ClassFood
				old.HealthData, enemy.HealthData = nil, nil
				beforeHealth := *health
				unitWords, oldWords, enemyWords := aiObjectDistanceInputWords546B13(unit), aiObjectDistanceInputWords546B13(old), aiObjectDistanceInputWords546B13(enemy)
				update := aiLocationDependencyStack546A70(unit, ai.DEPENDENCY_NO_NEW_ENEMY, shape, 0, 0x89abcdef, 0xffc12345)
				index := 1
				if shape == "or-false" || shape == "or-true-before-location" {
					index = 2
				}
				update.AIStack[index].Args[0] = uintptr(unsafe.Pointer(old))
				if tc.noOld {
					update.AIStack[index].Args[0] = 0
				}
				// Args2 is not the new enemy; CurrentEnemy is entry-cached.
				update.AIStack[index].Args[2] = uintptr(unsafe.Pointer(unit))
				update.CurrentEnemy = enemy
				if tc.same {
					update.CurrentEnemy = old
				}
				if tc.noEnemy {
					update.CurrentEnemy = nil
				}
				before := *update
				s.AI.StackChanged = false
				(&aiData{s: s}).nox_xxx_mobActionDependency(unit)
				want := before
				wantPopped := !tc.valid && (shape == "and" || shape == "or-false")
				if wantPopped {
					want.AIStackInd = 0
					want.Field2, want.Field67, want.Field74, want.Field91 = 0, 0, 0, 0
					want.Field120_1, want.Field120_2, want.Field120_3 = 0, 0, 0
					want.Field124, want.Field137 = 702, 702
				}
				if *update != want || s.AI.StackChanged != wantPopped {
					t.Fatalf("original new-enemy condition/pop differs: index=%d/%d changed=%t/%t old=%p enemy=%p", update.AIStackInd, want.AIStackInd, s.AI.StackChanged, wantPopped, old, before.CurrentEnemy)
				}
				if aiObjectDistanceInputWords546B13(unit) != unitWords || aiObjectDistanceInputWords546B13(old) != oldWords || aiObjectDistanceInputWords546B13(enemy) != enemyWords || *health != beforeHealth {
					t.Fatal("new-enemy dependency changed raw object inputs or HP words")
				}
				for _, obj := range []*server.Object{old, enemy} {
					if obj.HealthData != nil || obj.ObjClass != object.ClassFood || obj.ObjFlags != object.FlagDead|object.FlagNoUpdate|object.FlagDestroyed {
						t.Fatal("new-enemy dependency changed target class, flags or HP identity")
					}
				}
			})
		}
	}
}
