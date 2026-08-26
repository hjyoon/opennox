package opennox

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

func TestMonstersAllBelongToHost4DB6A0Guards(t *testing.T) {
	called := false
	h := monstersAllBelongToHostHooks4DB6A0{
		playerByInd: func(ind int) *server.Player {
			if ind != server.HostPlayerIndex {
				t.Fatalf("player index = %d, want %d", ind, server.HostPlayerIndex)
			}
			return nil
		},
		playerUnit: func(*server.Player) *server.Object { called = true; return nil },
	}
	monstersAllBelongToHostNative4DB6A0(h)
	if called {
		t.Fatal("nil player read PlayerUnit")
	}

	pl := &server.Player{}
	h.playerByInd = func(int) *server.Player { return pl }
	monstersAllBelongToHostNative4DB6A0(h)
	if !called {
		t.Fatal("non-nil player did not read PlayerUnit")
	}
}

func TestMonstersAllBelongToHost4DB6A0OrderAndWidths(t *testing.T) {
	initialUnit := &server.Object{}
	firstOwner := &server.Object{}
	secondOwner := &server.Object{}
	pl := &server.Player{PlayerUnit: initialUnit, PlayerInd: 0xe1}

	other := &server.Object{TypeInd: 7}
	loc := &server.Object{TypeInd: 0xfffe, ScriptIDVal: -1}
	other.ObjNext = loc

	firstData := &server.MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	first := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(firstData)}
	second := &server.Object{ObjClass: object.ClassMonster}
	thirdData := &server.MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	third := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(thirdData)}
	loc.Field129 = first
	first.Field128 = second

	var events []string
	add := func(s string) { events = append(events, s) }
	unitReads := 0
	indexReads := 0
	h := monstersAllBelongToHostHooks4DB6A0{
		playerByInd: func(ind int) *server.Player {
			add(fmt.Sprintf("player:%d", ind))
			return pl
		},
		playerUnit: func(got *server.Player) *server.Object {
			unitReads++
			add(fmt.Sprintf("unit:%d", unitReads))
			if got != pl {
				t.Fatalf("PlayerUnit source = %p, want %p", got, pl)
			}
			return got.PlayerUnit
		},
		saveLocationType: func() uint32 {
			add("type")
			return 0xfffe
		},
		firstObject: func() *server.Object { add("first-object"); return other },
		nextObject: func(obj *server.Object) *server.Object {
			add("next-object")
			return obj.ObjNext
		},
		typeInd: func(obj *server.Object) uint16 {
			add(fmt.Sprintf("type-ind:%d", obj.TypeInd))
			return obj.TypeInd
		},
		firstOwned: func(obj *server.Object) *server.Object {
			add("first-owned")
			return obj.Field129
		},
		nextOwned: func(obj *server.Object) *server.Object {
			switch obj {
			case first:
				add("next:first")
			case second:
				add("next:second")
			default:
				t.Fatalf("unexpected owned object %p", obj)
			}
			return obj.Field128
		},
		setOwner: func(owner, obj *server.Object) {
			switch obj {
			case first:
				add("owner:first")
				if owner != firstOwner {
					t.Fatalf("first owner = %p, want %p", owner, firstOwner)
				}
				// ObjSetOwner is allowed to rewrite both the list and later state.
				first.Field128 = third
				first.ObjClass = 0
				pl.PlayerUnit = secondOwner
			case second:
				add("owner:second")
				if owner != secondOwner {
					t.Fatalf("second owner = %p, want %p", owner, secondOwner)
				}
				second.ObjClass = object.ClassMonster
				second.UpdateData = unsafe.Pointer(thirdData)
				second.Field128 = third
			default:
				t.Fatalf("stale successor visited: %p", obj)
			}
		},
		class: func(obj *server.Object) object.Class {
			if obj == first {
				add("class:first")
			} else {
				add("class:second")
			}
			return obj.ObjClass
		},
		isSummoned: func(obj *server.Object) bool {
			add("summoned:second")
			return obj.UpdateDataMonster().StatusFlags.Has(object.MonStatusSummoned)
		},
		markMonitor: func(obj *server.Object) {
			add("monitor:second")
			obj.ObjSubClass |= object.SubClass(object.MonsterMonitor)
			pl.PlayerInd = 0xf2
		},
		playerIndex: func(got *server.Player) byte {
			indexReads++
			add(fmt.Sprintf("index:%d", indexReads))
			if got != pl {
				t.Fatalf("player-index source = %p, want %p", got, pl)
			}
			v := got.PlayerInd
			got.PlayerInd++
			return v
		},
		reportAcquire: func(ind int, obj *server.Object) {
			add(fmt.Sprintf("acquire:%#x", ind))
			if obj != second {
				t.Fatalf("acquired object = %p, want %p", obj, second)
			}
		},
		markMinimap: func(ind ntype.PlayerInd, obj *server.Object, flags uint32) {
			add(fmt.Sprintf("minimap:%#x:%d", byte(ind), flags))
			if obj != second {
				t.Fatalf("minimap object = %p, want %p", obj, second)
			}
		},
		clearScriptID: func(obj *server.Object) {
			add("clear-script")
			obj.ScriptIDVal = 0
		},
		delayedDelete: func(obj *server.Object) {
			add("delete")
			if obj.ScriptIDVal != 0 {
				t.Fatalf("delete preceded ScriptID clear: %d", obj.ScriptIDVal)
			}
		},
	}
	pl.PlayerUnit = firstOwner
	monstersAllBelongToHostNative4DB6A0(h)

	want := []string{
		"player:31", "unit:1", "type", "first-object", "type-ind:7", "next-object", "type-ind:65534",
		"first-owned", "next:first", "unit:2", "owner:first", "class:first",
		"next:second", "unit:3", "owner:second", "class:second", "summoned:second", "monitor:second",
		"index:1", "acquire:0xf2", "index:2", "minimap:0xf3:1", "clear-script", "delete",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events:\n got %v\nwant %v", events, want)
	}
	if !second.SubClass().AsMonster().Has(object.MonsterMonitor) {
		t.Fatal("summoned monster was not marked monitored")
	}
	if third.ObjSubClass != 0 {
		t.Fatal("traversal followed successor rewritten by ObjSetOwner")
	}
}
