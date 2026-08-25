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
