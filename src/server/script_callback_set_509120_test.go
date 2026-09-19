package server

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

const scriptCallbackNameTableSize509120 = 19 * 128

type scriptCallbackFixture509120 struct {
	object    Object
	names     []byte
	trigger   TriggerUpdateData
	monster   MonsterUpdateData
	hole      HoleCollideData
	generator MonsterGenUpdateData
}

func newScriptCallbackFixture509120(class object.Class) *scriptCallbackFixture509120 {
	f := &scriptCallbackFixture509120{names: bytes.Repeat([]byte{0xa5}, scriptCallbackNameTableSize509120)}
	f.object.ObjClass = class
	f.object.Field189 = unsafe.Pointer(&f.names[0])
	f.object.ScriptPickup = ScriptCallback{Flags: 0x10000001, Func: -101}
	f.trigger.ScriptCollide = ScriptCallback{Flags: 0x20000001, Func: -201}
	f.trigger.ScriptActivate = ScriptCallback{Flags: 0x20000002, Func: -202}
	f.trigger.ScriptDeactivate = ScriptCallback{Flags: 0x20000003, Func: -203}
	f.monster.ScriptLookingForEnemy = ScriptCallback{Flags: 0x30000001, Func: -301}
	f.monster.ScriptEnemySighted = ScriptCallback{Flags: 0x30000002, Func: -302}
	f.monster.ScriptChangeFocus = ScriptCallback{Flags: 0x30000003, Func: -303}
	f.monster.ScriptIsHit = ScriptCallback{Flags: 0x30000004, Func: -304}
	f.monster.ScriptRetreat = ScriptCallback{Flags: 0x30000005, Func: -305}
	f.monster.ScriptDeath = ScriptCallback{Flags: 0x30000006, Func: -306}
	f.monster.ScriptCollision = ScriptCallback{Flags: 0x30000007, Func: -307}
	f.monster.ScriptHearEnemy = ScriptCallback{Flags: 0x30000008, Func: -308}
	f.monster.ScriptEndOfWaypoint = ScriptCallback{Flags: 0x30000009, Func: -309}
	f.monster.ScriptLostEnemy = ScriptCallback{Flags: 0x3000000a, Func: -310}
	f.hole.Script = ScriptCallback{Flags: 0x40000001, Func: -401}
	f.generator.Field48 = 0x51000001
	f.generator.FuncInd52 = 0x52000002
	f.generator.Field56 = 0x53000003
	f.generator.FuncInd60 = 0x54000004
	f.generator.Field64 = 0x55000005
	f.generator.FuncInd68 = 0x56000006
	f.generator.ScriptCollision = ScriptCallback{Flags: 0x57000007, Func: -507}

	switch {
	case class.Has(object.ClassTrigger):
		f.object.UpdateData = unsafe.Pointer(&f.trigger)
	case class.Has(object.ClassMonster):
		f.object.UpdateData = unsafe.Pointer(&f.monster)
	case class.Has(object.ClassHole):
		f.object.CollideData = unsafe.Pointer(&f.hole)
	case class.Has(object.ClassMonsterGenerator):
		f.object.UpdateData = unsafe.Pointer(&f.generator)
	}
	return f
}

type scriptCallbackState509120 struct {
	flags    uint32
	function uint32
}

