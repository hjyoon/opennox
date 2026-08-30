package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type playerActionStateLegacyServer4FA2B0 struct {
	Server
	srv *server.Server
}

func (s *playerActionStateLegacyServer4FA2B0) S() *server.Server {
	return s.srv
}

func TestPlayerActionStateExport4FA2B0PreservesNativePointerChain(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &playerActionStateLegacyServer4FA2B0{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	var useData [97]byte
	useData[96] = 2
	weapon := &server.Object{UseData: server.UseDataPtr{Ptr: unsafe.Pointer(&useData[0])}}
	player := &server.Player{WeaponEquip: 0x10000}
	update := &server.PlayerUpdateData{
		State:          server.PlayerState1,
		EquippedWeapon: weapon,
		Player:         player,
	}
	unit := &server.Object{UpdateData: unsafe.Pointer(update)}
	var pin runtime.Pinner
	pin.Pin(&useData[0])
	pin.Pin(weapon)
	pin.Pin(player)
	pin.Pin(update)
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]unsafe.Pointer{
			"unit": unsafe.Pointer(unit), "update": unsafe.Pointer(update),
			"player": unsafe.Pointer(player), "weapon": unsafe.Pointer(weapon),
			"use-data": unsafe.Pointer(&useData[0]),
		} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, ptr)
			}
		}
	}

	if got := playerActionStateExportCall4FA2B0(unit); got != 29 {
		t.Fatalf("export result = %d, want 29", got)
	}
	runtime.KeepAlive(useData)
	runtime.KeepAlive(weapon)
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}
