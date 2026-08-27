package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestObjectPlayerStrengthUsesNativePlayerInfo(t *testing.T) {
	player := &Player{}
	player.Info().SetField2239(37)
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: player}),
	}

	if got := unit.Strength(); got != 37 {
		t.Fatalf("Strength() = %d, want 37", got)
	}
	unit.SetStrength(63)
	if got := player.Info().Field2239(); got != 63 {
		t.Fatalf("player info strength = %d, want 63", got)
	}
}
