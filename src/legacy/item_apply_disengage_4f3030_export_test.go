package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type itemApplyDisengageLegacyServer4F3030 struct {
	Server
	srv *server.Server
}

func (s *itemApplyDisengageLegacyServer4F3030) S() *server.Server {
	return s.srv
}

func TestItemApplyDisengageExport4F3030InvokesNativeCallbacks(t *testing.T) {
	oldGetServer := GetServer
	GetServer = func() Server {
		return &itemApplyDisengageLegacyServer4F3030{srv: new(server.Server)}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	item := new(server.Object)
	owner := new(server.Object)
	attrs := new(server.ModifierInitData)
	brilliance := new(server.ModifierEff)
	fireProtect := new(server.ModifierEff)
	var pin runtime.Pinner
	pin.Pin(item)
	pin.Pin(owner)
	pin.Pin(attrs)
	pin.Pin(brilliance)
	pin.Pin(fireProtect)
	defer pin.Unpin()

	brilliance.Disengage116 = modifierDisengagePointerNative4DFBB0(8)
	fireProtect.Disengage116 = modifierDisengagePointerNative4DFBB0(1)
	if brilliance.Disengage116 == nil || fireProtect.Disengage116 == nil {
		t.Fatal("registered native disengage callback pointer is nil")
	}
	attrs.Modifiers[2] = brilliance
	attrs.Modifiers[3] = fireProtect
	item.InitData = unsafe.Pointer(attrs)

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{
			unsafe.Pointer(item),
			unsafe.Pointer(owner),
			unsafe.Pointer(attrs),
			unsafe.Pointer(brilliance),
			unsafe.Pointer(fireProtect),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}

	owner.Field110 = 9
	itemApplyDisengageExportCall4F3030(item, owner)
	if got := owner.Field110; got != 0 {
		t.Fatalf("owner disengage mask = %#x, want 0 after Brilliance|FireProtect removal", got)
	}
	runtime.KeepAlive(item)
	runtime.KeepAlive(owner)
	runtime.KeepAlive(attrs)
	runtime.KeepAlive(brilliance)
	runtime.KeepAlive(fireProtect)
}
