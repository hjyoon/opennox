package server

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func triggerUpdateTestObject53B1B0(update *TriggerUpdateData) *Object {
	return &Object{
		ObjClass:   object.ClassTrigger | object.ClassClientPersist,
		ObjFlags:   object.FlagEnabled,
		UpdateData: unsafe.Pointer(update),
	}
}

func TestTriggerUpdate53B1B0ActivationAndDeactivation(t *testing.T) {
	server := new(Server)
	server.SetTickRate(30)
	server.SetFrame(100)
	target := &Object{}
	update := &TriggerUpdateData{
		Flags:           0x9,
		SoundActivate:   101,
		SoundDeactivate: 202,
	}
	unit := triggerUpdateTestObject53B1B0(update)
	var events []string
	runtime := TriggerUpdateRuntime53B1B0{
		ImmediateType: func(*Object) bool { return false },
		CollideTarget: func(*Object) *Object { return target },
		AudioEvent: func(id uint32, got *Object) {
			events = append(events, fmt.Sprintf("audio:%d:%t", id, got == unit))
		},
		ScriptCallback: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) {
			events = append(events, fmt.Sprintf("script:%d:%t:%t", event, caller == target, trigger == unit))
		},
	}

	if got := server.TriggerUpdate53B1B0(unit, runtime); got != 0xc {
		t.Fatalf("activation result = %#x, want 0xc", got)
	}
	if update.State != 1 || unit.Frame134 != 130 || unit.Field33 != 1 {
		t.Fatalf("activation state = (%d,%d,%d), want (1,130,1)", update.State, unit.Frame134, unit.Field33)
	}
	wantActivate := []string{"audio:101:true", "script:8:true:true"}
	if fmt.Sprint(events) != fmt.Sprint(wantActivate) {
		t.Fatalf("activation events = %v, want %v", events, wantActivate)
	}

	server.SetFrame(131)
	events = nil
	if got := server.TriggerUpdate53B1B0(unit, runtime); got != 0x8 {
		t.Fatalf("deactivation result = %#x, want 0x8", got)
	}
	if update.State != 0 || unit.Field33 != 0 {
		t.Fatalf("deactivation state = (%d,%d), want (0,0)", update.State, unit.Field33)
	}
	wantDeactivate := []string{"audio:202:true", "script:9:false:true"}
	if fmt.Sprint(events) != fmt.Sprint(wantDeactivate) {
		t.Fatalf("deactivation events = %v, want %v", events, wantDeactivate)
	}
}

func TestTriggerUpdate53B1B0ImmediateAndDisabledPaths(t *testing.T) {
	t.Run("immediate type", func(t *testing.T) {
		server := new(Server)
		server.SetTickRate(30)
		server.SetFrame(77)
		update := &TriggerUpdateData{Flags: 0x9}
		unit := triggerUpdateTestObject53B1B0(update)
		server.TriggerUpdate53B1B0(unit, TriggerUpdateRuntime53B1B0{
			ImmediateType:  func(*Object) bool { return true },
			CollideTarget:  func(*Object) *Object { return nil },
			AudioEvent:     func(uint32, *Object) {},
			ScriptCallback: func(*ScriptCallback, *Object, *Object, ScriptEventType) {},
		})
		if unit.Frame134 != 77 {
			t.Fatalf("immediate deadline = %d, want 77", unit.Frame134)
		}
	})
	t.Run("disabled", func(t *testing.T) {
		update := &TriggerUpdateData{Flags: 0xffffffff, State: 7}
		unit := triggerUpdateTestObject53B1B0(update)
		unit.ObjFlags &^= object.FlagEnabled
		got := new(Server).TriggerUpdate53B1B0(unit, TriggerUpdateRuntime53B1B0{
			ImmediateType: func(*Object) bool { return false },
		})
		if got != 0xf6 || update.Flags != 0xfffffff6 || update.State != 7 {
			t.Fatalf("disabled result = (%#x,%#x,%d), want (0xf6,0xfffffff6,7)", got, update.Flags, update.State)
		}
	})
}
