package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type inventoryPutLegacyServer4F3070 struct {
	Server
	srv *server.Server
}

func (s *inventoryPutLegacyServer4F3070) S() *server.Server {
	return s.srv
}

func TestInventoryPutExport4F3070PreservesNativePointers(t *testing.T) {
	oldGetServer := GetServer
	GetServer = func() Server {
		return &inventoryPutLegacyServer4F3070{srv: new(server.Server)}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	item, freeItem := alloc.New(server.Object{})
	defer freeItem()
	oldHead, freeOldHead := alloc.New(server.Object{})
	defer freeOldHead()
	oldPrevious, freeOldPrevious := alloc.New(server.Object{})
	defer freeOldPrevious()

	owner.InvFirstItem = oldHead
	item.Field125 = oldPrevious
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{
			unsafe.Pointer(owner),
			unsafe.Pointer(item),
			unsafe.Pointer(oldHead),
			unsafe.Pointer(oldPrevious),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}

	inventoryPutExportCall4F3070(owner, item, -1)
	if owner.InvFirstItem != item || item.Field125 != nil || item.InvNextItem != oldHead {
		t.Fatalf("inventory links = head %p previous %p next %p", owner.InvFirstItem, item.Field125, item.InvNextItem)
	}
	if oldHead.Field125 != item || item.InvHolder != owner || item.ObjOwner != owner {
		t.Fatalf("back-links = old previous %p holder %p owner %p", oldHead.Field125, item.InvHolder, item.ObjOwner)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
	runtime.KeepAlive(oldHead)
	runtime.KeepAlive(oldPrevious)
}
