package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerGoldWrappersKeepNativeObjectPointer(t *testing.T) {
	player := &server.Player{GoldVal: 20}
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{UpdateData: unsafe.Pointer(update)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= 0xffffffff {
		t.Fatal("expected native unit pointer above 4 GiB")
	}
	Nox_xxx_playerAddGold_4FA590(unit, 5)
	if player.GoldVal != 25 {
		t.Fatalf("gold after add = %d, want 25", player.GoldVal)
	}
	Nox_xxx_playerSubGold_4FA5D0(unit, 7)
	if player.GoldVal != 18 {
		t.Fatalf("gold after subtract = %d, want 18", player.GoldVal)
	}
	Nox_xxx_playerSubGold_4FA5D0(unit, 100)
	if player.GoldVal != 0 {
		t.Fatalf("gold after saturating subtract = %d, want 0", player.GoldVal)
	}
}
