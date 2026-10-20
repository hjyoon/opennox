package server

import "testing"

// Both original update routines use object+136 (a1[34]) for their cooldown. The
// separate object+536 field is Frame134, the last-damage timestamp. Keeping
// distinct values catches a native port that confuses the two field names.
func TestTriggerActivationUsesHeaderDeadline53B1B0(t *testing.T) {
	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(100)
	update := &TriggerUpdateData{Flags: 1}
	unit := triggerUpdateTestObject53B1B0(update)
	unit.Field34, unit.Frame134 = 77, 0x12345678
	callbackFrame := uint32(0)
	s.TriggerUpdate53B1B0(unit, TriggerUpdateRuntime53B1B0{
		ImmediateType: func(*Object) bool { return false },
		CollideTarget: func(*Object) *Object { return nil },
		AudioEvent:    func(uint32, *Object) {},
		ScriptCallback: func(*ScriptCallback, *Object, *Object, ScriptEventType) {
			callbackFrame = unit.Field34
		},
	})
	if callbackFrame != 100 || unit.Field34 != 130 || unit.Frame134 != 0x12345678 || update.State != 1 {
		t.Fatalf("activation callback/header/damage/state = %d/%d/%#x/%d, want 100/130/0x12345678/1",
			callbackFrame, unit.Field34, unit.Frame134, update.State)
	}
}

func TestTriggerDeactivationReadsTransferredHeaderDeadline53B1B0(t *testing.T) {
	s := new(Server)
	update := &TriggerUpdateData{Flags: 8, State: 1}
	unit := triggerUpdateTestObject53B1B0(update)
	unit.Field34, unit.Frame134 = 130, 10000
	calls := 0
	runtime := TriggerUpdateRuntime53B1B0{
		ImmediateType: func(*Object) bool { return false },
		AudioEvent:    func(uint32, *Object) {},
		ScriptCallback: func(block *ScriptCallback, caller, target *Object, event ScriptEventType) {
			if block != &update.ScriptDeactivate || caller != nil || target != unit || event != NoxEventTriggerDeactivated {
				t.Fatal("unexpected deactivation callback")
			}
			calls++
		},
	}
	s.SetFrame(130)
	s.TriggerUpdate53B1B0(unit, runtime)
	if calls != 0 || update.State != 1 {
		t.Fatal("trigger deactivated at its deadline instead of strictly after it")
	}
	s.SetFrame(131)
	s.TriggerUpdate53B1B0(unit, runtime)
	if calls != 1 || update.State != 0 || unit.Field34 != 130 || unit.Frame134 != 10000 {
		t.Fatalf("release calls/state/header/damage = %d/%d/%d/%d, want 1/0/130/10000",
			calls, update.State, unit.Field34, unit.Frame134)
	}
}

func TestToggleUsesHeaderDeadlineWithoutChangingDamageTime53B060(t *testing.T) {
	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(101)
	update := &ToggleUpdateData{Flags: 9}
	unit := toggleUpdateTestObject53B060(update)
	unit.Field34, unit.Frame134 = 100, 10000
	var events []string
	s.ToggleUpdate53B060(unit, toggleUpdateTestRuntime53B060(&events, nil))
	if len(events) != 2 || update.State != 3 || unit.Field34 != 131 || unit.Frame134 != 10000 {
		t.Fatalf("toggle events/state/header/damage = %v/%d/%d/%d, want two callbacks/3/131/10000",
			events, update.State, unit.Field34, unit.Frame134)
	}
	s.SetFrame(132)
	s.ToggleUpdate53B060(unit, toggleUpdateTestRuntime53B060(&events, nil))
	if update.State != 1 || unit.Frame134 != 10000 {
		t.Fatal("toggle release used or changed the damage timestamp")
	}
}
