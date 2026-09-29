package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

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

func TestObjectChangeGoldKeepsNativeObjectPointer(t *testing.T) {
	player := &server.Player{GoldVal: 250}
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatal("expected native unit pointer above 4 GiB")
	}

	asObjectS(unit).ChangeGold(-200)
	if player.GoldVal != 50 {
		t.Fatalf("gold after -200 = %d, want 50", player.GoldVal)
	}
}
