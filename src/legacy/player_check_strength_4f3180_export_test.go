package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type playerCheckStrengthLegacyServer4F3180 struct {
	Server
	srv *server.Server
}

func (s *playerCheckStrengthLegacyServer4F3180) S() *server.Server {
	return s.srv
}

func TestPlayerCheckStrengthExport4F3180PreservesNativePointersAndIgnoresCheat(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &playerCheckStrengthLegacyServer4F3180{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	player, freePlayerObject := alloc.New(server.Object{})
	defer freePlayerObject()
	item, freeItem := alloc.New(server.Object{})
	defer freeItem()
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	defer freeUpdate()
	playerValue, freePlayer := alloc.New(server.Player{})
	defer freePlayer()
	definition, freeDefinition := alloc.New(server.Modifier{})
	defer freeDefinition()

	player.ObjClass = object.ClassPlayer
	player.UpdateData = unsafe.Pointer(update)
	update.Player = playerValue
	playerValue.Info().SetField2239(29)
	item.ObjClass = object.ClassWeapon
	item.TypeInd = 0xbeef
	definition.TypeInd = 0xbeef
	definition.ReqStrength60 = 30
	srv.Modif.Dword_5d4594_251600 = definition

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{
			unsafe.Pointer(player),
			unsafe.Pointer(item),
			unsafe.Pointer(update),
			unsafe.Pointer(playerValue),
			unsafe.Pointer(definition),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}

	CheatEquipAll(true)
	t.Cleanup(func() { CheatEquipAll(false) })
	if got := playerCheckStrengthExportCall4F3180(player, item); got != 0 {
		t.Fatalf("CGo result with allow-all cheat = %d, want original result 0", got)
	}
	if got := Nox_xxx_playerCheckStrength_4F3180(player, item); got {
		t.Fatal("native wrapper accepted an item above the player's strength")
	}

	playerValue.Info().SetField2239(30)
	if got := playerCheckStrengthExportCall4F3180(player, item); got != 1 {
		t.Fatalf("CGo equality result = %d, want 1", got)
	}
	if got := Nox_xxx_playerCheckStrength_4F3180(player, item); !got {
		t.Fatal("native wrapper rejected exact required strength")
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(item)
	runtime.KeepAlive(update)
	runtime.KeepAlive(playerValue)
	runtime.KeepAlive(definition)
}
