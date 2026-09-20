package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

type monsterActionRefreshLegacyServer50A910 struct {
	Server
	srv *server.Server
}

func (s *monsterActionRefreshLegacyServer50A910) S() *server.Server { return s.srv }

func TestMonsterActionRefreshExport50A910PreservesNativePointers(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &monsterActionRefreshLegacyServer50A910{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	unit := &server.Object{ObjClass: object.ClassMonster}
	update := new(server.MonsterUpdateData)
	destroyed := &server.Object{ObjFlags: object.FlagDestroyed}
	unit.UpdateData = unsafe.Pointer(update)
	update.PreferredEnemy = destroyed
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_FACE_OBJECT)
	update.AIStack[0].Args[0] = uintptr(unsafe.Pointer(destroyed))

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"unit":      uintptr(unsafe.Pointer(unit)),
			"update":    uintptr(unsafe.Pointer(update)),
			"destroyed": uintptr(unsafe.Pointer(destroyed)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want address above the ABI32 range", name, pointer)
			}
		}
	}

	Nox_xxx_mobAction_50A910(unit)
	if update.PreferredEnemy != nil {
		t.Fatalf("destroyed preferred enemy survived: %p", update.PreferredEnemy)
	}
	if got := update.AIStack[0].Args[0]; got != 0 {
		t.Fatalf("destroyed action argument = %#x, want cleared", got)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(destroyed)
}
