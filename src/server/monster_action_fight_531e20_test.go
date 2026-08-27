package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func fightMonsterTestObject531E20(t *testing.T) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_FIGHT)}
	update.AIStack[0].SetArgs(types.Ptf(300, 400), uint32(100))
	update.MonsterDef = &MonsterDef{MeleeAttackRange112: 15}
	return unit
}

func TestMonsterActionFightStart531E20ExactOrderAndState(t *testing.T) {
	unit := fightMonsterTestObject531E20(t)
	update := unit.UpdateDataMonster()
	target := &Object{}
	update.CurrentEnemy = target
	update.ScriptChangeFocus = ScriptCallback{Flags: 0xa5, Func: 17}
	var sounds [17]uint32
	sounds[5] = 707
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	var events []string
	monsterActionFightStart531E20(unit, MonsterActionFightStartRuntime531E20{
		AudioEvent: func(id uint32, got *Object) {
			if id != 707 || got != unit || update.StatusFlags.Has(object.MonStatusAlert) {
				t.Fatalf("audio state = %d/%p/%v", id, got, update.StatusFlags)
			}
			events = append(events, "audio")
		},
		ScriptCallback: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) {
			if block != &update.ScriptChangeFocus || caller != target || trigger != unit || event != NoxEventMonsterFightStart ||
				update.StatusFlags.Has(object.MonStatusAlert) {
				t.Fatalf("script args/state = %p/%p/%p/%v/%v", block, caller, trigger, event, update.StatusFlags)
			}
			events = append(events, "script")
		},
		CopyFrameCounter: func() {
			if !update.StatusFlags.Has(object.MonStatusAlert) || update.StatusFlags.Has(object.MonStatusRunning) {
				t.Fatalf("copy status = %v", update.StatusFlags)
			}
			events = append(events, "copy")
		},
		UpdateSight: func(got *Object) {
			if got != unit || !update.StatusFlags.Has(object.MonStatusAlert) || update.StatusFlags.Has(object.MonStatusRunning) {
				t.Fatalf("sight state = %p/%v", got, update.StatusFlags)
			}
			events = append(events, "sight")
		},
	})
	want := []string{"audio", "script", "copy", "sight"}
	if len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
	if !update.StatusFlags.Has(object.MonStatusAlert) || !update.StatusFlags.Has(object.MonStatusRunning) {
		t.Fatalf("final status = %v", update.StatusFlags)
	}
}

func TestMonsterActionFightStartEnd531E90RunPolicies(t *testing.T) {
	unit := fightMonsterTestObject531E20(t)
	update := unit.UpdateDataMonster()
	update.StatusFlags = object.MonStatusNeverRun
	monsterActionFightStart531E20(unit, MonsterActionFightStartRuntime531E20{})
	if !update.StatusFlags.Has(object.MonStatusAlert) || update.StatusFlags.Has(object.MonStatusRunning) {
		t.Fatalf("never-run start = %v", update.StatusFlags)
	}
	update.StatusFlags = object.MonStatusAlert | object.MonStatusRunning | object.MonStatusAlwaysRun
	monsterActionFightEnd531E90(unit)
	if update.StatusFlags.Has(object.MonStatusAlert) || !update.StatusFlags.Has(object.MonStatusRunning) {
		t.Fatalf("always-run end = %v", update.StatusFlags)
	}
	update.StatusFlags = object.MonStatusAlert | object.MonStatusRunning
	monsterActionFightEnd531E90(unit)
	if update.StatusFlags.HasAny(object.MonStatusAlert | object.MonStatusRunning) {
		t.Fatalf("ordinary end = %v", update.StatusFlags)
	}
}

func fightHooks531EC0(frame uint32, events *[]ai.ActionType, pushed *[]*AIStackItem) monsterActionFightHooks531EC0 {
	return monsterActionFightHooks531EC0{
		frame:    func() uint32 { return frame },
		tickRate: func() uint32 { return 30 },
		push: func(action ai.ActionType, args ...any) *AIStackItem {
			*events = append(*events, action)
			item := &AIStackItem{Action: uint32(action)}
			item.SetArgs(args...)
			*pushed = append(*pushed, item)
			return item
		},
		pop: func() int {
			*events = append(*events, ai.ACTION_INVALID)
			return 0
		},
	}
}

