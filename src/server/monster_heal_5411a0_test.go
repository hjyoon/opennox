package server

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterHealCast5411A0 struct {
	caster *Object
	spell  spell.ID
	target *Object
}

func monsterHealTestHooks5411A0(candidates []*Object, casts *[]monsterHealCast5411A0) monsterHealHooks5411A0 {
	return monsterHealHooks5411A0{
		frame: func() uint32 { return 64 },
		quest: func() bool { return false },
		eachInRect: func(_ types.Rectf, fn func(*Object) bool) {
			for _, candidate := range candidates {
				if !fn(candidate) {
					return
				}
			}
		},
		isEnemy:     func(*Object, *Object) bool { return false },
		canInteract: func(*Object, *Object, int) bool { return true },
		cast: func(caster *Object, id spell.ID, target *Object) {
			*casts = append(*casts, monsterHealCast5411A0{caster: caster, spell: id, target: target})
		},
	}
}

func monsterHealTestObject5411A0(t *testing.T) *Object {
	unit := monsterActionTestObject50A910(t)
	unit.ObjFlags = object.FlagEnabled
	unit.HealthData = &HealthData{Cur: 100, Max: 100}
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	return unit
}

func TestMonsterHealSomeone5411A0CapabilityGate(t *testing.T) {
	unit := monsterHealTestObject5411A0(t)
	var casts []monsterHealCast5411A0
	if got := monsterHealSomeone5411A0(unit, monsterHealTestHooks5411A0(nil, &casts)); got != 0 || len(casts) != 0 {
		t.Fatalf("result/casts = %d/%d, want 0/0", got, len(casts))
	}
}

func TestMonsterHealSomeone5411A0SelfFirst(t *testing.T) {
	unit := monsterHealTestObject5411A0(t)
	unit.HealthData = &HealthData{Cur: 49, Max: 100}
	unit.UpdateDataMonster().StatusFlags = object.MonStatusCanCastSpells | object.MonStatusCanHealSelf | object.MonStatusCanHealOthers
	ally := monsterHealTestObject5411A0(t)
	ally.HealthData = &HealthData{Cur: 1, Max: 100}
	var casts []monsterHealCast5411A0
	if got := monsterHealSomeone5411A0(unit, monsterHealTestHooks5411A0([]*Object{ally}, &casts)); got != 0 {
		t.Fatalf("result = %d, want 0 for self heal", got)
	}
	if len(casts) != 1 || casts[0] != (monsterHealCast5411A0{caster: unit, spell: monsterHealSpell5411A0, target: unit}) {
		t.Fatalf("casts = %#v, want one self heal", casts)
	}
}

func TestMonsterHealSomeone5411A0ChoosesLastEligibleAlly(t *testing.T) {
	unit := monsterHealTestObject5411A0(t)
	unit.UpdateDataMonster().StatusFlags = object.MonStatusCanCastSpells | object.MonStatusCanHealOthers
	first := monsterHealTestObject5411A0(t)
	first.HealthData = &HealthData{Cur: 1, Max: 100}
	last := monsterHealTestObject5411A0(t)
	last.HealthData = &HealthData{Cur: 2, Max: 100}
	var casts []monsterHealCast5411A0
	if got := monsterHealSomeone5411A0(unit, monsterHealTestHooks5411A0([]*Object{first, last}, &casts)); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if len(casts) != 1 || casts[0].target != last || casts[0].spell != monsterHealSpell5411A0 {
		t.Fatalf("casts = %#v, want last eligible ally", casts)
	}
}

func TestMonsterHealSomeone5411A0CastingGate(t *testing.T) {
	unit := monsterHealTestObject5411A0(t)
	unit.UpdateDataMonster().StatusFlags = object.MonStatusCanCastSpells | object.MonStatusCanHealSelf
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
	unit.HealthData = &HealthData{Cur: 1, Max: 100}
	var casts []monsterHealCast5411A0
	if got := monsterHealSomeone5411A0(unit, monsterHealTestHooks5411A0(nil, &casts)); got != 0 || len(casts) != 0 {
		t.Fatalf("result/casts = %d/%d, want 0/0", got, len(casts))
	}
}