func (f *scriptCallbackFixture509120) state509120() map[string]scriptCallbackState509120 {
	state := make(map[string]scriptCallbackState509120)
	add := func(name string, callback *ScriptCallback) {
		state[name] = scriptCallbackState509120{flags: callback.Flags, function: uint32(callback.Func)}
	}
	add("pickup", &f.object.ScriptPickup)
	add("trigger-collide", &f.trigger.ScriptCollide)
	add("trigger-activate", &f.trigger.ScriptActivate)
	add("trigger-deactivate", &f.trigger.ScriptDeactivate)
	add("monster-looking", &f.monster.ScriptLookingForEnemy)
	add("monster-sighted", &f.monster.ScriptEnemySighted)
	add("monster-focus", &f.monster.ScriptChangeFocus)
	add("monster-hit", &f.monster.ScriptIsHit)
	add("monster-retreat", &f.monster.ScriptRetreat)
	add("monster-death", &f.monster.ScriptDeath)
	add("monster-collision", &f.monster.ScriptCollision)
	add("monster-heard", &f.monster.ScriptHearEnemy)
	add("monster-waypoint", &f.monster.ScriptEndOfWaypoint)
	add("monster-lost", &f.monster.ScriptLostEnemy)
	add("hole", &f.hole.Script)
	add("generator-collision", &f.generator.ScriptCollision)
	state["generator-52"] = scriptCallbackState509120{function: f.generator.FuncInd52}
	state["generator-60"] = scriptCallbackState509120{function: f.generator.FuncInd60}
	state["generator-68"] = scriptCallbackState509120{function: f.generator.FuncInd68}
	state["generator-adjacent-48"] = scriptCallbackState509120{function: f.generator.Field48}
	state["generator-adjacent-56"] = scriptCallbackState509120{function: f.generator.Field56}
	state["generator-adjacent-64"] = scriptCallbackState509120{function: f.generator.Field64}
	return state
}

func scriptCallbackCases509120() []struct {
	name   string
	class  object.Class
	event  int
	offset int
	target string
} {
	return []struct {
		name   string
		class  object.Class
		event  int
		offset int
		target string
	}{
		{name: "pickup", event: 14, offset: 0, target: "pickup"},
		{name: "trigger_collision", class: object.ClassTrigger, event: 0, offset: 512, target: "trigger-collide"},
		{name: "trigger_activate", class: object.ClassTrigger, event: 1, offset: 256, target: "trigger-activate"},
		{name: "trigger_deactivate", class: object.ClassTrigger, event: 2, offset: 384, target: "trigger-deactivate"},
		{name: "monster_enemy_sighted", class: object.ClassMonster, event: 3, offset: 640, target: "monster-sighted"},
		{name: "monster_looking", class: object.ClassMonster, event: 4, offset: 768, target: "monster-looking"},
		{name: "monster_death", class: object.ClassMonster, event: 5, offset: 896, target: "monster-death"},
		{name: "monster_focus", class: object.ClassMonster, event: 6, offset: 1024, target: "monster-focus"},
		{name: "monster_hit", class: object.ClassMonster, event: 7, offset: 1152, target: "monster-hit"},
		{name: "monster_retreat", class: object.ClassMonster, event: 8, offset: 1280, target: "monster-retreat"},
		{name: "monster_collision", class: object.ClassMonster, event: 9, offset: 1408, target: "monster-collision"},
		{name: "monster_heard", class: object.ClassMonster, event: 10, offset: 1536, target: "monster-heard"},
		{name: "monster_waypoint", class: object.ClassMonster, event: 11, offset: 1664, target: "monster-waypoint"},
		{name: "hole", class: object.ClassHole, event: 12, offset: 128, target: "hole"},
		{name: "monster_lost", class: object.ClassMonster, event: 13, offset: 1792, target: "monster-lost"},
		{name: "generator_52", class: object.ClassMonsterGenerator, event: 15, offset: 1920, target: "generator-52"},
		{name: "generator_60", class: object.ClassMonsterGenerator, event: 16, offset: 2048, target: "generator-60"},
		{name: "generator_68", class: object.ClassMonsterGenerator, event: 17, offset: 2304, target: "generator-68"},
		{name: "generator_collision", class: object.ClassMonsterGenerator, event: 18, offset: 2176, target: "generator-collision"},
	}
}

func TestScriptCallbackSet509120MapsEveryFunctionSlot(t *testing.T) {
	const wantFunction = int32(-0x1234567)
	for _, tc := range scriptCallbackCases509120() {
		t.Run(tc.name, func(t *testing.T) {
			f := newScriptCallbackFixture509120(tc.class)
			before := f.state509120()
			beforeNames := append([]byte(nil), f.names...)
			calls := 0

			if ok := scriptCallbackSet509120(&f.object, tc.event, "OnEvent", false, func(name string) int32 {
				calls++
				if name != "OnEvent" {
					t.Fatalf("index name = %q, want OnEvent", name)
				}
				return wantFunction
			}); !ok {
				t.Fatal("valid event was rejected")
			}
			if calls != 1 {
				t.Fatalf("index calls = %d, want 1", calls)
			}
			wantState := make(map[string]scriptCallbackState509120, len(before))
			for key, value := range before {
				wantState[key] = value
			}
			target := wantState[tc.target]
			function := wantFunction
			target.function = uint32(function)
			wantState[tc.target] = target
			if got := f.state509120(); !reflect.DeepEqual(got, wantState) {
				t.Fatalf("callback state = %#v, want %#v", got, wantState)
			}
			if !bytes.Equal(f.names, beforeNames) {
				t.Fatal("compiled callback mode changed the name table")
			}
		})
	}
}

