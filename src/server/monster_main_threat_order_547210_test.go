package server

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterMainThreat547210GuardProbePrecedesBlink(t *testing.T) {
	s, unit, update, enemy := newMonsterMainFleeTest547210(t)
	s.SetFrame(112)
	update.StatusFlags, update.Field376 = object.MonStatusCanCastSpells, 1
	update.AIStackInd = 1
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	update.AIStack[1].Action = uint32(ai.ACTION_FIGHT)
	step := 0
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		EnemyAggro: func(got *Object, radius float32) *Object {
			if step != 0 || got != unit || radius != 100 {
				t.Fatal("GUARD query order/arguments changed")
			}
			step++
			return &Object{ObjClass: object.ClassPlayer}
		},
		Distance: func(got, target *Object) float64 {
			if step != 1 || got != unit || target != enemy || !update.StatusFlags.Has(object.MonStatusInjured) ||
				unit.Obj130 != enemy || unit.Field131 != 11 || unit.Frame134 != 112 {
				t.Fatal("flee ran before GUARD stimulus")
			}
			step++
			return 20
		},
		CastSpell: func(int32, *Object, *SpellAcceptArg) {
			if step != 2 {
				t.Fatal("Blink preceded distance")
			}
			step++
		},
		RandomInt: func(int, int) int {
			if step != 3 {
				t.Fatal("RNG preceded Blink")
			}
			step++
			return 5
		},
	}) || step != 4 || update.Field371 != 117 || update.AIStackInd != 1 {
		t.Fatal("MainAI did not retain GUARD -> distance -> immediate Blink -> cooldown order")
	}
}

func TestMonsterMainThreat547210FearStillPreemptsGuardAndBlink(t *testing.T) {
	s, unit, update, _ := newMonsterMainFleeTest547210(t)
	s.SetFrame(112)
	unit.Buffs = 1 << ENCHANT_AFRAID
	update.StatusFlags, update.Field376 = object.MonStatusCanCastSpells, 1
	update.AIStackInd = 1
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	update.AIStack[1].Action = uint32(ai.ACTION_FIGHT)
	if !s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{
		EnemyAggro: func(*Object, float32) *Object { t.Fatal("GUARD ran after a fear early return"); return nil },
		CastSpell:  func(int32, *Object, *SpellAcceptArg) { t.Fatal("Blink ran after a fear early return") },
		RandomInt:  func(int, int) int { t.Fatal("RNG ran after a fear early return"); return 0 },
	}) || update.AIStackInd != 3 || update.AIStack[2].Type() != ai.DEPENDENCY_IS_ENCHANTED ||
		update.AIStackHead().Type() != ai.ACTION_FLEE {
		t.Fatal("fear prefix lost precedence over threat responses")
	}
}
