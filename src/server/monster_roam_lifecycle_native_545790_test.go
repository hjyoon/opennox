package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestMonsterRoamWaypointLifecycleOrdinaryAdmission545790(t *testing.T) {
	var s Server
	for _, entry := range []string{"start", "cancel"} {
		t.Run(entry, func(t *testing.T) {
			ordinary := func(unit *Object) {
				if entry == "start" {
					s.MonsterActionRoamStart545790(unit)
				} else {
					s.MonsterActionRoamCancel5457C0(unit)
				}
			}
			ordinary(nil)
			unit, update := Object{ObjClass: object.ClassPlayer}, MonsterUpdateData{Field91: 77, AIStackInd: 0}
			unit.UpdateData = unsafe.Pointer(&update)
			update.AIStack[0].Args[0] = 0x12345678
			before := update
			ordinary(&unit)
			if !reflect.DeepEqual(update, before) {
				t.Fatal("ordinary wrong-class admission changed")
			}
			unit.ObjClass, unit.UpdateData = object.ClassMonster, nil
			ordinary(&unit)
			var fault any
			func() {
				defer func() { fault = recover() }()
				if entry == "start" {
					unit.MonsterActionRoamStartNative545790()
				} else {
					unit.MonsterActionRoamCancelNative5457C0()
				}
			}()
			if fault == nil {
				t.Fatal("raw missing-update original fault was hidden by ordinary admission")
			}
		})
	}
	t.Run("cancel-empty-head", func(t *testing.T) {
		update := MonsterUpdateData{AIStackInd: -1, Field133: 0x12345678}
		unit := Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(&update)}
		before := update
		s.MonsterActionRoamCancel5457C0(&unit)
		if !reflect.DeepEqual(update, before) {
			t.Fatal("ordinary empty-head admission changed")
		}
	})
	t.Run("cancel-invalid-positive-head", func(t *testing.T) {
		update := MonsterUpdateData{AIStackInd: 24}
		unit := Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(&update)}
		var fault any
		func() { defer func() { fault = recover() }(); s.MonsterActionRoamCancel5457C0(&unit) }()
		if fault == nil {
			t.Fatal("ordinary existing invalid-index fault changed")
		}
	})
}