func TestScriptCallbackSet509120MapsEveryNameSlot(t *testing.T) {
	const name = "OnEvent"
	for _, tc := range scriptCallbackCases509120() {
		t.Run(tc.name, func(t *testing.T) {
			f := newScriptCallbackFixture509120(tc.class)
			beforeState := f.state509120()
			wantNames := append([]byte(nil), f.names...)
			copy(wantNames[tc.offset:], name)
			wantNames[tc.offset+len(name)] = 0

			if ok := scriptCallbackSet509120(&f.object, tc.event, name, true, func(string) int32 {
				t.Fatal("name-table mode called ScriptIndexByName")
				return 0
			}); !ok {
				t.Fatal("valid event was rejected")
			}
			if !reflect.DeepEqual(f.state509120(), beforeState) {
				t.Fatal("name-table mode changed a callback function or flags word")
			}
			if !bytes.Equal(f.names, wantNames) {
				for i := range f.names {
					if f.names[i] != wantNames[i] {
						t.Fatalf("name table byte %d = %#x, want %#x", i, f.names[i], wantNames[i])
					}
				}
				t.Fatal("name table mismatch")
			}
		})
	}
}

func TestScriptCallbackSet509120NameModeDoesNotRequireCallbackData(t *testing.T) {
	for _, tc := range []struct {
		name   string
		class  object.Class
		event  int
		offset int
	}{
		{name: "trigger", class: object.ClassTrigger, event: 1, offset: 256},
		{name: "monster", class: object.ClassMonster, event: 3, offset: 640},
		{name: "hole", class: object.ClassHole, event: 12, offset: 128},
		{name: "generator", class: object.ClassMonsterGenerator, event: 15, offset: 1920},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newScriptCallbackFixture509120(tc.class)
			f.object.UpdateData = nil
			f.object.CollideData = nil

			if ok := scriptCallbackSet509120(&f.object, tc.event, "Deferred", true, func(string) int32 {
				t.Fatal("name-table mode called ScriptIndexByName")
				return 0
			}); !ok {
				t.Fatal("valid event was rejected without callback data")
			}
			if got := f.names[tc.offset : tc.offset+len("Deferred")+1]; !bytes.Equal(got, []byte("Deferred\x00")) {
				t.Fatalf("stored name = %q, want %q", got, "Deferred\\x00")
			}
		})
	}
}

func TestScriptCallbackSet509120RejectsBeforeMutation(t *testing.T) {
	for _, class := range []object.Class{
		0,
		object.ClassTrigger,
		object.ClassMonster,
		object.ClassHole,
		object.ClassMonsterGenerator,
	} {
		t.Run(fmt.Sprintf("class_%#x", uint32(class)), func(t *testing.T) {
			f := newScriptCallbackFixture509120(class)
			beforeState := f.state509120()
			beforeNames := append([]byte(nil), f.names...)
			if ok := scriptCallbackSet509120(&f.object, 19, "Invalid", false, func(string) int32 {
				t.Fatal("invalid event called ScriptIndexByName")
				return 0
			}); ok {
				t.Fatal("invalid event was accepted")
			}
			if !reflect.DeepEqual(f.state509120(), beforeState) || !bytes.Equal(f.names, beforeNames) {
				t.Fatal("invalid event changed state")
			}
		})
	}

	f := newScriptCallbackFixture509120(object.ClassTrigger)
	f.object.Field189 = nil
	beforeState := f.state509120()
	if ok := scriptCallbackSet509120(&f.object, 1, "Ignored", false, func(string) int32 {
		t.Fatal("nil Field189 called ScriptIndexByName")
		return 0
	}); ok || !reflect.DeepEqual(f.state509120(), beforeState) {
		t.Fatal("nil Field189 did not stop before callback mutation")
	}
}

