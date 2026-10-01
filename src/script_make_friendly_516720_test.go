package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestScriptMakeFriendly516720WritesSubclassNotObjectFlags(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(1) {
		t.Fatal("cannot initialize native objects")
	}
	t.Cleanup(s.Objs.FreeObjects)
	unit := s.Objs.NewObject(&server.ObjectType{})
	ns := s.noxScriptP()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("native object is not above 4 GiB: %p", unit)
	}
	for _, subclass := range []object.SubClass{0, 2, 0x82, 0x182, 0xfffffeff} {
		unit.ObjSubClass = subclass
		unit.ObjFlags = object.FlagEnabled | object.FlagActive | object.FlagNoCollide
		flags := unit.ObjFlags
		ns.MakeFriendly(ns.toObj(unit))
		if unit.ObjSubClass != subclass|0x100 || unit.ObjFlags != flags || unit.ObjOwner != nil {
			t.Fatalf("MakeFriendly: subclass=%#x want=%#x flags=%#x want=%#x owner=%p", uint32(unit.ObjSubClass), uint32(subclass|0x100), uint32(unit.ObjFlags), uint32(flags), unit.ObjOwner)
		}
	}
	// A nil object must not inspect the namespace/server.
	noxScriptNS{}.MakeFriendly(nil)
}

func TestScriptMakeFriendly516720PreservesNativeOwnerAssignment(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(2) {
		t.Fatal("cannot initialize native objects")
	}
	t.Cleanup(s.Objs.FreeObjects)
	host := s.Objs.NewObject(&server.ObjectType{})
	unit := s.Objs.NewObject(&server.ObjectType{})
	update, freeUpdate := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	unit.ObjClass = object.ClassMonster
	unit.ObjSubClass = object.SubClass(0x82)
	unit.ObjFlags = object.FlagEnabled | object.FlagActive
	unit.UpdateData = unsafe.Pointer(update)
	update.CurrentEnemy = host
	s.Players.SetHost(nil, host)
	flags := unit.ObjFlags
	ns := s.noxScriptP()
	ns.MakeFriendly(ns.toObj(unit))
	if unit.ObjSubClass != 0x182 || unit.ObjFlags != flags || unit.ObjOwner != host || host.FirstOwned516() != unit || update.CurrentEnemy != nil {
		t.Fatalf("MakeFriendly: subclass=%#x flags=%#x owner=%p first-owned=%p enemy=%p", uint32(unit.ObjSubClass), uint32(unit.ObjFlags), unit.ObjOwner, host.FirstOwned516(), update.CurrentEnemy)
	}
}
