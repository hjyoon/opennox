package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterActionTestObject50A910(t *testing.T) *Object {
	t.Helper()
	obj := new(Object)
	update := new(MonsterUpdateData)
	obj.ObjClass = object.ClassMonster
	obj.UpdateData = unsafe.Pointer(update)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= uintptr(^uint32(0)) {
		t.Fatalf("test object address = %#x, want native high address", uintptr(unsafe.Pointer(obj)))
	}
	return obj
}

func monsterActionSetTarget50A910(item *AIStackItem, target *Object) {
	item.Args[2] = uintptr(unsafe.Pointer(target))
}

func TestMonsterActionRefresh50A910ObjectArgumentMetadata(t *testing.T) {
	want := map[ai.ActionType]uint8{
		ai.ACTION_ESCORT:                         2,
		ai.ACTION_MOVE_TO:                        2,
		ai.ACTION_FAR_MOVE_TO:                    2,
		ai.ACTION_MISSILE_ATTACK:                 2,
		ai.ACTION_CAST_SPELL_ON_OBJECT:           2,
		ai.ACTION_CAST_DURATION_SPELL:            2,
		ai.ACTION_FLEE:                           2,
		ai.ACTION_FACE_OBJECT:                    1,
		ai.ACTION_MOVE_TO_HOME:                   2,
		ai.DEPENDENCY_ALIVE:                      1,
		ai.DEPENDENCY_CAN_SEE:                    1,
		ai.DEPENDENCY_CANNOT_SEE:                 1,
		ai.DEPENDENCY_BLOCKED_LINE_OF_FIRE:       1,
		ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION: 2,
		ai.DEPENDENCY_OBJECT_FARTHER_THAN:        2,
		ai.DEPENDENCY_OBJECT_CLOSER_THAN:         2,
		ai.DEPENDENCY_NO_NEW_ENEMY:               1,
	}
	for action, got := range monsterActionObjectArgMask50A910 {
		if got != want[ai.ActionType(action)] {
			t.Errorf("action %d object mask = %#x, want %#x", action, got, want[ai.ActionType(action)])
		}
	}
}

func TestMonsterActionRefresh50A910ClearsDestroyedPointers(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	destroyed := monsterActionTestObject50A910(t)
	destroyed.ObjFlags = object.FlagDestroyed
	update := unit.UpdateDataMonster()
	update.PreferredEnemy = destroyed
	update.AIStackInd = 1
	update.AIStack[0].Action = uint32(ai.ACTION_FACE_OBJECT)
	update.AIStack[0].Args[0] = uintptr(unsafe.Pointer(destroyed))
	update.AIStack[1].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
	monsterActionSetTarget50A910(&update.AIStack[1], destroyed)

	if got := monsterActionRefresh50A910(unit, func(*Object, *Object, int) bool { return true }); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if update.PreferredEnemy != nil || update.AIStack[0].Args[0] != 0 || update.AIStack[1].Args[2] != 0 {
		t.Fatalf("destroyed pointers survived: preferred=%p first=%#x second=%#x", update.PreferredEnemy, update.AIStack[0].Args[0], update.AIStack[1].Args[2])
	}
}

func TestMonsterActionRefresh50A910RefreshesTrackedPositions(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	target := monsterActionTestObject50A910(t)
	target.PosVec = types.Pointf{X: 12.5, Y: -7.25}
	update := unit.UpdateDataMonster()
	update.CurrentEnemy = target
	update.AIStackInd = 3
	update.AIStack[0].Action = uint32(ai.ACTION_ESCORT)
	monsterActionSetTarget50A910(&update.AIStack[0], target)
	update.AIStack[1].Action = uint32(ai.ACTION_MOVE_TO)
	monsterActionSetTarget50A910(&update.AIStack[1], target)
	update.AIStack[2].Action = uint32(ai.ACTION_FIGHT)
	update.AIStack[3].Action = uint32(ai.ACTION_MISSILE_ATTACK)
	monsterActionSetTarget50A910(&update.AIStack[3], target)
	var calls int
	if got := monsterActionRefresh50A910(unit, func(gotUnit, gotTarget *Object, flags int) bool {
		calls++
		return gotUnit == unit && gotTarget == target && flags == 0
	}); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if calls != 2 {
		t.Fatalf("canInteract calls = %d, want 2", calls)
	}
	wantX, wantY := uintptr(math.Float32bits(target.PosVec.X)), uintptr(math.Float32bits(target.PosVec.Y))
	for i := range update.AIStack[:4] {
		if item := &update.AIStack[i]; item.Args[0] != wantX || item.Args[1] != wantY {
			t.Errorf("stack %d position = %#x/%#x, want %#x/%#x", i, item.Args[0], item.Args[1], wantX, wantY)
		}
	}
}

func TestMonsterActionRefresh50A910MoveTargetGate(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	target := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_MOVE_TO)
	monsterActionSetTarget50A910(&update.AIStack[0], target)
	monsterActionRefresh50A910(unit, func(*Object, *Object, int) bool { return false })
	if update.AIStack[0].Args[2] != 0 {
		t.Fatalf("unreachable move target = %#x, want cleared", update.AIStack[0].Args[2])
	}
}

func TestMonsterActionRefresh50A910NegativeStack(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	unit.UpdateDataMonster().AIStackInd = -1
	if got := monsterActionRefresh50A910(unit, func(*Object, *Object, int) bool {
		t.Fatal("canInteract called for empty stack")
		return false
	}); got != -1 {
		t.Fatalf("result = %d, want -1", got)
	}
}
