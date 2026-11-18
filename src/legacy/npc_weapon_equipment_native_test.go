package legacy

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func TestNPCWeaponEquipmentPublicCABI(t *testing.T) {
	s := playerVitalReportWorld(t)
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	for _, name := range []string{"Bow", "Quiver", "WoodenShield", "SteelShield", "LeatherArmor"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	s.Nox_xxx_equipWeapon_4157C0()
	s.Nox_xxx_equipArmor_415AB0()

	for route := range 2 {
		t.Run(fmt.Sprintf("route%d", route), func(t *testing.T) {
			update := game4ConsumerValue(t, server.MonsterUpdateData{Field517: 0x85667788})
			owner := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)})
			attrs := game4ConsumerValue(t, server.ModifierInitData{})
			newItem := func(name string, class object.Class, subclass object.SubClass, flags object.Flags) *server.Object {
				ind := s.Types.IndByID(name)
				if ind == 0 {
					t.Fatalf("missing native type %s", name)
				}
				return game4ConsumerValue(t, server.Object{TypeInd: uint16(ind), ObjClass: class, ObjSubClass: subclass,
					ObjFlags: flags, InvHolder: owner, InitData: unsafe.Pointer(attrs)})
			}
			first := newItem("Quiver", object.ClassWeapon, 2, object.FlagEquipped)
			second := newItem("Quiver", object.ClassWeapon, 2, object.FlagEquipped)
			leather := newItem("LeatherArmor", object.ClassArmor, 0, object.FlagEquipped)
			wood := newItem("WoodenShield", object.ClassArmor, 0, object.FlagEquipped)
			steel := newItem("SteelShield", object.ClassArmor, 0, object.FlagEquipped)
			bow := newItem("Bow", object.ClassWeapon, 4, 0)
			owner.InvFirstItem = first
			first.InvNextItem, second.InvNextItem, leather.InvNextItem = second, leather, wood
			wood.InvNextItem, steel.InvNextItem = steel, bow

			if got := callNPCDequipWeaponNative53A030(route, owner, bow); got != 0 {
				t.Fatalf("dequip on an unequipped bow returned %d", got)
			}
			if got := callNPCEquipWeaponNative53A2C0(route, owner, bow); got != 1 || !bow.Flags().Has(object.FlagEquipped) {
				t.Fatalf("bow equip returned %d and flags %v", got, bow.Flags())
			}
			wantWeaponWord := uint32(0)
			if cgoABIPointerSize == 4 {
				wantWeaponWord = uint32(uintptr(unsafe.Pointer(bow)))
			}
			if update.Field517 != 0x85667700 || update.Field516 != wantWeaponWord {
				t.Fatalf("bow equip animation=%#x weapon word=%#x", update.Field517, update.Field516)
			}
			if wood.Flags().Has(object.FlagEquipped) || steel.Flags().Has(object.FlagEquipped) || !leather.Flags().Has(object.FlagEquipped) ||
				!first.Flags().Has(object.FlagEquipped) || !second.Flags().Has(object.FlagEquipped) {
				t.Fatal("bow equip must dequip shields and retain quivers and other armor")
			}
			if got := callNPCDequipWeaponNative53A030(route, owner, bow); got != 1 || bow.Flags().Has(object.FlagEquipped) ||
				first.Flags().Has(object.FlagEquipped) || !second.Flags().Has(object.FlagEquipped) || update.Field516 != 0 {
				t.Fatalf("bow dequip did not clear the weapon and the first quiver, return=%d", got)
			}

			callDequipQuiverNative53A0F0(route, owner, -7, 13)
			if second.Flags().Has(object.FlagEquipped) {
				t.Fatal("standalone quiver dequip did not select the remaining equipped quiver")
			}
			wood.ObjFlags |= object.FlagEquipped
			steel.ObjFlags |= object.FlagEquipped
			callDequipShieldsNative53A3D0(route, owner)
			if wood.Flags().Has(object.FlagEquipped) || steel.Flags().Has(object.FlagEquipped) || !leather.Flags().Has(object.FlagEquipped) {
				t.Fatal("standalone shield dequip must clear both shields and leave other armor equipped")
			}
			if owner.InvFirstItem != first || steel.InvNextItem != bow || bow.InvHolder != owner {
				t.Fatal("equipment calls changed native inventory ownership")
			}
		})
	}
}
