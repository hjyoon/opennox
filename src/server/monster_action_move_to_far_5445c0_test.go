package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func farMoveToMonsterTestObject5445C0(t *testing.T) *Object {
	t.Helper()
	unit := moveToMonsterTestObject5443F0(t)
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_FAR_MOVE_TO)
	return unit
}

func TestMonsterActionFarMoveTo5445C0NativeOffsets(t *testing.T) {
	objectUpdate := unsafe.Offsetof(Object{}.UpdateData)
	monsterEnemy := unsafe.Offsetof(MonsterUpdateData{}.CurrentEnemy)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if objectUpdate != 872 || monsterEnemy != 1928 {
			t.Fatalf("native offsets = Object.UpdateData:%d Monster.CurrentEnemy:%d", objectUpdate, monsterEnemy)
		}
	} else if objectUpdate != 748 || monsterEnemy != 1196 {
		t.Fatalf("PE32 offsets = Object.UpdateData:%d Monster.CurrentEnemy:%d", objectUpdate, monsterEnemy)
	}
}

func TestMonsterActionFarMoveTo5445C0LowAggression(t *testing.T) {
	unit := farMoveToMonsterTestObject5445C0(t)
	unit.UpdateDataMonster().Aggression = 0.1
	var events []ai.ActionType
	hooks := moveToHooks5443F0(t, &events)
	pathCalls, audioCalls := 0, 0
	hooks.setMovePath = func(got *Object, target types.Pointf) bool {
		pathCalls++
		if got != unit || target != types.Ptf(300, 400) {
			t.Fatalf("move = %p/%v, want %p/(300,400)", got, target, unit)
		}
		return false
	}
	hooks.moveAudio = func(*Object) { audioCalls++ }
	if !monsterActionFarMoveTo5445C0(unit, hooks, func() int {
		t.Fatal("low aggression consulted threat notice")
		return 0
	}) {
		t.Fatal("FAR_MOVE_TO was not handled")
	}
	if pathCalls != 1 || audioCalls != 1 || len(events) != 0 {
		t.Fatalf("low-aggression calls/events = path:%d audio:%d events:%v", pathCalls, audioCalls, events)
	}
}

func TestMonsterActionFarMoveTo5445C0ThreatShortCircuits(t *testing.T) {
	for _, aggression := range []float32{0.5, 0.9} {
		unit := farMoveToMonsterTestObject5445C0(t)
		unit.UpdateDataMonster().Aggression = aggression
		var events []ai.ActionType
		hooks := moveToHooks5443F0(t, &events)
		hooks.setMovePath = func(*Object, types.Pointf) bool {
			t.Fatal("movement ran after threat notice handled action")
			return false
		}
		calls := 0
		if !monsterActionFarMoveTo5445C0(unit, hooks, func() int { calls++; return 1 }) {
			t.Fatalf("aggression %v was not handled", aggression)
		}
		if calls != 1 || len(events) != 0 {
			t.Fatalf("aggression %v: notice %d, events %v", aggression, calls, events)
		}
	}
}

func TestMonsterActionFarMoveTo5445C0NoticeCanChangeStackHead(t *testing.T) {
	unit := farMoveToMonsterTestObject5445C0(t)
	update := unit.UpdateDataMonster()
	update.Aggression = 0.5
	var events []ai.ActionType
	hooks := moveToHooks5443F0(t, &events)
	pathCalls := 0
	hooks.setMovePath = func(_ *Object, target types.Pointf) bool {
		pathCalls++
		if update.AIStackHead().Type() != ai.ACTION_ROAM || target != types.Ptf(20, 30) {
			t.Fatalf("movement used stale head: %v, target %v", update.AIStackHead().Type(), target)
		}
		return false
	}
	if !monsterActionFarMoveTo5445C0(unit, hooks, func() int {
		update.AIStackInd = 1
		update.AIStack[1] = AIStackItem{Action: uint32(ai.ACTION_ROAM)}
		update.AIStack[1].SetArgs(types.Ptf(20, 30), uint32(0))
		return 0
	}) || pathCalls != 1 {
		t.Fatalf("stack-changing notice path calls = %d", pathCalls)
	}
}

