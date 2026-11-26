package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerManaRefreshProtectResultUsesNativeWidth(t *testing.T) {
	lineMessageTestWorld(t)
	token := int32(-123)
	player := &server.Player{ProtUnitManaCur: uint32(token)}
	update := &server.PlayerUpdateData{Player: player, ManaCur: 17, ManaMax: 0xfffe}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	want := uintptr(token)
	if got := playerManaRefreshCall4EECF0(unit); got != want {
		t.Fatalf("protection result = %#x, want %#x", got, want)
	}
	if update.ManaCur != 0xfffe || update.ManaPrev != 17 {
		t.Fatalf("mana = current %#x previous %d, want 0xfffe and 17", update.ManaCur, update.ManaPrev)
	}
}
