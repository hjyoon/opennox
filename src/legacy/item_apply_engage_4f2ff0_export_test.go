package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type itemApplyEngageLegacyServer4F2FF0 struct {
	Server
	srv *server.Server
}

func (s *itemApplyEngageLegacyServer4F2FF0) S() *server.Server {
	return s.srv
}

func TestItemApplyEngageExport4F2FF0InvokesNativeCallbacks(t *testing.T) {
	oldGetServer := GetServer
	GetServer = func() Server {
		return &itemApplyEngageLegacyServer4F2FF0{srv: new(server.Server)}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	item, freeItem := alloc.New(server.Object{})
	defer freeItem()
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	attrs, freeAttrs := alloc.New(server.ModifierInitData{})
	defer freeAttrs()
	brilliance, freeBrilliance := alloc.New(server.ModifierEff{})
	defer freeBrilliance()
	fireProtect, freeFireProtect := alloc.New(server.ModifierEff{})
	defer freeFireProtect()

	brilliance.Engage112 = modifierEngagePointerNative4DFBB0(8)
	fireProtect.Engage112 = modifierEngagePointerNative4DFBB0(1)
	if brilliance.Engage112 == nil || fireProtect.Engage112 == nil {
		t.Fatal("registered native engage callback pointer is nil")
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
			brilliance.Engage112,
			fireProtect.Engage112,
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}

	itemApplyEngageExportCall4F2FF0(item, owner)
	if got := owner.Field110; got != 9 {
		t.Fatalf("owner engage mask = %#x, want Brilliance|FireProtect %#x", got, uint32(9))
	}
	runtime.KeepAlive(item)
	runtime.KeepAlive(owner)
	runtime.KeepAlive(attrs)
	runtime.KeepAlive(brilliance)
	runtime.KeepAlive(fireProtect)
}
