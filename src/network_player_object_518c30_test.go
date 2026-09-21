package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/server"
)

func playerObjectTestUnit518C30(state server.PlayerState) (*server.Object, *server.PlayerUpdateData) {
	player := &server.Player{}
	update := &server.PlayerUpdateData{State: state, Player: player}
	unit := &server.Object{
		TypeInd:    0x1234,
		ObjClass:   object.ClassPlayer,
		NetCode:    0x2345,
		PosVec:     types.Ptf(10.5, 11.5),
		Direction1: 0,
		UpdateData: unsafe.Pointer(update),
	}
	player.PlayerUnit = unit
	return unit, update
}

func TestPlayerObjectPacketNative518C30UsesNamedNativeFields(t *testing.T) {
	unit, _ := playerObjectTestUnit518C30(server.PlayerState0)
	s := &Server{Server: &server.Server{}}
	got := s.playerObjectPacketNative518C30(unit)
	want := [12]byte{
		0xc3,
		0x45, 0x23,
		0x34, 0x12,
		0x0a, 0x00,
		0x0c, 0x00,
		0x40,
		0xff,
		0x04,
	}
	if got != want {
		t.Fatalf("packet = % x, want % x", got, want)
	}
}

func TestPlayerObjectVisualStateNative518C30(t *testing.T) {
	data := []byte{0xc3, 2, 0, 4, 0, 10, 0, 20, 0, 0x70, 9, 40}
	state, ok := decodePlayerObjectVisualState518C30(data)
	if !ok {
		t.Fatal("valid player-object packet was rejected")
	}
	if state.X != 10 || state.Y != 20 || state.Direction != 8 || state.Frame != 9 || state.Animation != 40 {
		t.Fatalf("decoded state = %+v", state)
	}
	dr := &client.Drawable{AnimFrameSlave: 3, AnimInd: 4, AnimStart: 5}
	applyPlayerObjectVisualState518C30(dr, state, 100)
	if dr.Field_72 != 100 || dr.AnimFrameSlave != 9 || dr.Field_78 != 3 || dr.AnimDir != 8 || dr.AnimInd != 40 || dr.AnimStart != 100 {
		t.Fatalf("drawable state was not applied through native fields: %+v", dr)
	}
	if _, ok := decodePlayerObjectVisualState518C30(data[:11]); ok {
		t.Fatal("short player-object packet was accepted")
	}
}

func TestPlayerAnimationStateNative4FA2B0FixedMappings(t *testing.T) {
	s := &Server{Server: &server.Server{}}
	tests := []struct {
		state server.PlayerState
		want  byte
	}{
		{server.PlayerState0, 4},
		{server.PlayerState2, 21},
		{server.PlayerState3, 1},
		{server.PlayerState4, 2},
		{server.PlayerState5, 6},
		{server.PlayerState12, 3},
		{server.PlayerState15, 40},
		{server.PlayerState18, 48},
		{server.PlayerState19, 49},
		{server.PlayerState20, 47},
		{server.PlayerState21, 30},
		{server.PlayerState23, 50},
		{server.PlayerState24, 19},
		{server.PlayerStateShakeFist, 20},
		{server.PlayerStateLaugh, 15},
		{server.PlayerStatePoint, 16},
		{server.PlayerState30, 52},
		{server.PlayerState32, 54},
	}
	for _, tc := range tests {
		unit, _ := playerObjectTestUnit518C30(tc.state)
		if got := s.playerAnimationStateNative4FA2B0(unit); got != tc.want {
			t.Errorf("state %d mapped to %d, want %d", tc.state, got, tc.want)
		}
	}
}

func TestPlayerObjectDirection518C30Cardinals(t *testing.T) {
	tests := []struct {
		direction server.Dir16
		want      byte
	}{
		{0, 4},
		{64, 6},
		{128, 3},
		{192, 1},
	}
	for _, tc := range tests {
		if got := playerObjectDirection518C30(tc.direction); got != tc.want {
			t.Errorf("direction %d mapped to %d, want %d", tc.direction, got, tc.want)
		}
	}
}

