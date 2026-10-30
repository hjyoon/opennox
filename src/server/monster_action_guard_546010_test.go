package server

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

type guardActionRecord546010 struct {
	typeID ai.ActionType
	args   string
}

func guardTestObject546010(t *testing.T) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	unit.ObjFlags = object.FlagActive | object.FlagEnabled
	unit.NetCode = 15
	unit.PosVec = types.Ptf(3958, 1463)
	unit.PrevPos = unit.PosVec
	unit.Direction1 = 96
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	update.AIStack[0].SetArgs(unit.PosVec, uint32(96))
	update.Aggression = 0.5
	update.Field137 = 1
	return unit
}

func guardTestHooks546010(frame uint32, records *[]guardActionRecord546010) (monsterActionGuardHooks546010, func()) {
	var items []*AIStackItem
	hooks := monsterActionGuardHooks546010{
		frame:            func() uint32 { return frame },
		tickRate:         func() uint32 { return 30 },
		noticeThreat:     func() int { return 0 },
		lookAtDamager:    func() bool { return false },
		interestingSound: func() int { return 0 },
		random:           func(min, _ int) int { return min },
		isMimic:          func() bool { return false },
		isPlant:          func() bool { return false },
		healSomeone:      func() int { return 0 },
		push: func(action ai.ActionType, args ...any) *AIStackItem {
			*records = append(*records, guardActionRecord546010{typeID: action, args: fmt.Sprint(args)})
			item := &AIStackItem{Action: uint32(action)}
			item.SetArgs(args...)
			items = append(items, item)
			return item
		},
	}
	// The original accepts Push before writing live payloads. Observe those
	// returned records after the action, retaining the existing expectations.
	return hooks, func() {
		for i, item := range items {
			var args []any
			switch item.Type() {
			case ai.ACTION_WAIT:
				args = []any{item.ArgU32(0)}
			case ai.ACTION_FIGHT, ai.ACTION_MOVE_TO:
				args = []any{item.ArgPos(0), item.ArgU32(2)}
			case ai.ACTION_FACE_LOCATION:
				args = []any{item.ArgPos(0)}
			}
			(*records)[i].args = fmt.Sprint(args)
		}
	}
}

func TestMonsterActionGuard546010War01AShopkeeper(t *testing.T) {
	unit := guardTestObject546010(t)
	var records []guardActionRecord546010
	hooks, recordArgs := guardTestHooks546010(1, &records)
	healCalls := 0
	hooks.healSomeone = func() int {
		healCalls++
		return 0
	}

	beforeUnit := *unit
	beforeUpdate := *unit.UpdateDataMonster()
	monsterActionGuard546010(unit, hooks)
	recordArgs()
	if len(records) != 0 {
		t.Fatalf("shopkeeper pushed actions: %+v", records)
	}
	if healCalls != 1 {
		t.Fatalf("heal calls = %d, want 1", healCalls)
	}
	if *unit != beforeUnit || *unit.UpdateDataMonster() != beforeUpdate {
		t.Fatal("shopkeeper guard state changed")
	}
}

func TestMonsterActionGuard546010RecentSoundOrder(t *testing.T) {
	unit := guardTestObject546010(t)
	update := unit.UpdateDataMonster()
	update.Field97 = 7
	update.Field101 = 20
	update.Field99X = 12.5
	update.Field99Y = -8.25
	var records []guardActionRecord546010
	hooks, recordArgs := guardTestHooks546010(21, &records)
	monsterActionGuard546010(unit, hooks)
	recordArgs()

	want := []guardActionRecord546010{
		{ai.DEPENDENCY_NO_INTERESTING_SOUND, "[]"},
		{ai.DEPENDENCY_NO_VISIBLE_ENEMY, "[]"},
		{ai.ACTION_WAIT, "[51]"},
		{ai.ACTION_FACE_LOCATION, "[{12.5 -8.25}]"},
	}
	if fmt.Sprint(records) != fmt.Sprint(want) {
		t.Fatalf("pushes = %+v, want %+v", records, want)
	}
	if update.Field97 != 0 {
		t.Fatalf("sound marker = %d, want 0", update.Field97)
	}
}

func TestMonsterActionGuard546010AggressiveEnemyReturnsAfterFight(t *testing.T) {
	unit := guardTestObject546010(t)
	unit.UpdateDataMonster().Aggression = 0.8
	enemy := &Object{PosVec: types.Ptf(100, 200)}
	unit.UpdateDataMonster().CurrentEnemy = enemy
	var records []guardActionRecord546010
	hooks, recordArgs := guardTestHooks546010(44, &records)
	monsterActionGuard546010(unit, hooks)
	recordArgs()

	want := []guardActionRecord546010{{ai.ACTION_FIGHT, "[{100 200} 44]"}}
	if fmt.Sprint(records) != fmt.Sprint(want) {
		t.Fatalf("pushes = %+v, want %+v", records, want)
	}
}

func TestMonsterActionGuard546010ReturnsToGuardPoint(t *testing.T) {
	unit := guardTestObject546010(t)
	unit.NetCode = 0
	unit.PosVec = types.Ptf(3900, 1400)
	var records []guardActionRecord546010
	hooks, recordArgs := guardTestHooks546010(16, &records)
	monsterActionGuard546010(unit, hooks)
	recordArgs()

	want := []guardActionRecord546010{
		{ai.DEPENDENCY_NOT_UNDER_ATTACK, "[]"},
		{ai.ACTION_MOVE_TO, "[{3958 1463} 0]"},
	}
	if fmt.Sprint(records) != fmt.Sprint(want) {
		t.Fatalf("pushes = %+v, want %+v", records, want)
	}
}
