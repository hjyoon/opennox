package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type secondaryWeaponLegacyServer53AB90 struct {
	Server
	srv *server.Server
}

func (s *secondaryWeaponLegacyServer53AB90) S() *server.Server {
	return s.srv
}

func TestSecondaryWeaponExport53AB90PreservesNativePointers(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &secondaryWeaponLegacyServer53AB90{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	item, freeItem := alloc.New(server.Object{})
	defer freeItem()
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	defer freeUpdate()
	playerValue, freePlayer := alloc.New(server.Player{})
	defer freePlayer()
	owner.ObjClass = object.ClassPlayer
	owner.UpdateData = unsafe.Pointer(update)
	update.Player = playerValue
	playerValue.Info().SetPlayerClass(player.Warrior)

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(item), unsafe.Pointer(update), unsafe.Pointer(playerValue)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}
	srv.SecondaryWeaponReport53AB90(owner, item,
		func(*server.Object, player.Class) bool { return true },
		func(*server.Object, *server.Object) bool { return true },
		func(byte) { t.Fatal("valid seed item cleared client") },
	)
	if got := srv.SecondaryWeapon53AB90(owner); got != item {
		t.Fatalf("seed secondary weapon = %p, want %p", got, item)
	}

	secondaryWeaponExportCall53AB90(owner, nil)
	if got := srv.SecondaryWeapon53AB90(owner); got != nil {
		t.Fatalf("CGo-cleared secondary weapon = %p, want nil", got)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
	runtime.KeepAlive(update)
	runtime.KeepAlive(playerValue)
}
