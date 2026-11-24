package legacy

/*
#include "GAME1_1.h"
#include "GAME3_3.h"
#include "GAME4_3.h"
extern uint32_t dword_5d4594_2488720, dword_5d4594_2488724;
*/
import "C"

import (
	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func equipmentPickupRulesAllowed() bool {
	return !noxflags.HasGame(noxflags.GameModeCoop | noxflags.GameModeQuest)
}

//export sub_53A720
func sub_53A720(owner, item *nox_object_t, report, equipReport C.int) C.int {
	return C.int(weaponPickup53A720(asObjectS(owner), asObjectS(item), report, equipReport))
}

func weaponPickup53A720(owner, item *server.Object, report, equipReport C.int) int {
	if owner == nil || item == nil {
		return 0
	}
	world := GetServer().S()
	if owner.ObjClass.Has(object.ClassPlayer) && noxflags.HasGame(noxflags.GameModeQuest) &&
		uint32(item.ObjSubClass)&0x200000 != 0 {
		limit := int32(C.nox_float2int(C.float(nox_xxx_gamedataGetFloat_419D40(internCStr("ForceOfNatureStaffLimit")))))
		if nox_xxx_inventoryCountObjects_4E7D30(asObjectC(owner), int32(item.TypeInd)) >= limit {
			world.NetPriMsgToPlayer(owner, "pickup.c:MaxSameItem", 0)
			world.Audio.EventObj(925, owner, 0, 0)
			return 0
		}
	}
	if equipmentPickupRulesAllowed() && sub_409F40(2) != 0 {
		duplicate := uint32(item.ObjSubClass)&0x82 == 0 && C.sub_4E7EC0(asObjectC(owner), asObjectC(item)) != 0
		if uint32(item.ObjSubClass)&0x40 != 0 {
			for owned := owner.Field129; owned != nil; owned = owned.Field128 {
				if owned.ObjClass.Has(object.ClassWeapon) && uint32(owned.ObjSubClass)&0x40 != 0 {
					duplicate = true
					break
				}
			}
		}
		if duplicate {
			if owner.ObjClass.Has(object.ClassPlayer) {
				rejectEquipment(owner, "weapon.c:CannotPickupDuplicateWeapon", true)
			}
			return 0
		}
	}
	if equipmentPickupRulesAllowed() && owner.ObjClass.Has(object.ClassPlayer) {
		update := owner.UpdateDataPlayer()
		if update != nil && update.Player != nil && !Nox_xxx_playerClassCanUseItem_57B3D0(item, update.Player.PlayerClass()) {
			return rejectEquipment(owner, "weapon.c:WeaponEquipClassFail", true)
		}
	}
	if pickupDefaultCall4F31E0(owner, item, int32(report), int32(equipReport)) != 1 {
		return 0
	}

	if owner.ObjClass.Has(object.ClassPlayer) {
		update := owner.UpdateDataPlayer()
		if update != nil && update.Player != nil {
			equipped := 0
			if update.EquippedWeapon == nil && sub_419E60(asObjectC(owner)) == 0 &&
				nox_xxx_weaponInventoryEquipFlags_415820(asObjectC(item)) != 2 {
				equipped = playerEquipWeapon53A420(owner, item, equipReport, false)
			}
			if sub_419E60(asObjectC(owner)) == 0 && nox_xxx_weaponInventoryEquipFlags_415820(asObjectC(item)) == 2 &&
				update.Player.WeaponEquip&0xC != 0 && update.Player.WeaponEquip&2 == 0 {
				equipped = playerEquipWeapon53A420(owner, item, equipReport, false)
			}
			if equipped == 0 {
				reportPickedWeaponCharges(update.Player.PlayerInd, item)
			}
		}
	}
	world.WeaponPickupSound53A6C0(owner, item)
	world.DecayRemove5116F0(item)
	return 1
}

func reportPickedWeaponCharges(player byte, item *server.Object) {
	data := item.UseData.Ptr
	if data == nil {
		return
	}
	switch {
	case item.ObjClass.Has(object.ClassWand) && uint32(item.ObjSubClass)&0x47f0000 != 0:
		wand := (*server.WandUseData)(data)
		Nox_xxx_netReportCharges_4D82B0(player, item, wand.Charge, wand.MaxCharge)
	case item.ObjClass.Has(object.ClassWeapon) && uint32(item.ObjSubClass)&0x82 != 0:
		ammo := (*server.AmmoUseData)(data)
		Nox_xxx_netReportCharges_4D82B0(player, item, ammo.Charge1, ammo.Charge0)
	}
}

const armorPickupLegacyBase = 0x5D4594

//export nox_xxx_pickupArmor_53E7F0
func nox_xxx_pickupArmor_53E7F0(owner, item *nox_object_t, report, equipReport C.int) C.int {
	return C.int(armorPickup53E7F0(asObjectS(owner), asObjectS(item), report, equipReport))
}

func armorPickup53E7F0(owner, item *server.Object, report, equipReport C.int) int {
	world := GetServer().S()
	shoes := memmap.PtrUint32(armorPickupLegacyBase, 2488712)
	robe := memmap.PtrUint32(armorPickupLegacyBase, 2488716)
	if *shoes == 0 {
		*shoes = uint32(world.Types.IndByID("StreetSneakers"))
		*robe = uint32(world.Types.IndByID("WizardRobe"))
		C.dword_5d4594_2488720 = C.uint32_t(world.Types.IndByID("WoodenShield"))
		C.dword_5d4594_2488724 = C.uint32_t(world.Types.IndByID("SteelShield"))
	}
	if equipmentPickupRulesAllowed() && sub_409F40(2) != 0 && C.sub_4E7EC0(asObjectC(owner), asObjectC(item)) != 0 {
		return rejectEquipment(owner, "armor.c:CannotPickupDuplicateArmor", true)
	}
	if equipmentPickupRulesAllowed() && owner.ObjClass.Has(object.ClassPlayer) {
		update := owner.UpdateDataPlayer()
		if update != nil && update.Player != nil && !Nox_xxx_playerClassCanUseItem_57B3D0(item, update.Player.PlayerClass()) {
			return rejectEquipment(owner, "armor.c:ArmorEquipClassFail", true)
		}
	}
	if pickupDefaultCall4F31E0(owner, item, int32(report), int32(equipReport)) != 1 {
		return 0
	}
	if owner.ObjClass.Has(object.ClassPlayer) {
		equipped := objectFieldSameArmor53E7B0(owner, item)
		update := owner.UpdateDataPlayer()
		if sub_419E60(asObjectC(owner)) == 0 && shouldEquipPickedArmor53E7F0(item, equipped, update) {
			playerArmorEquip53E650(owner, item, equipReport, false)
		}
	}
	world.ArmorPickupSound53E7F0(owner, item)
	world.DecayRemove5116F0(item)
	return 1
}

func armorHasModifiers(item *server.Object) bool {
	if item.InitData == nil {
		return false
	}
	for _, modifier := range item.InitDataModifier().Modifiers {
		if modifier != nil {
			return true
		}
	}
	return false
}

func shouldEquipPickedArmor53E7F0(item, equipped *server.Object, update *server.PlayerUpdateData) bool {
	wooden := uint32(C.dword_5d4594_2488720)
	steel := uint32(C.dword_5d4594_2488724)
	robe := memmap.Uint32(armorPickupLegacyBase, 2488716)
	shoes := memmap.Uint32(armorPickupLegacyBase, 2488712)
	if equipped != nil {
		switch uint32(equipped.TypeInd) {
		case robe:
			return !armorHasModifiers(equipped)
		case wooden, steel, shoes:
		default:
			return false
		}
	}
	if uint32(item.ObjSubClass)&2 == 0 {
		return true
	}
	if equipped == nil {
		return update == nil || update.EquippedWeapon == nil ||
			uint32(update.EquippedWeapon.ObjSubClass)&shieldConflictingWeaponMask == 0
	}
	switch uint32(equipped.TypeInd) {
	case wooden:
		return !armorHasModifiers(equipped)
	case steel:
		switch uint32(item.TypeInd) {
		case wooden:
			return !armorHasModifiers(equipped) && armorHasModifiers(item)
		case steel:
			return !armorHasModifiers(equipped)
		}
	}
	return false
}