func TestScriptCallbackSet509120PreservesClassPriority(t *testing.T) {
	f := newScriptCallbackFixture509120(object.ClassTrigger | object.ClassMonster)
	f.object.UpdateData = unsafe.Pointer(&f.trigger)
	before := f.state509120()
	if ok := scriptCallbackSet509120(&f.object, 3, "MonsterOnly", false, func(string) int32 {
		t.Fatal("lower-priority Monster event reached index lookup")
		return 0
	}); ok || !reflect.DeepEqual(f.state509120(), before) {
		t.Fatal("Trigger|Monster object did not stop in the Trigger branch")
	}

	if ok := scriptCallbackSet509120(&f.object, 0, "Trigger", false, func(string) int32 { return 77 }); !ok {
		t.Fatal("Trigger event was rejected")
	}
	if f.trigger.ScriptCollide.Func != 77 || f.monster.ScriptEnemySighted.Func != -302 {
		t.Fatalf("priority callbacks = trigger %d / monster %d, want 77 / -302",
			f.trigger.ScriptCollide.Func, f.monster.ScriptEnemySighted.Func)
	}
}

func TestScriptCallbackSet509120CachesCallbackDataBeforeIndexLookup(t *testing.T) {
	t.Run("trigger", func(t *testing.T) {
		f := newScriptCallbackFixture509120(object.ClassTrigger)
		var replacement TriggerUpdateData
		replacement.ScriptActivate.Func = -1
		scriptCallbackSet509120(&f.object, 1, "Activate", false, func(string) int32 {
			f.object.UpdateData = unsafe.Pointer(&replacement)
			return 71
		})
		if f.trigger.ScriptActivate.Func != 71 || replacement.ScriptActivate.Func != -1 {
			t.Fatalf("callbacks = original %d / replacement %d, want 71 / -1",
				f.trigger.ScriptActivate.Func, replacement.ScriptActivate.Func)
		}
	})

	t.Run("monster", func(t *testing.T) {
		f := newScriptCallbackFixture509120(object.ClassMonster)
		var replacement MonsterUpdateData
		replacement.ScriptEnemySighted.Func = -1
		scriptCallbackSet509120(&f.object, 3, "Sighted", false, func(string) int32 {
			f.object.UpdateData = unsafe.Pointer(&replacement)
			return 72
		})
		if f.monster.ScriptEnemySighted.Func != 72 || replacement.ScriptEnemySighted.Func != -1 {
			t.Fatalf("callbacks = original %d / replacement %d, want 72 / -1",
				f.monster.ScriptEnemySighted.Func, replacement.ScriptEnemySighted.Func)
		}
	})

	t.Run("hole", func(t *testing.T) {
		f := newScriptCallbackFixture509120(object.ClassHole)
		var replacement HoleCollideData
		replacement.Script.Func = -1
		scriptCallbackSet509120(&f.object, 12, "Hole", false, func(string) int32 {
			f.object.CollideData = unsafe.Pointer(&replacement)
			return 73
		})
		if f.hole.Script.Func != 73 || replacement.Script.Func != -1 {
			t.Fatalf("callbacks = original %d / replacement %d, want 73 / -1",
				f.hole.Script.Func, replacement.Script.Func)
		}
	})

	t.Run("generator", func(t *testing.T) {
		f := newScriptCallbackFixture509120(object.ClassMonsterGenerator)
		var replacement MonsterGenUpdateData
		replacement.FuncInd52 = ^uint32(0)
		scriptCallbackSet509120(&f.object, 15, "Generate", false, func(string) int32 {
			f.object.UpdateData = unsafe.Pointer(&replacement)
			return 74
		})
		if f.generator.FuncInd52 != 74 || replacement.FuncInd52 != ^uint32(0) {
			t.Fatalf("callbacks = original %d / replacement %d, want 74 / %d",
				f.generator.FuncInd52, replacement.FuncInd52, ^uint32(0))
		}
	})
}
