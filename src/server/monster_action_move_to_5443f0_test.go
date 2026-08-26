package server

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func moveToMonsterTestObject5443F0(t *testing.T) *Object {
	t.Helper()
	unit := passiveMonsterTestObject547210(t)
	unit.PosVec = types.Ptf(100, 200)
	unit.SpeedBase = 2
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_MOVE_TO)}
	update.AIStack[0].SetArgs(types.Ptf(300, 400), uint32(0))
	return unit
}

func moveToHooks5443F0(t *testing.T, events *[]ai.ActionType) monsterActionMoveToHooks5443F0 {
	t.Helper()
	return monsterActionMoveToHooks5443F0{
		frame:    func() uint32 { return 100 },
		tickRate: func() uint32 { return 30 },
		random: func(minimum, maximum int) int {
			return (minimum + maximum) / 2
		},
		setMovePath: func(*Object, types.Pointf) bool { return false },
		pathReset:   func() bool { return false },
		moveAudio:   func(*Object) {},
		push: func(action ai.ActionType, args ...any) *AIStackItem {
			*events = append(*events, action)
			item := &AIStackItem{Action: uint32(action)}
			item.SetArgs(args...)
			return item
		},
		pop: func() int {
			*events = append(*events, ai.ACTION_INVALID)
			return 0
		},
	}
}

func TestMonsterActionMoveTo5443F0StationaryPops(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	unit.SpeedBase = 0
	var events []ai.ActionType
	monsterActionMoveTo5443F0(unit, moveToHooks5443F0(t, &events))
	if len(events) != 1 || events[0] != ai.ACTION_INVALID {
		t.Fatalf("events = %v, want pop", events)
	}
}

func TestMonsterActionMoveTo5443F0ArrivalFacesAndPops(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	var events []ai.ActionType
	hooks := moveToHooks5443F0(t, &events)
	hooks.setMovePath = func(got *Object, target types.Pointf) bool {
		if got != unit || target != (types.Ptf(300, 400)) {
			t.Fatalf("move target = %p/%v", got, target)
		}
		return true
	}
	monsterActionMoveTo5443F0(unit, hooks)
	if len(events) != 1 || events[0] != ai.ACTION_INVALID || unit.Direction2 != DirFromVec(types.Ptf(200, 200)) {
		t.Fatalf("arrival = events %v direction %d", events, unit.Direction2)
	}
}

func TestMonsterActionMoveTo5443F0FailureStack(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	update := unit.UpdateDataMonster()
	update.Field71 = 2
	var events []ai.ActionType
	var pushed []*AIStackItem
	hooks := moveToHooks5443F0(t, &events)
	hooks.setMovePath = func(*Object, types.Pointf) bool { return true }
	hooks.push = func(action ai.ActionType, args ...any) *AIStackItem {
		events = append(events, action)
		item := &AIStackItem{Action: uint32(action)}
		item.SetArgs(args...)
		pushed = append(pushed, item)
		return item
	}
	monsterActionMoveTo5443F0(unit, hooks)
	want := []ai.ActionType{ai.DEPENDENCY_TIME, ai.ACTION_RANDOM_WALK, ai.ACTION_WAIT}
	for i := range want {
		if len(events) != len(want) || events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
	if !update.StatusFlags.Has(object.MonStatusFrustrated) || pushed[0].ArgU32(0) != 190 || pushed[2].ArgU32(0) != 122 {
		t.Fatalf("failure state = %#x/%#v", update.StatusFlags, pushed)
	}
}

func TestMonsterActionMoveTo5443F0EscortRunBands(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	update := unit.UpdateDataMonster()
	update.StatusFlags = 0
	update.AIStackInd = 1
	update.AIStack[0].Action = uint32(ai.ACTION_ESCORT)
	update.AIStack[1] = AIStackItem{Action: uint32(ai.ACTION_MOVE_TO)}
	update.AIStack[1].SetArgs(types.Ptf(300, 200), uint32(0))
	update.Field329 = 10
	var events []ai.ActionType
	monsterActionMoveTo5443F0(unit, moveToHooks5443F0(t, &events))
	if !update.StatusFlags.Has(object.MonStatusRunning) {
		t.Fatal("distant escort target did not enable running")
	}
	update.AIStack[1].SetArgs(types.Ptf(110, 200), uint32(0))
	monsterActionMoveTo5443F0(unit, moveToHooks5443F0(t, &events))
	if update.StatusFlags.Has(object.MonStatusRunning) {
		t.Fatal("near escort target did not disable running")
	}
}
