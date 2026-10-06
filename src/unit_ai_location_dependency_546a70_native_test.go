package opennox

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

var aiLocationDependencyShapes546A70 = []string{
	"and", "or-false", "or-true-before-location", "or-true-after-location",
}

func aiLocationDependencyStack546A70(unit *server.Object, typ ai.ActionType, shape string, radius, x, y uint32) *server.MonsterUpdateData {
	update := unit.UpdateDataMonster()
	*update = server.MonsterUpdateData{
		Field2: 71, Field67: 72, Field74: 73, Field91: 74,
		Field97: 75, Field98: 76, Field101: 77, Field102: 78,
		Field120_0: 81, Field120_1: 82, Field120_2: 83, Field120_3: 84,
		Field124: 11, Field137: 12, Field136_1: 85, Field136_2: 86,
	}
	high := uintptr(0x12345678)
	high <<= 32
	update.AIStack[0] = server.AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{7, high | 8, 9, 10}}
	condition := server.AIStackItem{
		Action: uint32(typ), Field5: 0xfedcba98,
		Args: [4]uintptr{high | uintptr(radius), high | 0x89abcdef, high | uintptr(x), high | uintptr(y)},
	}
	index := 1
	switch shape {
	case "and":
		update.AIStack[index] = condition
	case "or-false", "or-true-before-location":
		// TIME 900 is genuinely true at fixture frame 702, even when the
		// unit position is NaN. TIME 0 is genuinely false.
		deadline := uintptr(0)
		if shape == "or-true-before-location" {
			deadline = 900
		}
		update.AIStack[index] = server.AIStackItem{Action: uint32(ai.DEPENDENCY_TIME), Args: [4]uintptr{high | deadline}}
		index++
		update.AIStack[index] = condition
	case "or-true-after-location":
		update.AIStack[index] = condition
		index++
		update.AIStack[index] = server.AIStackItem{Action: uint32(ai.DEPENDENCY_TIME), Args: [4]uintptr{high | 900}}
	default:
		panic("unknown original location dependency stack shape")
	}
	if shape != "and" {
		index++
		update.AIStack[index] = server.AIStackItem{Action: uint32(ai.DEPENDENCY_OR)}
	}
	index++
	update.AIStack[index] = server.AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{900, high | 88, 99, 100}}
	update.AIStackInd = int8(index)
	update.AIStack[20] = server.AIStackItem{Action: 17, Args: [4]uintptr{high | 1, high | 2, high | 3, high | 4}, Field5: 18}
	return update
}

