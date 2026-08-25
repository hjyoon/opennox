package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerObserverFindGoodSlave2NativeLayouts4EC3E0(t *testing.T) {
	type layout struct {
		objectSize      uintptr
		class           uintptr
		nextOwned       uintptr
		firstOwned      uintptr
		updateData      uintptr
		monsterDataSize uintptr
		monsterStatus   uintptr
	}
	var want layout
	switch unsafe.Sizeof(uintptr(0)) {
	case 4:
		want = layout{780, 8, 512, 516, 748, 2200, 1440}
	case 8:
		want = layout{928, 12, 560, 568, 872, 2896, 2116}
	default:
		t.Fatalf("unsupported pointer width %d", unsafe.Sizeof(uintptr(0)))
	}
	got := layout{
		objectSize:      unsafe.Sizeof(server.Object{}),
		class:           unsafe.Offsetof(server.Object{}.ObjClass),
		nextOwned:       unsafe.Offsetof(server.Object{}.Field128),
		firstOwned:      unsafe.Offsetof(server.Object{}.Field129),
		updateData:      unsafe.Offsetof(server.Object{}.UpdateData),
		monsterDataSize: unsafe.Sizeof(server.MonsterUpdateData{}),
		monsterStatus:   unsafe.Offsetof(server.MonsterUpdateData{}.StatusFlags),
	}
	if got != want {
		t.Fatalf("native layout = %+v, want %+v", got, want)
	}
}

func TestPlayerObserverFindGoodSlave2Native4EC3E0(t *testing.T) {
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
	}
	ignored := &server.Object{}
	owner.Field129 = nonMonster
	nonMonster.Field128 = dormant
	dormant.Field128 = summoned
	summoned.Field128 = ignored

	if got := playerObserverFindGoodSlave2_4EC3E0(owner); got != summoned {
		t.Fatalf("result = %p, want summoned monster %p", got, summoned)
	}
	if got := playerObserverFindGoodSlave2_4EC3E0(nil); got != nil {
		t.Fatalf("nil owner result = %p", got)
	}
}

func TestPlayerObserverFindGoodSlave2Native4EC3E0NilMonsterDataFaults(t *testing.T) {
	owner := &server.Object{Field129: &server.Object{ObjClass: object.ClassMonster}}
	defer func() {
		if recover() == nil {
			t.Fatal("nil Monster update data returned without fault")
		}
	}()
	playerObserverFindGoodSlave2_4EC3E0(owner)
}
