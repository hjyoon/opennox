package server

import (
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func passiveShopkeeperTestObject547210(t *testing.T) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	unit.ObjSubClass = object.SubClass(object.MonsterShopkeeper)
	unit.ObjFlags = object.FlagActive | object.FlagEnabled
	unit.HealthData = &HealthData{}
	update := unit.UpdateDataMonster()
	update.Aggression = 0.5
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	return unit
}

func TestMonsterMainPassiveShopkeeper547210(t *testing.T) {
	unit := passiveShopkeeperTestObject547210(t)
	before := *unit.UpdateDataMonster()
	if !new(Server).MonsterMainPassiveShopkeeper547210(unit) {
		t.Fatal("passive shopkeeper path was not handled")
	}
	after := *unit.UpdateDataMonster()
	if after != before {
		t.Fatal("passive shopkeeper path changed monster state")
	}
}

func TestMonsterMainPassiveShopkeeper547210RejectsActiveBranches(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Object, *MonsterUpdateData)
	}{
		{name: "not shopkeeper", setup: func(unit *Object, _ *MonsterUpdateData) { unit.ObjSubClass = 0 }},
		{name: "destroyed", setup: func(unit *Object, _ *MonsterUpdateData) { unit.ObjFlags |= object.FlagDestroyed }},
		{name: "buffed", setup: func(unit *Object, _ *MonsterUpdateData) { unit.Buffs = 1 }},
		{name: "inventory", setup: func(unit *Object, _ *MonsterUpdateData) { unit.InvFirstItem = unit }},
		{name: "nonidle", setup: func(_ *Object, update *MonsterUpdateData) { update.AIStack[0].Action = uint32(ai.ACTION_GUARD) }},
		{name: "stacked", setup: func(_ *Object, update *MonsterUpdateData) { update.AIStackInd = 1 }},
		{name: "enemy", setup: func(unit *Object, update *MonsterUpdateData) { update.CurrentEnemy = unit }},
		{name: "status", setup: func(_ *Object, update *MonsterUpdateData) { update.StatusFlags = object.MonStatusCanBlock }},
		{name: "injured", setup: func(unit *Object, _ *MonsterUpdateData) { unit.HealthData = &HealthData{Cur: 4, Max: 5} }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit := passiveShopkeeperTestObject547210(t)
			tc.setup(unit, unit.UpdateDataMonster())
			if new(Server).MonsterMainPassiveShopkeeper547210(unit) {
				t.Fatal("active main-AI branch was treated as passive")
			}
		})
	}
}
