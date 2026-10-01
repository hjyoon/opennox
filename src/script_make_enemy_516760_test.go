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

func TestScriptMakeEnemy516760ClearsSubclassNotObjectFlags(t *testing.T) {
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
	for _, subclass := range []object.SubClass{0, 2, 0x82, 0x182, 0xffffffff} {
		unit.ObjSubClass = subclass
		unit.ObjFlags = object.FlagEnabled | object.FlagActive | object.FlagEquipped
		flags := unit.ObjFlags
		ns.MakeEnemy(ns.toObj(unit))
		if unit.ObjSubClass != subclass&^0x100 || unit.ObjFlags != flags || unit.ObjOwner != nil {
			t.Fatalf("MakeEnemy: subclass=%#x want=%#x flags=%#x want=%#x owner=%p", uint32(unit.ObjSubClass), uint32(subclass&^0x100), uint32(unit.ObjFlags), uint32(flags), unit.ObjOwner)
		}
	}
	// A nil object must not inspect the namespace/server.
	noxScriptNS{}.MakeEnemy(nil)
}

func TestScriptMakeEnemy516760PreservesNativeOwnerRemoval(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(2) {
		t.Fatal("cannot initialize native objects")
	}
	t.Cleanup(s.Objs.FreeObjects)
	owner := s.Objs.NewObject(&server.ObjectType{})
	unit := s.Objs.NewObject(&server.ObjectType{})
	update, freeUpdate := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	unit.ObjClass = object.ClassMonster
	unit.ObjSubClass = object.SubClass(0x182)
	unit.ObjFlags = object.FlagEnabled | object.FlagActive | object.FlagEquipped
	unit.UpdateData = unsafe.Pointer(update)
	s.ObjSetOwner(owner, unit)
	update.CurrentEnemy = owner
	flags := unit.ObjFlags
	ns := s.noxScriptP()
	ns.MakeEnemy(ns.toObj(unit))
	// The non-player owner avoids ClearOwner's separate monitored-player
	// notification branch; MakeEnemy itself clears only Migrate, not Monitor.
	if unit.ObjSubClass != 0x82 || unit.ObjFlags != flags || unit.ObjOwner != nil || owner.FirstOwned516() != nil || update.CurrentEnemy != nil {
		t.Fatalf("MakeEnemy: subclass=%#x flags=%#x owner=%p first-owned=%p enemy=%p", uint32(unit.ObjSubClass), uint32(unit.ObjFlags), unit.ObjOwner, owner.FirstOwned516(), update.CurrentEnemy)
	}
}
