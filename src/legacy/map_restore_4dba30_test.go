package legacy

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestMapRestoreObjects4DBA30NativeBranchesAndCachedSuccessors(t *testing.T) {
	playerUnit := &server.Object{ScriptIDVal: 1}
	marker := &server.Object{TypeInd: 9, ScriptIDVal: 77, PosVec: types.Ptf(12, 34)}
	monsterData := new(server.MonsterUpdateData)
	monster := &server.Object{ScriptIDVal: 2, ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(monsterData)}
	elevatorData := &server.ElevatorUpdateData{Field_4: 1}
	elevator := &server.Object{ScriptIDVal: 3, ObjClass: object.ClassElevator, UpdateData: unsafe.Pointer(elevatorData)}
	shaft := &server.Object{ScriptIDVal: 4, ObjClass: object.ClassElevatorShaft}
	doorData := &server.DoorUpdateData{TargetDirection: 3, CurrentDirection: 2}
	door := &server.Object{ScriptIDVal: 5, ObjClass: object.ClassDoor, UpdateData: unsafe.Pointer(doorData)}
	stillDoorData := &server.DoorUpdateData{TargetDirection: 6, CurrentDirection: 6}
	stillDoor := &server.Object{ScriptIDVal: 6, ObjClass: object.ClassDoor, UpdateData: unsafe.Pointer(stillDoorData)}
	marker.ObjNext = monster
	monster.ObjNext = elevator
	elevator.ObjNext = shaft
	shaft.ObjNext = door
	door.ObjNext = stillDoor

	var events []string
	mapRestoreObjects4DBA30(marker, 9, mapRestoreObjectsHooks4DBA30{
		playerUnit: func() *server.Object { return playerUnit },
		moveToMarker: func(gotPlayer, gotMarker *server.Object) {
			if gotPlayer != playerUnit || gotMarker != marker {
				t.Fatalf("move args = %p/%p", gotPlayer, gotMarker)
			}
			events = append(events, "move")
		},
		delayedDelete: func(obj *server.Object) {
			events = append(events, fmt.Sprintf("delete:%d", obj.ScriptIDVal))
			obj.ObjNext = nil
		},
		resolveMonster: func(obj *server.Object) {
			events = append(events, fmt.Sprintf("monster:%d", obj.ScriptIDVal))
		},
		needSync: func(obj *server.Object) {
			events = append(events, fmt.Sprintf("sync:%d", obj.ScriptIDVal))
		},
		addUpdatable: func(obj *server.Object) {
			events = append(events, fmt.Sprintf("update:%d", obj.ScriptIDVal))
		},
		elevatorLink: func(obj *server.Object) *server.Object {
			if obj == shaft {
				return elevator
			}
			return nil
		},
	})
	want := []string{"move", "delete:0", "monster:2", "sync:3", "sync:4", "update:5"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if playerUnit.ScriptIDVal != 77 || marker.ScriptIDVal != 0 {
		t.Fatalf("script IDs = player %d marker %d", playerUnit.ScriptIDVal, marker.ScriptIDVal)
	}
}

func TestMapRestoreObjects4DBA30PreservesClassPriority(t *testing.T) {
	update := &server.ElevatorUpdateData{Field_4: 0}
	hybrid := &server.Object{ObjClass: object.ClassElevator | object.ClassDoor, UpdateData: unsafe.Pointer(update)}
	called := false
	mapRestoreObjects4DBA30(hybrid, 99, mapRestoreObjectsHooks4DBA30{
		playerUnit:     func() *server.Object { return &server.Object{} },
		resolveMonster: func(*server.Object) {},
		needSync:       func(*server.Object) { t.Fatal("inactive elevator requested sync") },
		addUpdatable:   func(*server.Object) { called = true },
		elevatorLink:   func(*server.Object) *server.Object { return nil },
	})
	if called {
		t.Fatal("Door branch ran after the higher-priority Elevator branch")
	}
}

func TestMapRestoreCleanup4DBA30NativeLists(t *testing.T) {
	glyph := &server.Object{TypeInd: 17, ScriptIDVal: 11}
	otherItem := &server.Object{TypeInd: 18, ScriptIDVal: 12}
	glyph.InvNextItem = otherItem
	migrating := &server.Object{
		ScriptIDVal:  20,
		ObjClass:     object.ClassMonster,
		ObjSubClass:  object.SubClass(0x2000),
		InvFirstItem: glyph,
	}
	negative := &server.Object{ScriptIDVal: 21, ObjFlags: object.Flags(0x80000000)}
	ordinary := &server.Object{ScriptIDVal: 22}
	migrating.ObjNext = negative
	negative.ObjNext = ordinary
	pixie := &server.Object{ScriptIDVal: 30}
	notPixie := &server.Object{ScriptIDVal: 31}
	pixie.ObjNext = notPixie

	var deleted []int32
	mapRestoreCleanup4DBA30(migrating, pixie, 17, mapRestoreCleanupHooks4DBA30{
		isOfflineMigratingMonster: func(obj *server.Object) bool {
			return obj == migrating || obj == negative
		},
		isCoopPlayerPixie: func(obj *server.Object) bool {
			return obj == pixie
		},
		delayedDelete: func(obj *server.Object) {
			deleted = append(deleted, obj.ScriptIDVal)
			obj.ObjNext = nil
			obj.InvNextItem = nil
		},
	})
	want := []int32{11, 20, 30}
	if !reflect.DeepEqual(deleted, want) {
		t.Fatalf("deleted = %v, want %v", deleted, want)
	}
}

func TestMapRestoreOwned4DBA30Notifications(t *testing.T) {
	nonMonster := &server.Object{ScriptIDVal: 1}
	summonedData := &server.MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	summoned := &server.Object{
		ScriptIDVal: 2,
		ObjClass:    object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterMonitor),
		UpdateData: unsafe.Pointer(summonedData),
	}
	monitor := &server.Object{
		ScriptIDVal: 3,
		ObjClass:    object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterMonitor),
	}
	ordinaryData := new(server.MonsterUpdateData)
	ordinary := &server.Object{ScriptIDVal: 4, ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(ordinaryData)}
	nonMonster.Field128 = summoned
	summoned.Field128 = monitor
	monitor.Field128 = ordinary

	var events []string
	mapRestoreOwned4DBA30(0xfe, nonMonster, mapRestoreOwnedHooks4DBA30{
		reportAcquire: func(ind byte, obj *server.Object) {
			events = append(events, fmt.Sprintf("acquire:%02x:%d", ind, obj.ScriptIDVal))
		},
		monitor: func(ind byte, obj *server.Object) {
			events = append(events, fmt.Sprintf("monitor:%02x:%d", ind, obj.ScriptIDVal))
		},
		markMinimap: func(ind byte, obj *server.Object) {
			events = append(events, fmt.Sprintf("mark:%02x:%d", ind, obj.ScriptIDVal))
		},
	})
	want := []string{"acquire:fe:2", "mark:fe:2", "monitor:fe:3", "mark:fe:3"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
