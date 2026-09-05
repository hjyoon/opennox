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
		frame:         func() uint32 { return frame },
		tickRate:      func() uint32 { return 30 },
		buffSelf:      func(*Object) bool { return false },
		castOffensive: func(*Object, *Object) bool { return false },
		castRelated:   func(*Object, *Object) bool { return false },
		distance:      func(*Object, *Object) float64 { return 0 },
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

func TestMonsterActionFight531EC0UpdatesTimestampBeforeDeadTargetPop(t *testing.T) {
	unit := fightMonsterTestObject531E20(t)
	update := unit.UpdateDataMonster()
	update.CurrentEnemy = &Object{HealthData: &HealthData{Cur: 0, Max: 1}}
	var events []ai.ActionType
	var pushed []*AIStackItem
	hooks := fightHooks531EC0(120, &events, &pushed)
	hooks.pop = func() int {
		if got := update.AIStack[0].ArgU32(2); got != 120 {
			t.Fatalf("timestamp at pop = %d, want 120", got)
		}
		events = append(events, ai.ACTION_INVALID)
		return 0
	}
	if !monsterActionFight531EC0(unit, hooks) {
		t.Fatal("dead-target fight was not handled")
	}
	if len(events) != 1 || events[0] != ai.ACTION_INVALID || len(pushed) != 0 {
		t.Fatalf("events/pushed = %v/%v", events, pushed)
	}
}

func TestMonsterActionFight531EC0SpellOrderAndShortCircuit(t *testing.T) {
	unit := fightMonsterTestObject531E20(t)
	update := unit.UpdateDataMonster()
	target := &Object{HealthData: &HealthData{Cur: 1, Max: 1}}
	update.CurrentEnemy = target
	update.StatusFlags |= object.MonStatusCanCastSpells
	var actions []ai.ActionType
	var pushed []*AIStackItem
	hooks := fightHooks531EC0(120, &actions, &pushed)
	var calls []string
	hooks.buffSelf = func(got *Object) bool {
		if got != unit || update.AIStack[0].ArgU32(2) != 120 {
			t.Fatalf("self args/state = %p/%d", got, update.AIStack[0].ArgU32(2))
		}
		calls = append(calls, "self")
		return false
	}
	hooks.castOffensive = func(gotUnit, gotTarget *Object) bool {
		if gotUnit != unit || gotTarget != target {
			t.Fatalf("offensive args = %p/%p", gotUnit, gotTarget)
		}
		calls = append(calls, "offensive")
		return true
	}
	hooks.castRelated = func(*Object, *Object) bool {
		t.Fatal("related selector ran after offensive spell")
		return false
	}
	hooks.distance = func(*Object, *Object) float64 {
		t.Fatal("physical selector ran after offensive spell")
		return 0
	}
	if !monsterActionFight531EC0(unit, hooks) {
		t.Fatal("spell fight was not handled")
	}
	if len(calls) != 2 || calls[0] != "self" || calls[1] != "offensive" || len(actions) != 0 {
		t.Fatalf("calls/actions = %v/%v", calls, actions)
	}
}

func TestMonsterActionFight531EC0AntiMagicSkipsAllSpellSelectors(t *testing.T) {
	unit := fightMonsterTestObject531E20(t)
	update := unit.UpdateDataMonster()
	update.CurrentEnemy = &Object{HealthData: &HealthData{Cur: 1, Max: 1}}
	unit.Buffs = uint32(1) << ENCHANT_ANTI_MAGIC
	var actions []ai.ActionType
	var pushed []*AIStackItem
	hooks := fightHooks531EC0(120, &actions, &pushed)
	hooks.buffSelf = func(*Object) bool { t.Fatal("self selector ran under anti-magic"); return false }
	hooks.castOffensive = func(*Object, *Object) bool { t.Fatal("offensive selector ran under anti-magic"); return false }
	hooks.castRelated = func(*Object, *Object) bool { t.Fatal("related selector ran under anti-magic"); return false }
	if !monsterActionFight531EC0(unit, hooks) {
		t.Fatal("anti-magic fight was not handled")
	}
	if len(actions) == 0 || actions[3] != ai.ACTION_MELEE_ATTACK {
		t.Fatalf("physical actions = %v", actions)
	}
}

