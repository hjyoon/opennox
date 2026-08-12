package opennox

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerObserverFindGoodSlave4E6280StartsAfterCamera(t *testing.T) {
	camera := &server.Object{}
	good := &server.Object{}
	pl := &server.Player{CameraFollowObj: camera, PlayerUnit: &server.Object{}}
	var calls []string
	got := playerObserverFindGoodSlave0_4E6280(
		pl,
		func(*server.Object) *server.Object {
			t.Fatal("read first slave while camera was set")
			return nil
		},
		func(obj *server.Object) *server.Object {
			calls = append(calls, "next")
			if obj != camera {
				t.Fatalf("next argument = %p, want camera %p", obj, camera)
			}
			return good
		},
	)
	if got != good {
		t.Fatalf("result = %p, want good slave %p", got, good)
	}
	if want := []string{"next"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerObserverFindGoodSlave4E6280StartsAtPlayerUnit(t *testing.T) {
	unit := &server.Object{}
	good := &server.Object{}
	pl := &server.Player{PlayerUnit: unit}
	got := playerObserverFindGoodSlave0_4E6280(
		pl,
		func(obj *server.Object) *server.Object {
			if obj != unit {
				t.Fatalf("first argument = %p, want PlayerUnit %p", obj, unit)
			}
			return good
		},
		func(*server.Object) *server.Object {
			t.Fatal("advanced past an unflagged slave")
			return nil
		},
	)
	if got != good {
		t.Fatalf("result = %p, want good slave %p", got, good)
	}
}

func TestPlayerObserverFindGoodSlave4E6280SkipsRejectedFlags(t *testing.T) {
	unit := &server.Object{}
	dead := &server.Object{ObjFlags: object.Flags(0x20)}
	flagged := &server.Object{ObjFlags: object.Flags(0x8000)}
	good := &server.Object{ObjFlags: object.Flags(0x4000)}
	next := map[*server.Object]*server.Object{dead: flagged, flagged: good}
	var calls []*server.Object
	got := playerObserverFindGoodSlave0_4E6280(
		&server.Player{PlayerUnit: unit},
		func(*server.Object) *server.Object { return dead },
		func(obj *server.Object) *server.Object {
			calls = append(calls, obj)
			return next[obj]
		},
	)
	if got != good {
		t.Fatalf("result = %p, want good slave %p", got, good)
	}
	if want := []*server.Object{dead, flagged}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("next arguments = %v, want %v", calls, want)
	}
}

func TestPlayerObserverFindGoodSlave4E6280ReloadsPlayerUnitForFallback(t *testing.T) {
	oldUnit := &server.Object{}
	newUnit := &server.Object{}
	rejected := &server.Object{ObjFlags: object.Flags(0x8020)}
	good := &server.Object{}
	pl := &server.Player{PlayerUnit: oldUnit}
	var calls []string
	got := playerObserverFindGoodSlave0_4E6280(
		pl,
		func(unit *server.Object) *server.Object {
			switch unit {
			case oldUnit:
				calls = append(calls, "first old")
				return rejected
			case newUnit:
				calls = append(calls, "first new")
				return good
			default:
				t.Fatalf("unexpected PlayerUnit %p", unit)
				return nil
			}
		},
		func(obj *server.Object) *server.Object {
			if obj != rejected {
				t.Fatalf("next argument = %p, want rejected slave %p", obj, rejected)
			}
			calls = append(calls, "next rejected")
			pl.PlayerUnit = newUnit
			return nil
		},
	)
	if got != good {
		t.Fatalf("result = %p, want fallback slave %p", got, good)
	}
	if want := []string{"first old", "next rejected", "first new"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerObserverFindGoodSlave4E6280RetriesEmptyFirstPass(t *testing.T) {
	unit := &server.Object{}
	good := &server.Object{}
	calls := 0
	got := playerObserverFindGoodSlave0_4E6280(
		&server.Player{PlayerUnit: unit},
		func(obj *server.Object) *server.Object {
			if obj != unit {
				t.Fatalf("first argument = %p, want PlayerUnit %p", obj, unit)
			}
			calls++
			if calls == 1 {
				return nil
			}
			return good
		},
		func(*server.Object) *server.Object {
			t.Fatal("called next for nil or unflagged candidate")
			return nil
		},
	)
	if got != good {
		t.Fatalf("result = %p, want second-pass slave %p", got, good)
	}
	if calls != 2 {
		t.Fatalf("first calls = %d, want 2", calls)
	}
}

func TestPlayerObserverFindGoodSlave4E6280NilPlayerPanicsBeforeCallbacks(t *testing.T) {
	called := false
	defer func() {
		if recover() == nil {
			t.Fatal("nil player returned without panic")
		}
		if called {
			t.Fatal("callback ran before nil player panic")
		}
	}()
	playerObserverFindGoodSlave0_4E6280(
		nil,
		func(*server.Object) *server.Object { called = true; return nil },
		func(*server.Object) *server.Object { called = true; return nil },
	)
}

func TestPlayerObserverGoodSlaveTraversalDependencies(t *testing.T) {
	owner := &server.Object{}
	nonMonster := &server.Object{}
	notSummonedData := &server.MonsterUpdateData{}
	notSummoned := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(notSummonedData)}
	summonedData := &server.MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	summoned := &server.Object{
		ObjClass:   object.ClassMonster,
		UpdateData: unsafe.Pointer(summonedData),
		ObjOwner:   owner,
	}
	owner.Field129 = nonMonster
	nonMonster.Field128 = notSummoned
	notSummoned.Field128 = summoned

	if got := playerObserverFindGoodSlave2_4EC3E0(owner); got != summoned {
		t.Fatalf("first good slave = %p, want summoned monster %p", got, summoned)
	}
	if got := playerObserverFindGoodSlave_4EC420(notSummoned); got != nil {
		t.Fatalf("next with missing owner = %p, want nil", got)
	}
	notSummoned.ObjOwner = owner
	if got := playerObserverFindGoodSlave_4EC420(notSummoned); got != summoned {
		t.Fatalf("next good slave = %p, want summoned monster %p", got, summoned)
	}
	if got := playerObserverFindGoodSlave2_4EC3E0(nil); got != nil {
		t.Fatalf("first good slave for nil owner = %p, want nil", got)
	}
	if got := playerObserverFindGoodSlave_4EC420(nil); got != nil {
		t.Fatalf("next good slave for nil current = %p, want nil", got)
	}
}
