package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestObjectAddGoldNegativeSubtractsMagnitude(t *testing.T) {
	player := &server.Player{GoldVal: 20}
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{UpdateData: unsafe.Pointer(update)}
	obj := asObjectS(unit)
	obj.AddGold(-7)
	if player.GoldVal != 13 {
		t.Fatalf("gold after -7 = %d, want 13", player.GoldVal)
	}
	obj.AddGold(3)
	if player.GoldVal != 16 {
		t.Fatalf("gold after +3 = %d, want 16", player.GoldVal)
	}
}