func TestMonsterActionFight531EC0MissileSchedule(t *testing.T) {
	unit := fightMonsterTestObject531E20(t)
	update := unit.UpdateDataMonster()
	target := &Object{PosVec: types.Ptf(321, 432), HealthData: &HealthData{Cur: 10, Max: 10}}
	update.CurrentEnemy = target
	update.MonsterDef = &MonsterDef{MissileAttackRange212: 100}
	update.MonsterDef.MissileName148[0] = 'x'
	var events []ai.ActionType
	var pushed []*AIStackItem
	hooks := fightHooks531EC0(120, &events, &pushed)
	hooks.distance = func(*Object, *Object) float64 { return 80 }
	if !monsterActionFight531EC0(unit, hooks) {
		t.Fatal("missile fight was not handled")
	}
	want := []ai.ActionType{
		ai.DEPENDENCY_NO_NEW_ENEMY,
		ai.DEPENDENCY_CAN_SEE,
		ai.ACTION_MISSILE_ATTACK,
		ai.ACTION_FACE_OBJECT,
		ai.DEPENDENCY_BLOCKED_LINE_OF_FIRE,
		ai.DEPENDENCY_OBJECT_FARTHER_THAN,
		ai.DEPENDENCY_OR,
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
	if pushed[0].Args[0] != targetArg || pushed[1].Args[0] != targetArg ||
		pushed[2].ArgPos(0) != target.PosVec || pushed[2].Args[2] != targetArg ||
		pushed[3].Args[0] != targetArg || pushed[4].Args[0] != targetArg ||
		pushed[5].ArgF32(0) != 100 || pushed[5].Args[2] != targetArg ||
		pushed[7].ArgPos(0) != target.PosVec || pushed[7].Args[2] != targetArg {
		t.Fatalf("scheduled args = %#v", pushed)
	}
}

func TestMonsterActionFight531EC0MixedAttackUsesStrictHalfRange(t *testing.T) {
	for _, tc := range []struct {
		name     string
		distance float64
		want     ai.ActionType
	}{
		{name: "inside", distance: 49.999, want: ai.ACTION_MELEE_ATTACK},
		{name: "boundary", distance: 50, want: ai.ACTION_MISSILE_ATTACK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit := fightMonsterTestObject531E20(t)
			update := unit.UpdateDataMonster()
			update.CurrentEnemy = &Object{HealthData: &HealthData{Cur: 1, Max: 1}}
			update.MonsterDef = &MonsterDef{MeleeAttackRange112: 15, MissileAttackRange212: 100}
			update.MonsterDef.MissileName148[0] = 'x'
			var events []ai.ActionType
			var pushed []*AIStackItem
			hooks := fightHooks531EC0(120, &events, &pushed)
			hooks.distance = func(*Object, *Object) float64 { return tc.distance }
			if !monsterActionFight531EC0(unit, hooks) {
				t.Fatal("mixed fight was not handled")
			}
			found := false
			for _, action := range events {
				if action == tc.want {
					found = true
				}
			}
			if !found {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
		})
	}
}

func TestMonsterActionFight531EC0NoPhysicalAttackPolicy(t *testing.T) {
	for _, tc := range []struct {
		name   string
		caster bool
		want   []ai.ActionType
	}{
		{name: "caster waits", caster: true},
		{name: "noncaster pursues", want: []ai.ActionType{ai.DEPENDENCY_NO_NEW_ENEMY, ai.DEPENDENCY_ALIVE, ai.ACTION_MOVE_TO}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit := fightMonsterTestObject531E20(t)
			update := unit.UpdateDataMonster()
			update.CurrentEnemy = &Object{HealthData: &HealthData{Cur: 1, Max: 1}}
			update.MonsterDef = &MonsterDef{}
			if tc.caster {
				update.StatusFlags |= object.MonStatusCanCastSpells
			}
			var events []ai.ActionType
			var pushed []*AIStackItem
			if !monsterActionFight531EC0(unit, fightHooks531EC0(120, &events, &pushed)) {
				t.Fatal("fight was not handled")
			}
			if len(events) != len(tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
			for i := range tc.want {
				if events[i] != tc.want[i] {
					t.Fatalf("events = %v, want %v", events, tc.want)
				}
			}
		})
	}
}

func TestMonsterActionFight531EC0RejectsMissingNativeServiceBeforeMutation(t *testing.T) {
	unit := fightMonsterTestObject531E20(t)
	update := unit.UpdateDataMonster()
	update.CurrentEnemy = &Object{HealthData: &HealthData{Cur: 1, Max: 1}}
	before := update.AIStack[0]
	var events []ai.ActionType
	var pushed []*AIStackItem
	hooks := fightHooks531EC0(120, &events, &pushed)
	hooks.distance = nil
	if monsterActionFight531EC0(unit, hooks) {
		t.Fatal("fight with missing distance service was handled")
	}
	if update.AIStack[0] != before || len(events) != 0 || len(pushed) != 0 {
		t.Fatalf("rejected branch mutated state: %#v/%v/%v", update.AIStack[0], events, pushed)
	}
}
