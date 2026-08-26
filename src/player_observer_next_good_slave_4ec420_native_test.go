package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerObserverFindGoodSlaveNativeLayouts4EC420(t *testing.T) {
	type layout struct {
		objectSize      uintptr
		class           uintptr
		owner           uintptr
		nextOwned       uintptr
		updateData      uintptr
		monsterDataSize uintptr
		monsterStatus   uintptr
	}
	var want layout
	switch unsafe.Sizeof(uintptr(0)) {
	case 4:
		want = layout{780, 8, 508, 512, 748, 2200, 1440}
	case 8:
		want = layout{928, 12, 552, 560, 872, 2960, 2180}
	default:
		t.Fatalf("unsupported pointer width %d", unsafe.Sizeof(uintptr(0)))
	}
	got := layout{
		objectSize:      unsafe.Sizeof(server.Object{}),
		class:           unsafe.Offsetof(server.Object{}.ObjClass),
		owner:           unsafe.Offsetof(server.Object{}.ObjOwner),
		nextOwned:       unsafe.Offsetof(server.Object{}.Field128),
		updateData:      unsafe.Offsetof(server.Object{}.UpdateData),
		monsterDataSize: unsafe.Sizeof(server.MonsterUpdateData{}),
		monsterStatus:   unsafe.Offsetof(server.MonsterUpdateData{}.StatusFlags),
	}
	if got != want {
		t.Fatalf("native layout = %+v, want %+v", got, want)
	}
}

func TestPlayerObserverFindGoodSlaveNative4EC420(t *testing.T) {
	owner := &server.Object{}
	nonMonster := &server.Object{ObjClass: object.Class(0xfffffffc)}
	dormantData := &server.MonsterUpdateData{StatusFlags: object.MonsterStatus(0xffffff7f)}
	dormant := &server.Object{
		ObjClass:   object.Class(0xffffff82),
		UpdateData: unsafe.Pointer(dormantData),
	}
	summonedData := &server.MonsterUpdateData{StatusFlags: object.MonsterStatus(0xffffff80)}
	summoned := &server.Object{
		ObjClass:   object.ClassMonster,
		UpdateData: unsafe.Pointer(summonedData),
		ObjOwner:   &server.Object{},
	}
	ignored := &server.Object{}
	current := &server.Object{ObjOwner: owner, Field128: nonMonster}
	nonMonster.Field128 = dormant
	dormant.Field128 = summoned
	summoned.Field128 = ignored

	if got := playerObserverFindGoodSlave_4EC420(current); got != summoned {
		t.Fatalf("result = %p, want summoned monster %p", got, summoned)
	}
	if got := playerObserverFindGoodSlave_4EC420(nil); got != nil {
		t.Fatalf("nil current result = %p", got)
	}
	current.ObjOwner = nil
	if got := playerObserverFindGoodSlave_4EC420(current); got != nil {
		t.Fatalf("ownerless current result = %p", got)
	}
}

func TestPlayerObserverFindGoodSlaveNative4EC420NilMonsterDataFaults(t *testing.T) {
	current := &server.Object{
		ObjOwner: &server.Object{},
		Field128: &server.Object{ObjClass: object.ClassMonster},
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil Monster update data returned without fault")
		}
	}()
	playerObserverFindGoodSlave_4EC420(current)
}
