package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type playerManaRechargeLegacyServer4FD030 struct {
	Server
	srv *server.Server
}

func (s *playerManaRechargeLegacyServer4FD030) S() *server.Server {
	return s.srv
}

func TestPlayerManaRechargeExport4FD030PreservesNativePointerAndWidths(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &playerManaRechargeLegacyServer4FD030{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	player := &server.Player{ProtUnitManaCur: 0}
	update := &server.PlayerUpdateData{ManaCur: 10, ManaPrev: 99, ManaMax: 100, Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	var pin runtime.Pinner
	pin.Pin(player)
	pin.Pin(update)
	pin.Pin(unit)
	defer pin.Unpin()

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"unit":   uintptr(unsafe.Pointer(unit)),
			"update": uintptr(unsafe.Pointer(update)),
			"player": uintptr(unsafe.Pointer(player)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	if got := playerManaRechargeExportCall4FD030(unit, -3); got != 100 {
		t.Fatalf("Player result = %d, want maximum mana 100", got)
	}
	if update.ManaPrev != 10 || update.ManaCur != 7 {
		t.Fatalf("previous/current = %d/%d, want 10/7", update.ManaPrev, update.ManaCur)
	}

	unit.ObjClass = object.ClassMonster
	want := uint16(uintptr(unsafe.Pointer(unit)))
	if got := playerManaRechargeExportCall4FD030(unit, math.MaxInt16); got != want {
		t.Fatalf("non-Player result = %#x, want pointer low word %#x", got, want)
	}
	if update.ManaPrev != 10 || update.ManaCur != 7 {
		t.Fatalf("non-Player changed mana to previous/current %d/%d", update.ManaPrev, update.ManaCur)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}