func TestMonsterActionFarMoveTo5445C0FightsCurrentEnemyThenMoves(t *testing.T) {
	unit := farMoveToMonsterTestObject5445C0(t)
	update := unit.UpdateDataMonster()
	update.Aggression = 0.9
	enemy := &Object{PosVec: types.Ptf(450, 600)}
	update.CurrentEnemy = enemy
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(enemy)) <= uintptr(^uint32(0)) {
		t.Fatalf("enemy pointer %#x did not exercise native width", uintptr(unsafe.Pointer(enemy)))
	}
	var events []ai.ActionType
	hooks := moveToHooks5443F0(t, &events)
	hooks.push = func(action ai.ActionType, args ...any) *AIStackItem {
		events = append(events, action)
		update.AIStackInd++
		item := &update.AIStack[update.AIStackInd]
		*item = AIStackItem{Action: uint32(action)}
		item.SetArgs(args...)
		return item
	}
	pathCalls := 0
	hooks.setMovePath = func(got *Object, target types.Pointf) bool {
		pathCalls++
		if got != unit || target != enemy.PosVec || update.AIStackHead().Type() != ai.ACTION_FIGHT {
			t.Fatalf("movement after FIGHT = %p/%v, head %v", got, target, update.AIStackHead().Type())
		}
		return false
	}
	if !monsterActionFarMoveTo5445C0(unit, hooks, func() int { return 0 }) {
		t.Fatal("FAR_MOVE_TO was not handled")
	}
	if len(events) != 1 || events[0] != ai.ACTION_FIGHT || pathCalls != 1 ||
		update.AIStack[1].ArgPos(0) != enemy.PosVec || update.AIStack[1].ArgU32(2) != 100 {
		t.Fatalf("fight transition = events %v path %d args %#v", events, pathCalls, update.AIStack[1].Args)
	}
}

func TestMonsterActionFarMoveTo5445C0EnemyPushFailureMovesFar(t *testing.T) {
	unit := farMoveToMonsterTestObject5445C0(t)
	update := unit.UpdateDataMonster()
	update.Aggression = 0.9
	update.CurrentEnemy = &Object{PosVec: types.Ptf(450, 600)}
	var events []ai.ActionType
	hooks := moveToHooks5443F0(t, &events)
	hooks.push = func(action ai.ActionType, args ...any) *AIStackItem {
		events = append(events, action)
		return nil
	}
	pathCalls := 0
	hooks.setMovePath = func(_ *Object, target types.Pointf) bool {
		pathCalls++
		if target != types.Ptf(300, 400) {
			t.Fatalf("target after failed FIGHT push = %v", target)
		}
		return false
	}
	if !monsterActionFarMoveTo5445C0(unit, hooks, func() int { return 0 }) ||
		len(events) != 1 || events[0] != ai.ACTION_FIGHT || pathCalls != 1 {
		t.Fatalf("failed push = events %v, path calls %d", events, pathCalls)
	}
}

func TestMonsterActionFarMoveTo5445C0EnemyClearedDuringPush(t *testing.T) {
	unit := farMoveToMonsterTestObject5445C0(t)
	update := unit.UpdateDataMonster()
	update.Aggression = 0.9
	update.CurrentEnemy = &Object{PosVec: types.Ptf(450, 600)}
	var events []ai.ActionType
	hooks := moveToHooks5443F0(t, &events)
	hooks.push = func(action ai.ActionType, _ ...any) *AIStackItem {
		events = append(events, action)
		update.AIStackInd = 1
		update.AIStack[1] = AIStackItem{Action: uint32(action)}
		update.CurrentEnemy = nil
		return &update.AIStack[1]
	}
	hooks.pop = func() int {
		events = append(events, ai.ACTION_INVALID)
		update.AIStackInd = 0
		return 0
	}
	pathCalls := 0
	hooks.setMovePath = func(_ *Object, target types.Pointf) bool {
		pathCalls++
		if target != types.Ptf(300, 400) || update.AIStackHead().Type() != ai.ACTION_FAR_MOVE_TO {
			t.Fatalf("movement after enemy disappeared = head %v target %v", update.AIStackHead().Type(), target)
		}
		return false
	}
	if !monsterActionFarMoveTo5445C0(unit, hooks, func() int { return 0 }) || pathCalls != 1 ||
		len(events) != 2 || events[0] != ai.ACTION_FIGHT || events[1] != ai.ACTION_INVALID {
		t.Fatalf("cleared enemy = events %v, path calls %d", events, pathCalls)
	}
}

func TestMonsterActionFarMoveTo5445C0RejectsInvalidAdmission(t *testing.T) {
	var events []ai.ActionType
	hooks := moveToHooks5443F0(t, &events)
	for _, unit := range []*Object{
		nil,
		{},
		{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(new(MonsterUpdateData))},
	} {
		if monsterActionFarMoveTo5445C0(unit, hooks, func() int { t.Fatal("invalid unit consulted threat"); return 0 }) {
			t.Fatalf("invalid unit %#v was handled", unit)
		}
	}
	unit := farMoveToMonsterTestObject5445C0(t)
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_MOVE_TO)
	if monsterActionFarMoveTo5445C0(unit, hooks, func() int { t.Fatal("wrong action consulted threat"); return 0 }) {
		t.Fatal("wrong action was handled")
	}
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_FAR_MOVE_TO)
	if monsterActionFarMoveTo5445C0(unit, hooks, nil) {
		t.Fatal("missing threat hook was handled")
	}
	if len(events) != 0 {
		t.Fatalf("invalid admission emitted events: %v", events)
	}
}
