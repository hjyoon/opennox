package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestAIEnemyDistanceDependency546A70NativeOriginalDistanceAndGroup(t *testing.T) {
	s, unit, target, health := aiAliveDependencyNative546A70(t)
	// Literal GAME.EXE outcomes: 00546E8E reads the entry-cached update's
	// CurrentEnemy, then uses 00546B3E (<= or unordered, nil fails).
	// 00546E99 passes nil; otherwise FCOMP tests only C0, so ordered >=
	// passes and unordered fails. Both retain 004E6C00's surface distance.
	for _, tc := range []struct {
		name                                  string
		unitX, unitY, targetX, targetY        uint32
		unitKind, unitR, unitW, unitH         uint32
		targetKind, targetR, targetW, targetH uint32
		radius                                uint32
		missing, far, close                   bool
	}{
		{name: "equal-zero-clamped", radius: 0x3c23d70a, far: true, close: true},
		{name: "above-zero-radius", radius: 0, far: true},
		{name: "minimum-below-one", radius: 0x3f800000, close: true},
		{name: "exact-one", targetX: 0x3f800000, radius: 0x3f800000, far: true, close: true},
		{name: "retained-above-one", targetX: 0x3f800000, targetY: 0x39000000, radius: 0x3f800000, far: true},
		{name: "retained-below-one", targetX: 0x3f7fffff, targetY: 0x39a7c5ac, radius: 0x3f800000, close: true},
		{name: "unspilled-x-difference", unitX: 0x4b800000, targetX: 0xbf800000, radius: 0x4b800000, far: true},
		{name: "unspilled-y-difference", unitY: 0x4b800000, targetY: 0xbf800000, radius: 0x4b800000, far: true},
		{name: "equal-three-four-five", targetX: 0x40400000, targetY: 0x40800000, radius: 0x40a00000, far: true, close: true},
		{name: "circle-surfaces-equal", targetX: 0x41200000, unitKind: 2, unitR: 0x40000000, targetKind: 2, targetR: 0x40400000, radius: 0x40a00000, far: true, close: true},
		{name: "retained-source-circle", targetX: 0x3f800000, targetY: 0x39000000, unitKind: 2, unitR: 0x3f000000, radius: 0x3f000000, far: true},
		{name: "retained-target-circle", targetX: 0x3f800000, targetY: 0x39000000, targetKind: 2, targetR: 0x3f000000, radius: 0x3f000000, far: true},
		{name: "retained-both-circles", targetX: 0x3f800000, targetY: 0x39000000, unitKind: 2, unitR: 0x3e800000, targetKind: 2, targetR: 0x3e800000, radius: 0x3f000000, far: true},
		{name: "retained-source-box", targetX: 0x3f800000, targetY: 0x39000000, unitKind: 3, unitW: 0x3f800000, unitH: 0x3f000000, radius: 0x3f000000, far: true},
		{name: "retained-target-box", targetX: 0x3f800000, targetY: 0x39000000, targetKind: 3, targetW: 0x3f000000, targetH: 0x3f800000, radius: 0x3f000000, far: true},
		{name: "box-surfaces-equal", targetX: 0x41200000, unitKind: 3, unitW: 0x40c00000, unitH: 0x40800000, targetKind: 3, targetW: 0x40000000, targetH: 0x41000000, radius: 0x40400000, far: true, close: true},
		{name: "box-width-unordered-selects-height", targetX: 0x41200000, unitKind: 3, unitW: 0x7fc12345, unitH: 0x40800000, radius: 0x41000000, far: true, close: true},
		{name: "box-height-unordered-clamps", targetX: 0x41200000, unitKind: 3, unitW: 0x40800000, unitH: 0x7fc12345, radius: 0x3c23d70a, far: true, close: true},
		{name: "unknown-full-DWORD-kind", targetX: 0x3f800000, targetY: 0x39000000, unitKind: 0x60000002, unitR: 0x3f000000, radius: 0x3f800000, far: true},
		{name: "overlap-clamps", targetX: 0x3f800000, unitKind: 2, unitR: 0x40000000, radius: 0x3c23d70a, far: true, close: true},
		{name: "negative-circle-radii", targetX: 0x41200000, unitKind: 2, unitR: 0xc0000000, targetKind: 2, targetR: 0xc0400000, radius: 0x41700000, far: true, close: true},
		{name: "source-qnan-position-clamps", unitX: 0x7fc12345, radius: 0x3c23d70a, far: true, close: true},
		{name: "target-snan-position-clamps", targetY: 0xff800001, radius: 0x3c23d70a, far: true, close: true},
		{name: "circle-qnan-clamps", targetX: 0x41200000, targetKind: 2, targetR: 0x7fc12345, radius: 0x3c23d70a, far: true, close: true},
		{name: "infinity-minus-infinity-clamps", unitX: 0x7f800000, targetX: 0x7f800000, radius: 0x3c23d70a, far: true, close: true},
		{name: "positive-infinity-distance", targetX: 0x7f800000, radius: 0x3f800000, far: true},
		{name: "equal-infinite-radius", targetX: 0x7f800000, radius: 0x7f800000, far: true, close: true},
		{name: "finite-below-infinite-radius", targetX: 0x7f7fffff, unitX: 0xff7fffff, radius: 0x7f800000, close: true},
		{name: "maximum-finite-diagonal", targetX: 0x7f7fffff, targetY: 0x7f7fffff, radius: 0x7f7fffff, far: true},
		{name: "positive-qnan-radius", targetX: 0x3f800000, radius: 0x7fc12345, close: true},
		{name: "negative-qnan-radius", targetX: 0x3f800000, radius: 0xffc12345, close: true},
		{name: "snan-radius", targetX: 0x3f800000, radius: 0x7f800001, close: true},
		{name: "negative-infinite-radius", targetX: 0x3f800000, radius: 0xff800000, far: true},
		{name: "source-then-target-chop", targetX: 0x3f800000, unitKind: 2, unitR: 0x7f7fffff, targetKind: 2, targetR: 0xff7fffff, radius: 0x65000000, far: true, close: true},
		{name: "reverse-shapes-chop-and-clamp", targetX: 0x3f800000, unitKind: 2, unitR: 0xff7fffff, targetKind: 2, targetR: 0x7f7fffff, radius: 0x3c23d70a, far: true, close: true},
		{name: "nil-enemy", targetX: 0x3f800000, radius: 0x3f800000, missing: true, far: true},
		{name: "nil-enemy-qnan-radius", targetX: 0x3f800000, radius: 0x7fc12345, missing: true, far: true},
		{name: "nil-enemy-unordered-position", unitX: 0x7fc12345, targetY: 0xff800001, radius: 0xffc12345, missing: true, far: true},
	} {
		for _, typ := range []ai.ActionType{ai.DEPENDENCY_ENEMY_FARTHER_THAN, ai.DEPENDENCY_ENEMY_CLOSER_THAN} {
			valid := tc.far
			if typ == ai.DEPENDENCY_ENEMY_CLOSER_THAN {
				valid = tc.close
			}
			for _, shape := range aiLocationDependencyShapes546A70 {
				t.Run(tc.name+"/"+typ.String()+"/"+shape, func(t *testing.T) {
					unit.PosVec = types.Pointf{X: math.Float32frombits(tc.unitX), Y: math.Float32frombits(tc.unitY)}
					target.PosVec = types.Pointf{X: math.Float32frombits(tc.targetX), Y: math.Float32frombits(tc.targetY)}
					unit.Shape = server.Shape{Kind: server.ShapeKind(tc.unitKind)}
					target.Shape = server.Shape{Kind: server.ShapeKind(tc.targetKind)}
					unit.Shape.Circle.R, target.Shape.Circle.R = math.Float32frombits(tc.unitR), math.Float32frombits(tc.targetR)
					unit.Shape.Box.W, unit.Shape.Box.H = math.Float32frombits(tc.unitW), math.Float32frombits(tc.unitH)
					target.Shape.Box.W, target.Shape.Box.H = math.Float32frombits(tc.targetW), math.Float32frombits(tc.targetH)
					// Neither original branch inspects class, flags or health.
					unit.ObjFlags, target.ObjFlags = object.FlagDead|object.FlagNoUpdate|object.FlagDestroyed, object.FlagDead|object.FlagNoUpdate|object.FlagDestroyed
					target.ObjClass, target.HealthData = object.ClassFood, nil
					beforeHealth := *health
					unitWords, targetWords := aiObjectDistanceInputWords546B13(unit), aiObjectDistanceInputWords546B13(target)
					update := aiLocationDependencyStack546A70(unit, typ, shape, tc.radius, 0x89abcdef, 0xffc12345)
					index := 1
					if shape == "or-false" || shape == "or-true-before-location" {
						index = 2
					}
					// Args2 is a different, valid native identity, never the enemy.
					// Reading it instead of cached CurrentEnemy gives wrong results.
					update.AIStack[index].Args[2] = uintptr(unsafe.Pointer(unit))
					update.CurrentEnemy = target
					if tc.missing {
						update.CurrentEnemy = nil
					}
					before := *update
					s.AI.StackChanged = false
					(&aiData{s: s}).nox_xxx_mobActionDependency(unit)
					want := before
					wantPopped := !valid && (shape == "and" || shape == "or-false")
					if wantPopped {
						want.AIStackInd = 0
						want.Field2, want.Field67, want.Field74, want.Field91 = 0, 0, 0, 0
						want.Field120_1, want.Field120_2, want.Field120_3 = 0, 0, 0
						want.Field124, want.Field137 = 702, 702
					}
					if *update != want || s.AI.StackChanged != wantPopped {
						t.Fatalf("original enemy distance condition/pop differs: index=%d/%d changed=%t/%t condition=%s radius=%08x enemy=%p", update.AIStackInd, want.AIStackInd, s.AI.StackChanged, wantPopped, typ, tc.radius, before.CurrentEnemy)
					}
					if aiObjectDistanceInputWords546B13(unit) != unitWords || aiObjectDistanceInputWords546B13(target) != targetWords || *health != beforeHealth || target.HealthData != nil || target.ObjClass != object.ClassFood || target.ObjFlags != object.FlagDead|object.FlagNoUpdate|object.FlagDestroyed {
						t.Fatal("enemy distance dependency changed raw object inputs, class, flags or health")
					}
				})
			}
		}
	}
}
