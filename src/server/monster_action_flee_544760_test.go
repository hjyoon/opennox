package server

import (
	"testing"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func fleeMonsterTestObject544760(t *testing.T) *Object {
	t.Helper()
	unit := passiveMonsterTestObject547210(t)
	unit.PosVec = types.Ptf(100, 200)
	unit.SpeedBase = 2
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_FLEE)}
	update.AIStack[0].SetArgs(types.Ptf(300, 400), uint32(0))
	return unit
}

func TestMonsterActionRunState534750(t *testing.T) {
	s := new(Server)
	unit := fleeMonsterTestObject544760(t)
	update := unit.UpdateDataMonster()
	update.StatusFlags = 0
	s.MonsterActionRunStart534750(unit)
	if !update.StatusFlags.Has(0x4000) {
		t.Fatal("run start did not set RUNNING")
	}
	s.MonsterActionRunEnd534780(unit)
	if update.StatusFlags.Has(0x4000) {
		t.Fatal("run end did not clear RUNNING")
	}
	update.StatusFlags = 0x10000
	s.MonsterActionRunStart534750(unit)
	if update.StatusFlags.Has(0x4000) {
		t.Fatal("NEVER_RUN was ignored")
	}
	update.StatusFlags = 0xc000
	s.MonsterActionRunEnd534780(unit)
	if !update.StatusFlags.Has(0x4000) {
		t.Fatal("ALWAYS_RUN was ignored")
	}
}

func TestMonsterActionFlee544760StationaryPops(t *testing.T) {
	unit := fleeMonsterTestObject544760(t)
	unit.SpeedBase = 0
	popped := 0
	monsterActionFlee544760(unit, monsterActionFleeHooks544760{pop: func() int { popped++; return 0 }})
	if popped != 1 {
		t.Fatalf("pop count = %d", popped)
	}
}

func TestMonsterActionFlee544760GeneratesAndMoves(t *testing.T) {
	unit := fleeMonsterTestObject544760(t)
	update := unit.UpdateDataMonster()
	update.Field70 = 80
	var generatedTarget types.Pointf
	moves, audio := 0, 0
	monsterActionFlee544760(unit, monsterActionFleeHooks544760{
		frame:    func() uint32 { return 100 },
		tickRate: func() uint32 { return 30 },
		generatePath: func(path []types.Pointf, got *Object, target *types.Pointf) int {
			if len(path) != 32 || got != unit {
				t.Fatalf("path args = %d/%p", len(path), got)
			}
			generatedTarget = *target
			path[0], path[1] = unit.PosVec, types.Ptf(120, 220)
			return 2
		},
		actuallyMove: func(*Object) bool { moves++; return false },
		moveAudio:    func(*Object) { audio++ },
		pop:          func() int { return 0 },
	})
	if generatedTarget != (types.Ptf(300, 400)) || update.Field2 != 2 || update.Field67 != 0 || update.Field70 != 100 || moves != 1 || audio != 1 {
		t.Fatalf("flee state = target %v fields %d/%d/%d calls %d/%d", generatedTarget, update.Field2, update.Field67, update.Field70, moves, audio)
	}
}

func TestMonsterActionFlee544760EnemyRefreshesTarget(t *testing.T) {
	unit := fleeMonsterTestObject544760(t)
	update := unit.UpdateDataMonster()
	enemy := &Object{PosVec: types.Ptf(110, 200)}
	update.CurrentEnemy = enemy
	update.FleeRange = 50
	update.Field2 = 3
	update.Field70 = 80
	monsterActionFlee544760(unit, monsterActionFleeHooks544760{
		frame:        func() uint32 { return 100 },
		tickRate:     func() uint32 { return 30 },
		actuallyMove: func(*Object) bool { return false },
		moveAudio:    func(*Object) {},
		pop:          func() int { return 0 },
	})
	if update.AIStackHead().ArgPos(0) != enemy.PosVec || update.Field2 != 0 {
		t.Fatalf("enemy target/path = %v/%d", update.AIStackHead().ArgPos(0), update.Field2)
	}
}
