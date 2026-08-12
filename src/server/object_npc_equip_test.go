package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestObjectSetNPCItemEquipFlagsMatchesGAMEEXEContract(t *testing.T) {
	const (
		weaponFlags = uint32(0x00F00F00)
		armorFlags  = uint32(0x0F00F00F)
	)
	tests := []struct {
		name        string
		targetClass object.Class
		itemClass   object.Class
		equipped    bool
		wantLookup  string
	}{
		{name: "equip weapon for client-persist object", targetClass: object.ClassMonster | object.ClassClientPersist, itemClass: object.ClassWeapon, equipped: true, wantLookup: "weapon"},
		{name: "unequip weapon for immobile object", targetClass: object.ClassMonster | object.ClassImmobile, itemClass: object.ClassWeapon, wantLookup: "weapon"},
		{name: "equip wand for player-class monster", targetClass: object.ClassMonster | object.ClassPlayer, itemClass: object.ClassWand, equipped: true, wantLookup: "weapon"},
		{name: "unequip armor for ordinary NPC", targetClass: object.ClassMonster, itemClass: object.ClassArmor, wantLookup: "armor"},
		{name: "non-weapon class uses armor lookup", targetClass: object.ClassMonster, itemClass: object.ClassFlag, equipped: true, wantLookup: "armor"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ud := &MonsterUpdateData{
				Field512:         0x89ABCDEF,
				Field513:         0x13579BDF,
				WeaponEquipFlags: 0xAA55AA55,
				ArmorEquipFlags:  0x55AA55AA,
				Field516:         0x2468ACE0,
				Field517:         0x10203040,
			}
			obj := &Object{
				ObjClass:   tc.targetClass,
				Field37:    0x80000015,
				Field38:    0x11223344,
				UpdateData: unsafe.Pointer(ud),
			}
			for i := range obj.Field140 {
				obj.Field140[i] = 0xA1000000 | uint32(i<<12) | uint32(i&3)
				if i%2 != 0 {
					obj.Field140[i] |= 0x04000000
				}
				if i%3 == 0 {
					obj.Field140[i] |= 0x400
				}
			}
			item := &Object{ObjClass: tc.itemClass, TypeInd: 0x2468, Field37: 0xCAFEBABE}
			wantWeapon := ud.WeaponEquipFlags
			wantArmor := ud.ArmorEquipFlags
			if tc.wantLookup == "weapon" {
				if tc.equipped {
					wantWeapon |= weaponFlags
				} else {
					wantWeapon &^= weaponFlags
				}
			} else if tc.equipped {
				wantArmor |= armorFlags
			} else {
				wantArmor &^= armorFlags
			}
			wantSlots := obj.Field140
			applyGAMEEXENPCEquipSlots(obj, wantSlots[:])
			var calls []string
			lookup := func(name string, flags uint32) func(*Object) uint32 {
				return func(gotItem *Object) uint32 {
					if gotItem != item {
						t.Errorf("%s lookup item: got %p, want %p", name, gotItem, item)
					}
					if obj.Field38 != math.MaxUint32 {
						t.Errorf("%s lookup ran before NeedSync: Field38 = %#08x", name, obj.Field38)
					}
					calls = append(calls, name)
					return flags
				}
			}

			obj.SetNPCItemEquipFlags(item, tc.equipped, lookup("weapon", weaponFlags), lookup("armor", armorFlags))

			if len(calls) != 1 || calls[0] != tc.wantLookup {
				t.Errorf("lookups: got %v, want [%s]", calls, tc.wantLookup)
			}
			if ud.WeaponEquipFlags != wantWeapon || ud.ArmorEquipFlags != wantArmor {
				t.Errorf("equip flags: got (%#08x, %#08x), want (%#08x, %#08x)", ud.WeaponEquipFlags, ud.ArmorEquipFlags, wantWeapon, wantArmor)
			}
			if obj.Field38 != math.MaxUint32 || obj.Field140 != wantSlots {
				t.Errorf("sync state differs: Field38=%#08x slots=%#v", obj.Field38, obj.Field140)
			}
			if ud.Field512 != 0x89ABCDEF || ud.Field513 != 0x13579BDF || ud.Field516 != 0x2468ACE0 || ud.Field517 != 0x10203040 || obj.UpdateData != unsafe.Pointer(ud) {
				t.Fatal("NPC equipment update modified adjacent update data")
			}
			if item.ObjClass != tc.itemClass || item.TypeInd != 0x2468 || item.Field37 != 0xCAFEBABE {
				t.Fatal("NPC equipment update modified the item")
			}
		})
	}
}

func applyGAMEEXENPCEquipSlots(obj *Object, slots []uint32) {
	if obj.Class().HasAny(object.ClassPlayer | object.ClassImmobile | object.ClassClientPersist) {
		for i := range slots {
			slots[i] = slots[i]&0xFFFFF000 | 0x04000000
		}
		return
	}
	for i := range slots {
		slots[i] |= 0x400
		if obj.Field37&(uint32(1)<<i) != 0 {
			slots[i] |= 0x04000000
		} else if slots[i]&0x400 == 0 {
			slots[i] &^= 0x04000000
		}
	}
}