func TestAILocationDependency546A70NativeOriginalDistanceAndGroup(t *testing.T) {
	s, unit, _, _ := aiAliveDependencyNative546A70(t)
	// GAME.EXE 00546B63/00546B92 retains X/Y differences, Y-square,
	// X-square, sum and square root without an m32real spill. CRT startup
	// selects 53-bit precision; 0043E2C1 selects chop on each frame.
	// Literal branch expectations follow FCOMP C0 (far) and C0|C3
	// (close), not the port's distance helper or its float32 expression.
	for _, tc := range []struct {
		name               string
		unitX, unitY, x, y uint32
		radius             uint32
		far, close         bool
	}{
		{"equal-zero", 0, 0, 0, 0, 0, true, true},
		{"signed-zero", 0x80000000, 0, 0, 0x80000000, 0x80000000, true, true},
		{"zero-below-one", 0, 0, 0, 0, 0x3f800000, false, true},
		{"zero-above-negative-radius", 0, 0, 0, 0, 0xbf800000, true, false},
		{"equal-one", 0, 0, 0x3f800000, 0, 0x3f800000, true, true},
		{"retained-above-one", 0, 0, 0x3f800000, 0x39000000, 0x3f800000, true, false},
		{"retained-below-one", 0, 0, 0x3f7fffff, 0x39a7c5ac, 0x3f800000, false, true},
		{"equal-three-four-five", 0, 0, 0x40400000, 0x40800000, 0x40a00000, true, true},
		{"negative-three-four-five", 0, 0, 0xc0400000, 0xc0800000, 0x40a00000, true, true},
		{"translated-three-four-five", 0x41200000, 0xc1a00000, 0x41500000, 0xc1800000, 0x40a00000, true, true},
		{"unspilled-x-subtraction", 0x4b800000, 0, 0xbf800000, 0, 0x4b800000, true, false},
		{"unspilled-y-subtraction", 0, 0x4b800000, 0, 0xbf800000, 0x4b800000, true, false},
		{"chop-positive-subtraction", 1, 0, 0x3f800000, 0, 0x3f800000, false, true},
		{"chop-negative-subtraction", 0x80000001, 0, 0xbf800000, 0, 0x3f800000, false, true},
		{"chop-positive-addition", 0x80000001, 0, 0x3f800000, 0, 0x3f800000, true, true},
		{"equal-smallest-subnormal", 0, 0, 1, 0, 1, true, true},
		{"retained-subnormal-diagonal", 0, 0, 1, 1, 1, true, false},
		{"equal-largest-subnormal", 0, 0, 0x007fffff, 0, 0x007fffff, true, true},
		{"equal-smallest-normal", 0, 0, 0x00800000, 0, 0x00800000, true, true},
		{"square-does-not-overflow", 0, 0, 0x7f7fffff, 0, 0x7f7fffff, true, true},
		{"maximum-finite-diagonal", 0, 0, 0x7f7fffff, 0x7f7fffff, 0x7f7fffff, true, false},
		{"finite-below-infinite-radius", 0, 0, 0x3f800000, 0, 0x7f800000, false, true},
		{"infinite-distance-equal-radius", 0, 0, 0x7f800000, 0, 0x7f800000, true, true},
		{"infinite-distance-above-finite-radius", 0, 0, 0xff800000, 0, 0x3f800000, true, false},
		{"finite-above-negative-infinite-radius", 0, 0, 0x3f800000, 0, 0xff800000, true, false},
		{"positive-qnan-radius", 0, 0, 0x3f800000, 0, 0x7fc12345, false, true},
		{"negative-qnan-radius", 0, 0, 0x3f800000, 0, 0xffc12345, false, true},
		{"snan-radius", 0, 0, 0x3f800000, 0, 0x7f800001, false, true},
		{"qnan-location-x", 0, 0, 0x7fc12345, 0, 0x3f800000, false, true},
		{"snan-location-y", 0, 0, 0, 0xff800001, 0x3f800000, false, true},
		{"qnan-origin-x", 0x7fc12345, 0, 0, 0, 0x3f800000, false, true},
		{"snan-origin-y", 0, 0x7f800001, 0, 0, 0x3f800000, false, true},
		{"infinity-minus-infinity", 0x7f800000, 0, 0x7f800000, 0, 0x3f800000, false, true},
		{"opposite-infinities", 0x7f800000, 0, 0xff800000, 0, 0x7f800000, true, true},
	} {
		for _, typ := range []ai.ActionType{ai.DEPENDENCY_LOCATION_FARTHER_THAN, ai.DEPENDENCY_LOCATION_CLOSER_THAN} {
			valid := tc.far
			if typ == ai.DEPENDENCY_LOCATION_CLOSER_THAN {
				valid = tc.close
			}
			for _, shape := range aiLocationDependencyShapes546A70 {
				t.Run(tc.name+"/"+typ.String()+"/"+shape, func(t *testing.T) {
					unit.PosVec = types.Pointf{X: math.Float32frombits(tc.unitX), Y: math.Float32frombits(tc.unitY)}
					// Neither distance condition reads flags or health.
					unit.ObjFlags = object.FlagDead | object.FlagNoUpdate | object.FlagDestroyed
					update := aiLocationDependencyStack546A70(unit, typ, shape, tc.radius, tc.x, tc.y)
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
						t.Fatalf("original location condition/pop differs: index=%d/%d heard=%d/%d deadline=%d/%d reset=%d,%d/%d,%d changed=%t/%t", update.AIStackInd, want.AIStackInd, update.Field97, want.Field97, update.Field101, want.Field101, update.Field124, update.Field137, want.Field124, want.Field137, s.AI.StackChanged, wantPopped)
					}
					if math.Float32bits(unit.PosVec.X) != tc.unitX || math.Float32bits(unit.PosVec.Y) != tc.unitY {
						t.Fatal("location dependency changed native unit position words")
					}
				})
			}
		}
	}
}
