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

// Exercise the ordinary C equip boundary and actual dequip writer. A later
// inventory/class change must not replace the pointer in an older cached MUD.
func TestNPCWeaponMeleeCache53A2C0OrdinaryWritersAndOwnerLifetime(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	if !srv.Objs.Init(1) {
		t.Fatal("native owner pool")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	owner := srv.Objs.NewObject(&server.ObjectType{})
	data0, free0 := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(free0)
	data1, free1 := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(free1)
	weapon, freeWeapon := alloc.New(server.Object{})
	t.Cleanup(freeWeapon)
	other, freeOther := alloc.New(server.Object{})
	t.Cleanup(freeOther)
	attrs, freeAttrs := alloc.New(server.ModifierInitData{})
	t.Cleanup(freeAttrs)
	for _, p := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(data0), unsafe.Pointer(data1), unsafe.Pointer(weapon), unsafe.Pointer(other), unsafe.Pointer(attrs)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatal("ordinary binding must retain native pointers above4GiB")
		}
	}
	oldServer, oldLookup := GetServer, objectNPCWeaponEquipFlags
	GetServer = func() Server { return &playerAttackLegacyServer538960{srv: srv} }
	objectNPCWeaponEquipFlags = func(*server.Object) uint32 { return uint32(object.WeaponBow) }
	t.Cleanup(func() { GetServer, objectNPCWeaponEquipFlags = oldServer, oldLookup })
	owner.ObjClass = object.ClassMonster
	owner.ObjSubClass = object.SubClass(object.MonsterNPC)
	owner.UpdateData = unsafe.Pointer(data0)
	owner.InvFirstItem = weapon
	*weapon = server.Object{TypeInd: 0x3401, ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(object.WeaponBow), InvHolder: owner, InitData: unsafe.Pointer(attrs)}
	data0.Field517 = 0x11223344
	data1.Field517 = 0x55667788
	if got := playerEquipWeaponNativeCall53A420(owner, weapon); got != 1 {
		t.Fatalf("ordinary public C equip=%d", got)
	}
	if got := owner.NPCWeaponForMeleeUpdate532130(data0); got != weapon {
		t.Fatalf("ordinary equip published native pointer=%p want=%p", got, weapon)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && data0.Field516 != 0 {
		t.Fatal("native pointer truncated into the PE32 compatibility word")
	}
	owner.UpdateData = unsafe.Pointer(data1)
	owner.ObjClass = 0
	owner.InvFirstItem = other
	*other = server.Object{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(0x4000)}
	if got := owner.NPCWeaponForMeleeUpdate532130(data0); got != weapon {
		t.Fatal("live class/inventory/update replaced the cached native pointer")
	}
	owner.ObjClass = object.ClassMonster
	owner.InvFirstItem = weapon
	if got := npcWeaponDequipNative53A030(owner, weapon); got != 1 {
		t.Fatalf("ordinary dequip=%d", got)
	}
	if owner.NPCWeaponForMeleeUpdate532130(data1) != nil {
		t.Fatal("dequip did not publish known nil for its exact live record")
	}
	if owner.NPCWeaponForMeleeUpdate532130(data0) != weapon {
		t.Fatal("dequip erased another cached record's pointer")
	}
	other.ObjClass = object.ClassWeapon
	other.ObjFlags = object.FlagEquipped
	other.InvHolder = owner
	owner.InvFirstItem = other
	if owner.NPCWeaponForMeleeUpdate532130(data1) != nil {
		t.Fatal("known nil was replaced by reconstructed inventory")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (data0.Field516 != 0 || data1.Field516 != 0) {
		t.Fatal("compatibility words widened or overlapped")
	}
	if data0.Field517 != 0x11223300 || data1.Field517 != 0x55667700 {
		t.Fatal("adjacent animation word changed beyond the original low-byte store")
	}
	owner.InvFirstItem = nil
	srv.Objs.FreeObject(owner)
	reused := srv.Objs.NewObject(&server.ObjectType{})
	if reused != owner {
		t.Fatal("single-object pool did not exercise owner-address reuse")
	}
	reused.ObjClass = object.ClassMonster
	reused.ObjSubClass = object.SubClass(object.MonsterNPC)
	reused.UpdateData = unsafe.Pointer(data0)
	data0.Field516 = 0
	if reused.NPCWeaponForMeleeUpdate532130(data0) != nil {
		t.Fatal("freed owner's cached pointer leaked into its reused address")
	}
}
