package opennox

import (
	"encoding/json"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestElevatorBoardingObservationCopiesLiveValuesWithoutMutation(t *testing.T) {
	health := server.HealthData{Cur: 17, Max: 20}
	playerData := server.PlayerUpdateData{State: server.PlayerState13}
	player := &server.Object{
		ObjClass: object.ClassPlayer, ObjFlags: object.FlagActive | object.FlagOnObject,
		ScriptIDVal: 1000000001, PosVec: types.Ptf(99, 199), NewPos: types.Ptf(100, 200),
		ForceVec: types.Ptf(-3, 4),
		ZVal:     4, Field27: -0.5, HealthData: &health, UpdateData: unsafe.Pointer(&playerData),
	}
	player.Shape.Kind, player.Shape.Circle.R = server.ShapeKindCircle, 10
	data := server.ElevatorUpdateData{Field_2: 201, Field_3: 3, Field_4: 20}
	elevator := &server.Object{
		ObjClass: object.ClassElevator, ObjFlags: object.FlagEnabled, Extent: 200,
		ScriptIDVal: 1000052653, PosVec: types.Ptf(100, 200), NewPos: types.Ptf(100, 200),
		UpdateData: unsafe.Pointer(&data),
	}
	elevator.Shape.Kind = server.ShapeKindBox
	elevator.Shape.Box.W, elevator.Shape.Box.H = 32, 32
	elevator.Shape.Box.Calc()
	playerBefore, elevatorBefore, dataBefore, healthBefore, playerDataBefore := *player, *elevator, data, health, playerData
	got, err := e2eReadElevatorBoarding(674, player, elevator)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*player, playerBefore) || !reflect.DeepEqual(*elevator, elevatorBefore) ||
		data != dataBefore || health != healthBefore || !reflect.DeepEqual(playerData, playerDataBefore) {
		t.Fatal("observing a boarded passenger must not change any live record")
	}
	if got.Frame != 674 || got.Player.ScriptID != 1000000001 || got.Elevator.Extent != 200 ||
		got.ShaftExtent != 201 || got.Height != 20 || got.State != 3 || got.Player.Z != 4 ||
		got.HeightDifference != 16 || !got.PlayerInBox || !got.PlayerNewPosInBox ||
		got.PlayerHealth != 17 || got.PlayerMaximumHealth != 20 || got.LinkPresent || got.Player.Force != types.Ptf(-3, 4) {
		t.Fatalf("missing original boarding-height observations: %+v", got)
	}
	data.Field_4, player.ZVal, health.Cur = 42, 40, 3
	player.PosVec.X, elevator.Shape.Box.W = 600, 64
	if got.Height != 20 || got.Player.Z != 4 || got.PlayerHealth != 17 || got.Player.Pos.X != 99 || got.Elevator.Shape.Box.W != 32 {
		t.Fatal("later world updates must not change an earlier observation")
	}
	if _, err := json.Marshal(got); err != nil {
		t.Fatal(err)
	}
}

func TestElevatorBoardingObservationRejectsMissingOrWrongRecords(t *testing.T) {
	player := &server.Object{ObjClass: object.ClassPlayer}
	data := server.ElevatorUpdateData{}
	elevator := &server.Object{ObjClass: object.ClassElevator, UpdateData: unsafe.Pointer(&data)}
	for _, test := range []struct{ player, elevator *server.Object }{
		{nil, elevator}, {&server.Object{ObjClass: object.ClassMonster}, elevator},
		{player, nil}, {player, &server.Object{ObjClass: object.ClassElevator}},
		{player, &server.Object{ObjClass: object.ClassElevatorShaft, UpdateData: unsafe.Pointer(&data)}},
	} {
		if _, err := e2eReadElevatorBoarding(1, test.player, test.elevator); err == nil {
			t.Fatal("missing or differently typed native records must not be read as an elevator")
		}
	}
}

func TestElevatorBoardingObservationRetainsNoPointers(t *testing.T) {
	var check func(reflect.Type)
	check = func(typ reflect.Type) {
		switch typ.Kind() {
		case reflect.Pointer, reflect.UnsafePointer, reflect.Map, reflect.Interface,
			reflect.Slice, reflect.Func, reflect.Chan:
			t.Fatalf("boarding snapshots must not retain native or Go references: %v", typ)
		case reflect.Struct:
			for i := 0; i < typ.NumField(); i++ {
				check(typ.Field(i).Type)
			}
		case reflect.Array:
			check(typ.Elem())
		}
	}
	check(reflect.TypeOf(e2eElevatorBoardingObservation{}))
}
