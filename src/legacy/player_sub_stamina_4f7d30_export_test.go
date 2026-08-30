package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

type playerSubStaminaLegacyServer4F7D30 struct {
	Server
	srv *server.Server
}

func (s *playerSubStaminaLegacyServer4F7D30) S() *server.Server {
	return s.srv
}

func TestPlayerSubStaminaExport4F7D30PreservesNativePointerAndSignedAmount(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &playerSubStaminaLegacyServer4F7D30{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	update := &server.MonsterUpdateData{Stamina: 5}
	unit := &server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	var pin runtime.Pinner
	pin.Pin(update)
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if pointer := uintptr(unsafe.Pointer(unit)); pointer <= math.MaxUint32 {
			t.Fatalf("unit pointer = %#x, want native address above 4 GiB", pointer)
		}
		if pointer := uintptr(unsafe.Pointer(update)); pointer <= math.MaxUint32 {
			t.Fatalf("update pointer = %#x, want native address above 4 GiB", pointer)
		}
	}

	if got := playerSubStaminaExportCall4F7D30(unit, -1); got != 1 || update.Stamina != 6 {
		t.Fatalf("export result = %d stamina=%d, want 1/6", got, update.Stamina)
	}
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}
