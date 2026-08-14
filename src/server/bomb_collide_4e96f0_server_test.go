package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestBombCollide4E96F0NativeLayout(t *testing.T) {
	wantObjectClass := uintptr(8)
	wantObjectFlags := uintptr(16)
	wantObjectTeam := uintptr(48)
	wantObjectOwner := uintptr(508)
	wantObjectUpdate := uintptr(748)
	wantMonsterSize := uintptr(2200)
	wantScriptCollision := uintptr(1272)
	wantBombTarget := uintptr(2176)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectClass = 12
		wantObjectFlags = 20
		wantObjectTeam = 52
		wantObjectOwner = 552
		wantObjectUpdate = 872
		wantMonsterSize = 2824
		wantScriptCollision = 1880
		wantBombTarget = 2784
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"BombCollideData size", unsafe.Sizeof(BombCollideData{}), 8},
		{"BombCollideData.Reserved", unsafe.Offsetof(BombCollideData{}.Reserved), 0},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantObjectFlags},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), wantObjectTeam},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantObjectOwner},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"MonsterUpdateData size", unsafe.Sizeof(MonsterUpdateData{}), wantMonsterSize},
		{"MonsterUpdateData.ScriptCollision", unsafe.Offsetof(MonsterUpdateData{}.ScriptCollision), wantScriptCollision},
		{"MonsterUpdateData.BombCollideTarget", unsafe.Offsetof(MonsterUpdateData{}.BombCollideTarget), wantBombTarget},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestBombCollideNative4E96F0ArgumentsEventCachedUpdateAndStoreOrder(t *testing.T) {
	update := &MonsterUpdateData{ScriptCollision: ScriptCallback{Flags: 0xa55a5aa5, Func: -17}}
	replacement := &MonsterUpdateData{}
	bomb := &Object{UpdateData: unsafe.Pointer(update)} // Deliberately not ClassMonster.
	other := &Object{ObjClass: object.ClassPlayer}
	collision := new(uint32)
	var events []string
	bombCollideNative4E96F0(bomb, other, unsafe.Pointer(collision), bombCollideNativeDeps4E96F0{
		gameModeCoop: func() int32 {
			events = append(events, "mode")
			return 0
		},
		firstPlayerUnit: func() *Object {
			t.Fatal("non-Coop path read first player")
			return nil
		},
		scriptCallback: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) unsafe.Pointer {
			events = append(events, "script")
			if block != &update.ScriptCollision || caller != other || trigger != bomb || event != NoxEventBombCollide {
				t.Fatalf("script args = (%p, %p, %p, %d)", block, caller, trigger, event)
			}
			if block.Flags != 0xa55a5aa5 || block.Func != -17 {
				t.Fatalf("script block = %#v", block)
			}
			bomb.UpdateData = unsafe.Pointer(replacement)
			return unsafe.Pointer(collision)
		},
		damageClear: func(obj *Object, damage int32) {
			events = append(events, "damage")
			if obj != bomb || damage != 999 {
				t.Fatalf("damage args = (%p, %d), want (%p, 999)", obj, damage, bomb)
			}
			if update.BombCollideTarget != other || replacement.BombCollideTarget != nil {
				t.Fatalf("target at damage = (%p, %p), want (%p, nil)", update.BombCollideTarget, replacement.BombCollideTarget, other)
			}
		},
	})
	if update.BombCollideTarget != other || replacement.BombCollideTarget != nil {
		t.Fatalf("final target = (%p, %p), want (%p, nil)", update.BombCollideTarget, replacement.BombCollideTarget, other)
	}
	if len(events) != 3 || events[0] != "mode" || events[1] != "script" || events[2] != "damage" {
		t.Fatalf("events = %#v, want [mode script damage]", events)
	}
}

func TestBombCollideNative4E96F0NilBombFaultsBeforeMode(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil bomb returned without a panic")
		}
	}()
	bombCollideNative4E96F0(nil, nil, nil, bombCollideNativeDeps4E96F0{
		gameModeCoop: func() int32 {
			t.Fatal("nil bomb reached mode check")
			return 0
		},
	})
}

func TestUnitsHaveSameTeamNative4EC520OwnerChains(t *testing.T) {
	teamOwner := &Object{TeamVal: ObjectTeam{ID: 6}}
	first := &Object{ObjOwner: teamOwner}
	secondOwner := &Object{TeamVal: ObjectTeam{ID: 6}}
	second := &Object{ObjOwner: secondOwner}
	if !UnitsHaveSameTeam4EC520(first, second) {
		t.Fatal("matching owner teams returned false")
	}
	secondOwner.TeamVal.ID = 7
	if UnitsHaveSameTeam4EC520(first, second) {
		t.Fatal("distinct owner teams returned true")
	}
	shared := &Object{}
	first.ObjOwner = shared
	second.ObjOwner = shared
	if !UnitsHaveSameTeam4EC520(first, second) {
		t.Fatal("shared owner identity returned false")
	}
	if UnitsHaveSameTeam4EC520(nil, second) || UnitsHaveSameTeam4EC520(first, nil) {
		t.Fatal("nil input returned true")
	}
}
