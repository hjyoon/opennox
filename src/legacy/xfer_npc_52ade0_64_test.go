//go:build amd64 || arm64

package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

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

	ud := &server.MonsterUpdateData{WeaponEquipFlags: 0xffffffff, ArmorEquipFlags: 0xffffffff}
	weapon := &server.Object{ObjClass: object.ClassWeapon, ObjFlags: object.FlagEquipped}
	armor := &server.Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped}
	ignored := &server.Object{ObjClass: object.ClassFood}
	weapon.InvNextItem = armor
	armor.InvNextItem = ignored
	owner := &server.Object{
		ObjClass:     object.ClassMonster | object.ClassClientPersist,
		ObjSubClass:  0x10,
		InvFirstItem: weapon,
		UpdateData:   unsafe.Pointer(ud),
	}

	npcRestoreEquipped52ADE0(owner)

	if ud.WeaponEquipFlags != 0x12 || ud.ArmorEquipFlags != 0x34 {
		t.Fatalf("appearance flags = (%#x, %#x), want (0x12, 0x34)", ud.WeaponEquipFlags, ud.ArmorEquipFlags)
	}
	if !weapon.ObjFlags.Has(object.FlagEquipped) || !armor.ObjFlags.Has(object.FlagEquipped) {
		t.Fatal("restore changed inventory equipped flags")
	}
}
