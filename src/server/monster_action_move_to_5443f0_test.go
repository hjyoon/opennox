package server

import (
	"testing"
	"unsafe"

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

func moveToHomeMonsterTestObject544950(t *testing.T) *Object {
	t.Helper()
	unit := moveToMonsterTestObject5443F0(t)
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_MOVE_TO_HOME)
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

func TestMonsterActionMoveToHome544950UsesSharedUpdateBody(t *testing.T) {
	unit := moveToHomeMonsterTestObject544950(t)
	var events []ai.ActionType
	hooks := moveToHooks5443F0(t, &events)
	pathCalls, audioCalls := 0, 0
	hooks.setMovePath = func(got *Object, target types.Pointf) bool {
		pathCalls++
		if got != unit || target != (types.Ptf(300, 400)) {
			t.Fatalf("move target = %p/%v, want %p/%v", got, target, unit, types.Ptf(300, 400))
		}
		return true
	}
	hooks.moveAudio = func(got *Object) {
		audioCalls++
		if got != unit {
			t.Fatalf("audio unit = %p, want %p", got, unit)
		}
	}

	if monsterActionMoveTo5443F0(unit, hooks) {
		t.Fatal("ACTION_MOVE_TO wrapper admitted ACTION_MOVE_TO_HOME")
	}
	if !monsterActionMoveToHome544950(unit, hooks) {
		t.Fatal("ACTION_MOVE_TO_HOME update was not handled")
	}
	if pathCalls != 1 || audioCalls != 1 || len(events) != 1 || events[0] != ai.ACTION_INVALID {
		t.Fatalf("calls/events = path:%d audio:%d events:%v", pathCalls, audioCalls, events)
	}
	if unit.Direction2 != DirFromVec(types.Ptf(200, 200)) {
		t.Fatalf("arrival direction = %d", unit.Direction2)
	}
}

func TestMonsterActionMoveToHome544950RejectsInvalidAdmission(t *testing.T) {
	var events []ai.ActionType
	hooks := moveToHooks5443F0(t, &events)
	for _, unit := range []*Object{
		nil,
		{},
		{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(new(MonsterUpdateData))},
	} {
		if monsterActionMoveToHome544950(unit, hooks) {
			t.Fatalf("invalid unit %#v was handled", unit)
		}
	}
	unit := moveToHomeMonsterTestObject544950(t)
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_GUARD)
	if monsterActionMoveToHome544950(unit, hooks) {
		t.Fatal("wrong stack head was handled")
	}
	hooks.pop = nil
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_MOVE_TO_HOME)
	if monsterActionMoveToHome544950(unit, hooks) {
		t.Fatal("missing required runtime hook was handled")
	}
	if len(events) != 0 {
		t.Fatalf("invalid admission emitted events: %v", events)
	}
}

func TestMonsterActionMoveToHome544950PreservesNativeObjectPointer(t *testing.T) {
	unit := moveToHomeMonsterTestObject544950(t)
	want := uintptr(unsafe.Pointer(unit))
	if unsafe.Sizeof(uintptr(0)) == 8 && want <= uintptr(^uint32(0)) {
		t.Fatalf("unit pointer = %#x, want value above PE32 range", want)
	}
	update := unit.UpdateDataMonster()
	update.StatusFlags = 0
	s := new(Server)
	s.MonsterActionRunStart534750(unit)
	if !update.StatusFlags.Has(object.MonStatusRunning) {
		t.Fatal("MOVE_TO_HOME start did not enable running")
	}

	var events []ai.ActionType
	hooks := moveToHooks5443F0(t, &events)
	hooks.setMovePath = func(got *Object, target types.Pointf) bool {
		if uintptr(unsafe.Pointer(got)) != want {
			t.Fatalf("path unit = %#x, want %#x", uintptr(unsafe.Pointer(got)), want)
		}
		return false
	}
	hooks.moveAudio = func(got *Object) {
		if uintptr(unsafe.Pointer(got)) != want {
			t.Fatalf("audio unit = %#x, want %#x", uintptr(unsafe.Pointer(got)), want)
		}
	}
	if !monsterActionMoveToHome544950(unit, hooks) {
		t.Fatal("native-width MOVE_TO_HOME unit was not handled")
	}
	s.MonsterActionRunEnd534780(unit)
	if update.StatusFlags.Has(object.MonStatusRunning) {
		t.Fatal("MOVE_TO_HOME end/cancel did not disable running")
	}
}
