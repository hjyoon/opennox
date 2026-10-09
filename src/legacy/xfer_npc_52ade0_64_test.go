//go:build amd64 || arm64

package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestNPCNormalizeEquipped52BA70(t *testing.T) {
	tests := []struct {
		name       string
		firstClass object.Class
		firstSub   object.SubClass
		lastClass  object.Class
		lastSub    object.SubClass
		wantFirst  bool
		wantLast   bool
	}{
		{
			name:       "two-handed weapon wins over later shield",
			firstClass: object.ClassWeapon,
			firstSub:   0x00000004,
			lastClass:  object.ClassArmor,
			lastSub:    0x00000002,
			wantFirst:  true,
		},
		{
			name:       "shield wins over later two-handed weapon",
			firstClass: object.ClassArmor,
			firstSub:   0x00000002,
			lastClass:  object.ClassWand,
			lastSub:    0x00000400,
			wantFirst:  true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			first := &server.Object{ObjClass: tc.firstClass, ObjSubClass: tc.firstSub, ObjFlags: object.FlagEquipped}
			last := &server.Object{ObjClass: tc.lastClass, ObjSubClass: tc.lastSub, ObjFlags: object.FlagEquipped}
			other := &server.Object{ObjClass: object.ClassFood, ObjFlags: object.FlagEquipped}
			first.InvNextItem = last
			last.InvNextItem = other
			owner := &server.Object{InvFirstItem: first}

			npcNormalizeEquipped52BA70(owner)

			if got := first.ObjFlags.Has(object.FlagEquipped); got != tc.wantFirst {
				t.Fatalf("first equipped = %v, want %v", got, tc.wantFirst)
			}
			if got := last.ObjFlags.Has(object.FlagEquipped); got != tc.wantLast {
				t.Fatalf("last equipped = %v, want %v", got, tc.wantLast)
			}
			if !other.ObjFlags.Has(object.FlagEquipped) {
				t.Fatal("unrelated equipped item was modified")
			}
		})
	}
}

func TestNPCRestoreEquipped52ADE0(t *testing.T) {
	srv := npcRestoreTestServer52ADE0(t)
	oldWeaponFlags := objectNPCWeaponEquipFlags
	oldArmorFlags := objectNPCArmorEquipFlags
	defer func() {
		objectNPCWeaponEquipFlags = oldWeaponFlags
		objectNPCArmorEquipFlags = oldArmorFlags
	}()
	objectNPCWeaponEquipFlags = func(item *server.Object) uint32 {
		if !item.ObjClass.HasAny(object.ClassWeapon | object.ClassWand) {
			t.Fatalf("weapon lookup received class %#x", item.ObjClass)
		}
		return 0x12
	}
	objectNPCArmorEquipFlags = func(item *server.Object) uint32 {
		if !item.ObjClass.Has(object.ClassArmor) {
			t.Fatalf("armor lookup received class %#x", item.ObjClass)
		}
		return 0x34
	}

	ud, freeUD := alloc.New(server.MonsterUpdateData{})
	*ud = server.MonsterUpdateData{
		WeaponEquipFlags: 0x80000000, ArmorEquipFlags: 0x80000000,
		Field516: 0x12345678, Field517: 0x11223344,
	}
	attrs, freeAttrs := alloc.New(server.ModifierInitData{})
	t.Cleanup(freeUD)
	t.Cleanup(freeAttrs)
	weapon := npcRestoreTestObject52ADE0(t)
	weapon.ObjClass, weapon.ObjSubClass, weapon.ObjFlags = object.ClassWeapon, object.SubClass(object.WeaponBow), object.FlagEquipped
	weapon.InitData = unsafe.Pointer(attrs)
	armor := npcRestoreTestObject52ADE0(t)
	armor.ObjClass, armor.ObjSubClass, armor.ObjFlags = object.ClassArmor, object.SubClass(object.ArmorBreastplate), object.FlagEquipped
	armor.TypeInd, armor.InitData = uint16(srv.Types.IndByID("LeatherArmor")), unsafe.Pointer(attrs)
	ignored := npcRestoreTestObject52ADE0(t)
	ignored.ObjClass = object.ClassFood
	weapon.InvNextItem = armor
	armor.InvNextItem = ignored
	owner := npcRestoreTestObject52ADE0(t)
	*owner = server.Object{
		ObjClass:     object.ClassMonster | object.ClassClientPersist,
		ObjSubClass:  0x10,
		InvFirstItem: weapon,
		UpdateData:   unsafe.Pointer(ud),
	}
	weapon.InvHolder, armor.InvHolder, ignored.InvHolder = owner, owner, owner

	npcRestoreEquipped52ADE0(owner)

	if ud.WeaponEquipFlags != 0x80000012 || ud.ArmorEquipFlags != 0x80000034 {
		t.Fatalf("appearance flags = (%#x, %#x), want original equip OR updates (0x80000012, 0x80000034)", ud.WeaponEquipFlags, ud.ArmorEquipFlags)
	}
	if !weapon.ObjFlags.Has(object.FlagEquipped) || !armor.ObjFlags.Has(object.FlagEquipped) {
		t.Fatal("restore changed inventory equipped flags")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && ud.Field516 != 0 || ud.Field517 != 0x11223300 {
		t.Fatalf("weapon restoration did not run native equip: compatibility=%#x animation=%#x", ud.Field516, ud.Field517)
	}
}

func TestNPCApplyDormantHealth52ADE0(t *testing.T) {
	health := &server.HealthData{Cur: 11, Field2: 22, Max: 33}
	npcApplyDormantHealth52ADE0(health)
	if health.Cur != 0 || health.Max != 0 || health.Field2 != 22 {
		t.Fatalf("health = {%d %d %d}, want {0 22 0}", health.Cur, health.Field2, health.Max)
	}
}
