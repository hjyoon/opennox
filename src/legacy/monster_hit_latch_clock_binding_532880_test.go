package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestHitLatchProductionClockBinding532880(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	unit, freeUnit := alloc.New(server.Object{})
	update, freeUpdate := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(freeUnit)
	t.Cleanup(freeUpdate)
	*unit = server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	invoke := func() {
		t.Helper()
		if got := ccall.CallUPtrPtr(hitLatchPublicC532880(), unit.CObj()); got != uintptr(unsafe.Pointer(update)) {
			t.Fatalf("native cached update return=%#x want=%p", got, update)
		}
	}
	t.Run("ordinary live server clock", func(t *testing.T) {
		*update = server.MonsterUpdateData{}
		srv.SetFrame(0x80000004)
		invoke()
		if update.Field130 != 0x80000004 {
			t.Fatal("real live server Frame DWORD was not stored")
		}
	})
	t.Run("zero frame rereads and later latch stays", func(t *testing.T) {
		*update = server.MonsterUpdateData{}
		srv.SetFrame(0)
		invoke()
		if update.Field130 != 0 {
			t.Fatal("zero frame must remain zero")
		}
		srv.SetFrame(117)
		invoke()
		if update.Field130 != 117 {
			t.Fatal("zero timestamp must re-read the next frame")
		}
		srv.SetFrame(999)
		invoke()
		if update.Field130 != 117 {
			t.Fatal("nonzero timestamp must stay latched")
		}
	})
	t.Run("nonzero no class gate or server lookup", func(t *testing.T) {
		*update = server.MonsterUpdateData{Field130: 0xffffffff}
		unit.ObjClass, unit.ObjFlags = 0, object.FlagDead
		oldServer := GetServer
		GetServer = func() Server { panic("nonzero hit latch unexpectedly resolved server clock") }
		defer func() { GetServer = oldServer }()
		invoke()
		if update.Field130 != 0xffffffff {
			t.Fatal("original nonzero timestamp changed")
		}
	})
	t.Log("ordinary actual C hit-latch clock modes=3 calls=5 real-server-frame-and-zero-reread=true nonzero-no-server-no-class-gate=true")
}
