package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type playerTryDequipServer struct {
	Server
	srv *server.Server
}

func (s *playerTryDequipServer) S() *server.Server { return s.srv }

func TestPlayerTryDequip4F2FB0PublicCEntry(t *testing.T) {
	srv := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	bridge := &playerTryDequipServer{srv: srv}
	oldGetServer := GetServer
	GetServer = func() Server { return bridge }
	t.Cleanup(func() { GetServer = oldGetServer })

	owner, freeOwner := alloc.NewHigh(server.Object{})
	t.Cleanup(freeOwner)
	item, freeItem := alloc.NewHigh(server.Object{})
	t.Cleanup(freeItem)
	update, freeUpdate := alloc.NewHigh(server.MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	initData, freeInit := alloc.NewHigh(server.ModifierInitData{})
	t.Cleanup(freeInit)
	owner.ObjClass = object.ClassMonster
	item.ObjClass, item.ObjSubClass, item.ObjFlags = object.ClassWeapon, object.SubClass(0x100), object.FlagEquipped
	update.Field517 = 0xcafe1234
	owner.UpdateData = unsafe.Pointer(update)
	owner.InvFirstItem, item.InvHolder = item, owner
	item.InitData = unsafe.Pointer(initData)

	if unsafe.Sizeof(uintptr(0)) > 4 &&
		(uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 || uintptr(unsafe.Pointer(item)) <= math.MaxUint32) {
		t.Fatal("object pointers must be above 4 GiB")
	}
	if got := playerTryDequipExportCall4F2FB0(owner, item); got != 1 {
		t.Fatalf("C dequip returned %d, want 1", got)
	}
	if item.Flags().Has(object.FlagEquipped) || owner.InvFirstItem != item || update.Field517 != 0xcafe1200 {
		t.Fatalf("native item state after C dequip: flags=%v item=%p update=%#x",
			item.Flags(), owner.InvFirstItem, update.Field517)
	}
	if got := playerTryDequipExportCall4F2FB0(owner, item); got != 0 {
		t.Fatalf("second C dequip returned %d, want 0", got)
	}
}
