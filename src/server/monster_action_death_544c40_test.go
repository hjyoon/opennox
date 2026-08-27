package server

import (
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterActionDyingStart544C40ExactOrder(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	marker := unsafe.Pointer(unit)
	update.MonsterDef = &MonsterDef{DieFunc228: marker}
	update.ScriptDeath = ScriptCallback{Flags: 0xa5, Func: 17}
	var sounds [16]uint32
	sounds[15] = 715
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	var events []string
	if !new(Server).MonsterActionDyingStart544C40(unit, MonsterActionDyingRuntime544C40{
		AudioEvent: func(id uint32, got *Object) {
			if id != 715 || got != unit {
				t.Fatalf("death audio = %d/%p", id, got)
			}
			events = append(events, "audio")
		},
		ScriptCallback: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) {
			if block != &update.ScriptDeath || caller != nil || trigger != unit || event != NoxEventMonsterDead {
				t.Fatalf("death callback = %p/%p/%p/%d", block, caller, trigger, event)
			}
			events = append(events, "script")
		},
		CanDieFunc: func(got unsafe.Pointer) bool { return got == marker },
		DieFunc: func(got unsafe.Pointer, obj *Object) {
			if got != marker || obj != unit {
				t.Fatalf("die function = %p/%p", got, obj)
			}
			events = append(events, "die")
		},
	}) {
		t.Fatal("native dying start was not handled")
	}
	want := []string{"audio", "script", "die"}
	if len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestMonsterActionDyingUpdate544D60PopsCompletedAnimation(t *testing.T) {
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	unit := monsterActionTestObject50A910(t)
	unit.serverHandle = s.handle
	update := unit.UpdateDataMonster()
	update.AIStackInd = 1
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_DEAD)}
	update.AIStack[1] = AIStackItem{Action: uint32(ai.ACTION_DYING), Field5: 1}
	update.Field120_3 = 1
	if !s.MonsterActionDyingUpdate544D60(unit) {
		t.Fatal("native dying update was not handled")
	}
	if update.AIStackInd != 0 || update.AIStackHead().Type() != ai.ACTION_DEAD {
		t.Fatalf("action stack after dying = %#v", update.GetAIStack())
	}
}

func TestMonsterActionDead544D80OrdinaryMonsterLifecycle(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	deadFunc := unsafe.Pointer(unit)
	update.MonsterDef = &MonsterDef{DeadFunc232: deadFunc}
	update.Field74 = 2
	update.Waypoints[0] = &Waypoint{}
	update.Waypoints[1] = &Waypoint{}
	update.Field91 = 9
	update.Field282_1 = 2
	update.SeenEnemies[0] = &Object{}
	update.SeenEnemies[1] = &Object{}
	update.CurrentEnemy = &Object{}
	update.Field523_2 = 3
	unit.VelVec = types.Ptf(1, 2)
	unit.ForceVec = types.Ptf(3, 4)
	unit.Pos24 = types.Ptf(5, 6)
	unit.Update = unsafe.Pointer(unit)
	var events []string
	runtime := MonsterActionDeadRuntime544D80{
		IsZombie:    func(*Object) bool { return false },
		CanDeadFunc: func(got unsafe.Pointer) bool { return got == deadFunc },
		DeadFunc: func(got unsafe.Pointer, obj *Object) {
			if got != deadFunc || obj != unit || unit.VelVec != (types.Pointf{}) ||
				unit.ForceVec != (types.Pointf{}) || unit.Pos24 != (types.Pointf{}) {
				t.Fatalf("dead callback state = %p/%p/%v/%v/%v", got, obj, unit.VelVec, unit.ForceVec, unit.Pos24)
			}
			events = append(events, "dead")
		},
		RemoveUpdatable: func(obj *Object) {
			wantFlags := object.FlagAllowOverlap | object.FlagShort
			if obj != unit || unit.ObjFlags&wantFlags != wantFlags {
				t.Fatalf("remove-updatable state = %p/%#x", obj, unit.ObjFlags)
			}
			events = append(events, "remove")
		},
	}
	if !new(Server).MonsterActionDeadStart544D80(unit, runtime) {
		t.Fatal("native dead start was not handled")
	}
	wantFlags := object.FlagAllowOverlap | object.FlagShort
	if unit.ObjFlags&wantFlags != wantFlags {
		t.Fatalf("dead start flags = %#x", unit.ObjFlags)
	}
	if !new(Server).MonsterActionDeadUpdate544EC0(unit, runtime) {
		t.Fatal("native dead update was not handled")
	}
	if len(events) != 2 || events[0] != "dead" || events[1] != "remove" {
		t.Fatalf("events = %v, want [dead remove]", events)
	}
	if unit.Update != nil || update.Field523_2 != 0 || update.Field74 != 0 || update.Field91 != 0 ||
		update.Field282_1 != 0 || update.CurrentEnemy != nil {
		t.Fatalf("dead cleanup scalar state = update:%p %#v", unit.Update, update)
	}
	for i, waypoint := range update.Waypoints {
		if waypoint != nil {
			t.Fatalf("waypoint %d survived dead cleanup", i)
		}
	}
	for i, enemy := range update.SeenEnemies {
		if enemy != nil {
			t.Fatalf("seen enemy %d survived dead cleanup", i)
		}
	}
}

func TestMonsterActionDeadStart544D80PreflightsUnsupportedCallback(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.MonsterDef = &MonsterDef{DeadFunc232: unsafe.Pointer(unit)}
	unit.VelVec = types.Ptf(1, 2)
	before := *unit
	var reason string
	if new(Server).MonsterActionDeadStart544D80(unit, MonsterActionDeadRuntime544D80{
		IsZombie: func(*Object) bool { return false },
		Unsupported: func(got string, _ *Object) {
			reason = got
		},
	}) {
		t.Fatal("unsupported dead callback was accepted")
	}
	if reason != "monster dead function" || *unit != before {
		t.Fatalf("preflight result = %q, mutated=%v", reason, *unit != before)
	}
}
