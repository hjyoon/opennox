package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterScriptHitFixture515A30(t *testing.T) (*Server, *Object, *MonsterUpdateData) {
	t.Helper()
	s := unitFollowTestServer5158C0(t)
	unit := monsterActionTestObject50A910(t)
	unit.serverHandle = s.handle
	update := unit.UpdateDataMonster()
	update.MonsterDef = &MonsterDef{MeleeAttackRange112: 12.5}
	update.MonsterDef.MissileName148[0] = 'x'
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_WAIT)}
	return s, unit, update
}

func TestMonsterScriptHitMissileNative515B80KeepsHighPointerAndLocationBits(t *testing.T) {
	s, unit, update := monsterScriptHitFixture515A30(t)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer %p is not above 4 GiB", unit)
	}
	pos := types.Pointf{X: math.Float32frombits(0x7fa12345), Y: math.Float32frombits(0x80000000)}
	s.MonsterScriptHitMissile515B80(unit, pos)
	if update.AIStackInd != 1 || update.AIStack[0].Type() != ai.ACTION_REPORT ||
		update.AIStack[0].ArgU32(0) != uint32(ai.ACTION_MISSILE_ATTACK) ||
		update.AIStack[1].Type() != ai.ACTION_MISSILE_ATTACK ||
		update.AIStack[1].ArgU32(0) != 0x7fa12345 ||
		update.AIStack[1].ArgU32(1) != 0x80000000 ||
		update.AIStack[1].ArgU32(2) != 0 || !s.AI.StackChanged {
		t.Fatalf("missile stack = %#v index %d, changed %t", update.AIStack[:2], update.AIStackInd, s.AI.StackChanged)
	}
	runtime.KeepAlive(unit)
}

func TestMonsterScriptHitMeleeNative515A30KeepsOriginalStack(t *testing.T) {
	s, unit, update := monsterScriptHitFixture515A30(t)
	unit.Shape.Circle.R = 3.25
	s.MonsterScriptHitMelee515A30(unit, types.Ptf(12.5, -7.25))
	if update.AIStackInd != 3 || update.AIStack[0].Type() != ai.ACTION_REPORT ||
		update.AIStack[0].ArgU32(0) != uint32(ai.ACTION_MELEE_ATTACK) ||
		update.AIStack[1].Type() != ai.ACTION_MELEE_ATTACK ||
		update.AIStack[2].Type() != ai.DEPENDENCY_LOCATION_FARTHER_THAN ||
		update.AIStack[2].ArgU32(0) != math.Float32bits(15.75) ||
		update.AIStack[2].ArgU32(2) != math.Float32bits(12.5) ||
		update.AIStack[2].ArgU32(3) != math.Float32bits(-7.25) ||
		update.AIStack[3].Type() != ai.ACTION_MOVE_TO ||
		update.AIStack[3].ArgU32(0) != math.Float32bits(12.5) ||
		update.AIStack[3].ArgU32(1) != math.Float32bits(-7.25) ||
		update.AIStack[3].ArgU32(2) != 0 || !s.AI.StackChanged {
		t.Fatalf("melee stack = %#v index %d, changed %t", update.AIStack[:4], update.AIStackInd, s.AI.StackChanged)
	}
	runtime.KeepAlive(unit)
}

func TestMonsterScriptHitNative515A30PreservesEligibility(t *testing.T) {
	s, unit, update := monsterScriptHitFixture515A30(t)
	unit.ObjFlags = object.FlagDead
	s.MonsterScriptHitMissile515B80(unit, types.Ptf(1, 2))
	s.MonsterScriptHitMelee515A30(unit, types.Ptf(1, 2))
	unit.ObjFlags = 0
	update.MonsterDef.MissileName148[0] = 0
	update.MonsterDef.MeleeAttackRange112 = 0
	s.MonsterScriptHitMissile515B80(unit, types.Ptf(1, 2))
	s.MonsterScriptHitMelee515A30(unit, types.Ptf(1, 2))
	if update.AIStackInd != 0 || update.AIStack[0].Type() != ai.ACTION_WAIT || s.AI.StackChanged {
		t.Fatalf("ineligible attack changed stack = %#v index %d, changed %t", update.AIStack[0], update.AIStackInd, s.AI.StackChanged)
	}
	unit.ObjSubClass = object.SubClass(object.MonsterNPC)
	update.WeaponEquipFlags = 0x10000
	s.MonsterScriptHitMissile515B80(unit, types.Ptf(3, 4))
	if update.AIStackInd != 1 || update.AIStack[1].Type() != ai.ACTION_MISSILE_ATTACK {
		t.Fatalf("NPC equipped missile stack = %#v index %d", update.AIStack[:2], update.AIStackInd)
	}
	runtime.KeepAlive(unit)
}