func TestMonsterActionFight531EC0MeleeSchedule(t *testing.T) {
	unit := fightMonsterTestObject531E20(t)
	target := &Object{PosVec: types.Ptf(321, 432), HealthData: &HealthData{Cur: 10, Max: 10}}
	unit.UpdateDataMonster().CurrentEnemy = target
	var events []ai.ActionType
	var pushed []*AIStackItem
	if !monsterActionFight531EC0(unit, fightHooks531EC0(120, &events, &pushed)) {
		t.Fatal("melee fight was not handled")
	}
	want := []ai.ActionType{
		ai.DEPENDENCY_NO_NEW_ENEMY,
		ai.DEPENDENCY_ALIVE,
		ai.DEPENDENCY_CAN_SEE,
		ai.ACTION_MELEE_ATTACK,
		ai.ACTION_FACE_OBJECT,
		ai.DEPENDENCY_OBJECT_FARTHER_THAN,
		ai.ACTION_MOVE_TO,
	}
	if len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
	targetArg := uintptr(unsafe.Pointer(target))
	if unit.UpdateDataMonster().AIStack[0].ArgU32(2) != 120 ||
		pushed[0].Args[0] != targetArg || pushed[1].Args[0] != targetArg ||
		pushed[2].Args[0] != targetArg || pushed[4].Args[0] != targetArg ||
		pushed[5].ArgF32(0) != 15 || pushed[5].Args[2] != targetArg ||
		pushed[6].ArgPos(0) != target.PosVec || pushed[6].Args[2] != targetArg {
		t.Fatalf("scheduled args = %#v", pushed)
	}
}

func TestMonsterActionFight531EC0TimeoutAndLostTarget(t *testing.T) {
	unit := fightMonsterTestObject531E20(t)
	var events []ai.ActionType
	var pushed []*AIStackItem
	if !monsterActionFight531EC0(unit, fightHooks531EC0(401, &events, &pushed)) ||
		len(events) != 1 || events[0] != ai.ACTION_INVALID {
		t.Fatalf("timeout events = %v", events)
	}

	unit = fightMonsterTestObject531E20(t)
	update := unit.UpdateDataMonster()
	update.Field300 = 77
	update.Field98 = 77
	update.Field97 = 9
	events = nil
	pushed = nil
	hooks := fightHooks531EC0(120, &events, &pushed)
	hooks.findDeadTarget = func(pos types.Pointf, netCode uint32) bool {
		return pos == (types.Ptf(300, 400)) && netCode == 77
	}
	if !monsterActionFight531EC0(unit, hooks) || len(events) != 1 || events[0] != ai.ACTION_INVALID || update.Field97 != 0 {
		t.Fatalf("lost-target state = %v/%d", events, update.Field97)
	}
}

func TestMonsterActionFight531EC0RejectsUnportedAttackBranchesBeforeMutation(t *testing.T) {
	for _, setup := range []func(*Object, *MonsterUpdateData){
		func(_ *Object, update *MonsterUpdateData) { update.StatusFlags |= object.MonStatusCanCastSpells },
		func(_ *Object, update *MonsterUpdateData) { update.MonsterDef.MissileName148[0] = 'x' },
	} {
		unit := fightMonsterTestObject531E20(t)
		update := unit.UpdateDataMonster()
		update.CurrentEnemy = &Object{HealthData: &HealthData{Cur: 1, Max: 1}}
		setup(unit, update)
		before := update.AIStack[0]
		var events []ai.ActionType
		var pushed []*AIStackItem
		if monsterActionFight531EC0(unit, fightHooks531EC0(120, &events, &pushed)) {
			t.Fatal("unported branch was handled")
		}
		if update.AIStack[0] != before || len(events) != 0 {
			t.Fatalf("rejected branch mutated state: %#v/%v", update.AIStack[0], events)
		}
	}
}
