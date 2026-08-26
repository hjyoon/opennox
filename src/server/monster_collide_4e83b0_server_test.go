package server

import (
	"testing"
	"unsafe"
)

func TestMonsterCollide4E83B0NativeLayout(t *testing.T) {
	wantCollision := uintptr(1272)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantCollision = 2012
	}
	if got := unsafe.Sizeof(ScriptCallback{}); got != 8 {
		t.Fatalf("ScriptCallback size = %d, want 8", got)
	}
	if got := unsafe.Offsetof(ScriptCallback{}.Flags); got != 0 {
		t.Fatalf("ScriptCallback.Flags offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(ScriptCallback{}.Func); got != 4 {
		t.Fatalf("ScriptCallback.Func offset = %d, want 4", got)
	}
	if got := unsafe.Offsetof(MonsterUpdateData{}.ScriptCollision); got != wantCollision {
		t.Fatalf("MonsterUpdateData.ScriptCollision offset = %d, want %d", got, wantCollision)
	}
}

func TestMonsterCollide4E83B0NativeArgumentsEventAndExactReturn(t *testing.T) {
	update := &MonsterUpdateData{}
	update.ScriptCollision = ScriptCallback{Flags: 0xa5a55a5a, Func: -17}
	monster := &Object{UpdateData: unsafe.Pointer(update)} // Deliberately not ClassMonster.
	other := &Object{}
	result := new(uint32)
	calls := 0
	got := MonsterCollideScript4E83B0(monster, other,
		func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) unsafe.Pointer {
			calls++
			if block != &update.ScriptCollision || caller != other || trigger != monster || event != NoxEventMonsterCollide {
				t.Fatalf("callback args = (%p, %p, %p, %d)", block, caller, trigger, event)
			}
			if block.Flags != 0xa5a55a5a || block.Func != -17 {
				t.Fatalf("callback block changed: %#v", block)
			}
			monster.UpdateData = nil
			return unsafe.Pointer(result)
		},
	)
	if got != unsafe.Pointer(result) || calls != 1 {
		t.Fatalf("result/calls = (%p, %d), want (%p, 1)", got, calls, result)
	}
	if update.ScriptCollision.Flags != 0xa5a55a5a || update.ScriptCollision.Func != -17 {
		t.Fatalf("collision block changed: %#v", update.ScriptCollision)
	}
}

func TestMonsterCollide4E83B0NativePassesNilOther(t *testing.T) {
	update := &MonsterUpdateData{}
	monster := &Object{UpdateData: unsafe.Pointer(update)}
	calls := 0
	got := MonsterCollideScript4E83B0(monster, nil,
		func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) unsafe.Pointer {
			calls++
			if block != &update.ScriptCollision || caller != nil || trigger != monster || event != NoxEventMonsterCollide {
				t.Fatalf("callback args = (%p, %p, %p, %d)", block, caller, trigger, event)
			}
			return nil
		},
	)
	if got != nil || calls != 1 {
		t.Fatalf("result/calls = (%p, %d), want (nil, 1)", got, calls)
	}
}

func TestMonsterCollide4E83B0NativeNilUpdateFaultsBeforeCallback(t *testing.T) {
	monster := &Object{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil update returned without a panic")
		}
	}()
	MonsterCollideScript4E83B0(monster, nil,
		func(*ScriptCallback, *Object, *Object, ScriptEventType) unsafe.Pointer {
			t.Fatal("nil update reached callback")
			return nil
		},
	)
}
