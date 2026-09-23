package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func retreatToMasterMonster5456D0() (*Object, *Object) {
	owner := &Object{PosVec: types.Ptf(100, 0)}
	update := &MonsterUpdateData{AIStackInd: 0, Field329: 20, ResumeLevel: 0.5}
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_RETREAT_TO_MASTER)}
	unit := &Object{
		ObjClass:   object.ClassMonster,
		ObjOwner:   owner,
		UpdateData: unsafe.Pointer(update),
		HealthData: &HealthData{Cur: 10, Max: 100},
	}
	return unit, owner
}

func retreatToMasterHooks5456D0(events *[]ai.ActionType, pushed *[]*AIStackItem) monsterActionRetreatToMasterHooks5456D0 {
	return monsterActionRetreatToMasterHooks5456D0{
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

func TestMonsterActionRetreatToMaster5456D0SchedulesMove(t *testing.T) {
	unit, owner := retreatToMasterMonster5456D0()
	var events []ai.ActionType
	var pushed []*AIStackItem
	if !monsterActionRetreatToMaster5456D0(unit, retreatToMasterHooks5456D0(&events, &pushed)) {
		t.Fatal("valid retreat-to-master action was rejected")
	}
	want := []ai.ActionType{ai.DEPENDENCY_OBJECT_FARTHER_THAN, ai.ACTION_MOVE_TO}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if len(pushed) != 2 || pushed[0].ArgF32(0) != 20 || pushed[0].ArgU32(1) != 0 || pushed[0].ArgObj(2) != owner {
		t.Fatalf("distance dependency = %#v", pushed[0].Args)
	}
	if pushed[1].ArgPos(0) != owner.PosVec || pushed[1].ArgObj(2) != owner {
		t.Fatalf("move action = %#v, want owner at %v", pushed[1].Args, owner.PosVec)
	}
}

func TestMonsterActionRetreatToMaster5456D0StopsWhenRecoveredOrOwnerMissing(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*Object)
	}{
		{name: "recovered", setup: func(unit *Object) { unit.HealthData.Cur = 50 }},
		{name: "owner missing", setup: func(unit *Object) { unit.ObjOwner = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, _ := retreatToMasterMonster5456D0()
			tc.setup(unit)
			var events []ai.ActionType
			var pushed []*AIStackItem
			monsterActionRetreatToMaster5456D0(unit, retreatToMasterHooks5456D0(&events, &pushed))
			if want := []ai.ActionType{ai.ACTION_INVALID}; !reflect.DeepEqual(events, want) || len(pushed) != 0 {
				t.Fatalf("events = %v, pushed = %d, want one pop", events, len(pushed))
			}
		})
	}
}

func TestMonsterActionRetreatToMaster5456D0AntiMagicCasterContinues(t *testing.T) {
	unit, _ := retreatToMasterMonster5456D0()
	unit.HealthData.Cur = unit.HealthData.Max
	unit.UpdateDataMonster().StatusFlags |= object.MonStatusCanCastSpells
	unit.Buffs = uint32(1) << ENCHANT_ANTI_MAGIC
	var events []ai.ActionType
	var pushed []*AIStackItem
	monsterActionRetreatToMaster5456D0(unit, retreatToMasterHooks5456D0(&events, &pushed))
	if len(events) != 2 || events[0] != ai.DEPENDENCY_OBJECT_FARTHER_THAN || events[1] != ai.ACTION_MOVE_TO {
		t.Fatalf("anti-magic caster events = %v", events)
	}
}

func TestMonsterActionRetreatToMaster5456D0WaitsInsideRadius(t *testing.T) {
	unit, owner := retreatToMasterMonster5456D0()
	owner.PosVec = types.Ptf(50, 0) // Field329 + 30 is exactly 50; original comparison is strict.
	var events []ai.ActionType
	var pushed []*AIStackItem
	monsterActionRetreatToMaster5456D0(unit, retreatToMasterHooks5456D0(&events, &pushed))
	if len(events) != 0 || len(pushed) != 0 {
		t.Fatalf("boundary emitted events %v", events)
	}
}

func TestMonsterActionRetreatToMaster5456D0RejectsWrongAction(t *testing.T) {
	unit, _ := retreatToMasterMonster5456D0()
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_GUARD)
	var events []ai.ActionType
	var pushed []*AIStackItem
	if monsterActionRetreatToMaster5456D0(unit, retreatToMasterHooks5456D0(&events, &pushed)) || len(events) != 0 {
		t.Fatalf("wrong action was handled: %v", events)
	}
}
