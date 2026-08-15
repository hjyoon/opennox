package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestMonsterGeneratorCollide4EBE10NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantUpdateData := uintptr(748)
	wantUpdateSize := uintptr(164)
	wantScriptCollision := uintptr(72)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantUpdateData = 872
		wantUpdateSize = 216
		wantScriptCollision = 120
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"ScriptCallback size", unsafe.Sizeof(ScriptCallback{}), 8},
		{"ScriptCallback.Flags", unsafe.Offsetof(ScriptCallback{}.Flags), 0},
		{"ScriptCallback.Func", unsafe.Offsetof(ScriptCallback{}.Func), 4},
		{"MonsterGenUpdateData size", unsafe.Sizeof(MonsterGenUpdateData{}), wantUpdateSize},
		{"MonsterGenUpdateData.ScriptCollision", unsafe.Offsetof(MonsterGenUpdateData{}.ScriptCollision), wantScriptCollision},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestMonsterGeneratorCollide4EBE10NativeArgumentsEventAndIgnoredReturn(t *testing.T) {
	update := &MonsterGenUpdateData{
		ScriptCollision: ScriptCallback{Flags: 0xa5a55a5a, Func: -17},
	}
	source := &Object{UpdateData: unsafe.Pointer(update)}
	target := &Object{ObjClass: object.ClassPlayer | object.Class(0x40000000)}
	collision := &types.Pointf{
		X: math.Float32frombits(0x7fc12345),
		Y: math.Float32frombits(0x80000000),
	}
	result := new(uint32)
	calls := 0
	newUpdate := &MonsterGenUpdateData{}
	newClass := object.ClassMonster

	new(Server).MonsterGeneratorCollide4EBE10(source, target, collision,
		func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) unsafe.Pointer {
			calls++
			if block != &update.ScriptCollision || caller != target || trigger != source || event != NoxEventGeneratorCollide {
				t.Fatalf("callback args = (%p, %p, %p, %d)", block, caller, trigger, event)
			}
			if block.Flags != 0xa5a55a5a || block.Func != -17 {
				t.Fatalf("callback block = %#v", *block)
			}
			source.UpdateData = unsafe.Pointer(newUpdate)
			target.ObjClass = newClass
			return unsafe.Pointer(result)
		},
	)
	if calls != 1 {
		t.Fatalf("callback calls = %d, want 1", calls)
	}
	if update.ScriptCollision.Flags != 0xa5a55a5a || update.ScriptCollision.Func != -17 {
		t.Fatalf("entry callback block changed: %#v", update.ScriptCollision)
	}
	if source.UpdateData != unsafe.Pointer(newUpdate) || target.ObjClass != newClass {
		t.Fatal("callback mutations were not retained")
	}
	if math.Float32bits(collision.X) != 0x7fc12345 || math.Float32bits(collision.Y) != 0x80000000 {
		t.Fatalf("collision changed: %#v", *collision)
	}
}

func TestMonsterGeneratorCollide4EBE10NativeTargetGuardsDoNotReadSource(t *testing.T) {
	calls := 0
	call := func(*ScriptCallback, *Object, *Object, ScriptEventType) unsafe.Pointer {
		calls++
		return nil
	}
	srv := new(Server)
	srv.MonsterGeneratorCollide4EBE10(nil, nil, nil, call)
	srv.MonsterGeneratorCollide4EBE10(nil, &Object{ObjClass: object.ClassMonster}, &types.Pointf{}, call)
	if calls != 0 {
		t.Fatalf("callback calls = %d, want 0", calls)
	}
}

func TestMonsterGeneratorCollide4EBE10NativeNilSourceFaultsAfterPlayerGate(t *testing.T) {
	called := false
	defer func() {
		if recover() == nil {
			t.Fatal("nil source returned without panic")
		}
		if called {
			t.Fatal("nil source reached script callback")
		}
	}()
	new(Server).MonsterGeneratorCollide4EBE10(
		nil,
		&Object{ObjClass: object.ClassPlayer},
		nil,
		func(*ScriptCallback, *Object, *Object, ScriptEventType) unsafe.Pointer {
			called = true
			return nil
		},
	)
}