func TestPlayerReportSelf518CAFUsesNativeIdentityGate(t *testing.T) {
	first := new(server.Object)
	second := new(server.Object)
	var got []*server.Object
	report := func(unit *server.Object) { got = append(got, unit) }
	playerReportSelf518CAF(first, second, report)
	playerReportSelf518CAF(first, first, report)
	playerReportSelf518CAF[*server.Object](nil, nil, report)
	if len(got) != 2 || got[0] != first || got[1] != nil {
		t.Fatalf("reported objects = %v, want [%p nil]", got, first)
	}
}

func TestPlayerReportVitalsNative4D9900ReportsChangesAndReloadsCaches(t *testing.T) {
	first := &server.Player{PlayerInd: 7}
	second := &server.Player{PlayerInd: 9}
	update := &server.PlayerUpdateData{
		ManaCur:  75,
		ManaPrev: 74,
		Field2_1: 38,
		Player:   first,
	}
	unit := &server.Object{
		ObjClass:   object.ClassPlayer,
		HealthData: &server.HealthData{Cur: 37, Max: 75},
		UpdateData: unsafe.Pointer(update),
	}
	var calls []string
	playerReportVitalsNative4D9900(unit, func(ind byte, got *server.Object) {
		if ind != 7 || got != unit {
			t.Fatalf("health report = ind:%d unit:%p, want 7/%p", ind, got, unit)
		}
		calls = append(calls, "health")
		unit.HealthData.Cur = 36
		update.Player = second
	}, func(ind byte, got *server.Object) {
		if ind != 9 || got != unit {
			t.Fatalf("mana report = ind:%d unit:%p, want 9/%p", ind, got, unit)
		}
		calls = append(calls, "mana")
		update.ManaCur = 77
	})
	if len(calls) != 2 || calls[0] != "health" || calls[1] != "mana" {
		t.Fatalf("report order = %v, want [health mana]", calls)
	}
	if update.Field2_1 != 36 || update.ManaPrev != 77 {
		t.Fatalf("reported caches = health:%d mana:%d, want 36/77", update.Field2_1, update.ManaPrev)
	}
}

func TestPlayerReportVitalsNative4D9900SkipsUnchangedAndInvalidObjects(t *testing.T) {
	report := func(byte, *server.Object) {
		t.Fatal("unexpected vital report")
	}
	playerReportVitalsNative4D9900(nil, report, report)
	playerReportVitalsNative4D9900(&server.Object{ObjClass: object.ClassMonster}, report, report)
	playerReportVitalsNative4D9900(&server.Object{ObjClass: object.ClassPlayer}, report, report)

	player := &server.Player{PlayerInd: 4}
	update := &server.PlayerUpdateData{ManaCur: 75, ManaPrev: 75, Field2_1: 37, Player: player}
	unit := &server.Object{
		ObjClass:   object.ClassPlayer,
		HealthData: &server.HealthData{Cur: 37, Max: 75},
		UpdateData: unsafe.Pointer(update),
	}
	playerReportVitalsNative4D9900(unit, report, report)
}

func TestPlayerReportVitalsNative4D9900ReportsManaWithoutHealthData(t *testing.T) {
	player := &server.Player{PlayerInd: 5}
	update := &server.PlayerUpdateData{ManaCur: 13, ManaPrev: 12, Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var manaReports int
	playerReportVitalsNative4D9900(unit, func(byte, *server.Object) {
		t.Fatal("health reported without HealthData")
	}, func(ind byte, got *server.Object) {
		manaReports++
		if ind != 5 || got != unit {
			t.Fatalf("mana report = ind:%d unit:%p, want 5/%p", ind, got, unit)
		}
	})
	if manaReports != 1 || update.ManaPrev != 13 {
		t.Fatalf("mana reports/cache = %d/%d, want 1/13", manaReports, update.ManaPrev)
	}
}
