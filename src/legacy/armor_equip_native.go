package legacy

/*
#include "GAME4_3.h"
extern uint64_t qword_581450_9512;
*/
import "C"

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_recalculateArmorVal_53E300
func nox_xxx_recalculateArmorVal_53E300(owner *nox_object_t) C.int {
	recalculateEquipmentArmor53E300(asObjectS(owner))
	return 0
}

func recalculateEquipmentArmor53E300(owner *server.Object) {
	cap := math.Float64frombits(uint64(C.qword_581450_9512))
	value := server.EquippedArmorValue53E300(owner, itemDefendNative415C00, cap)
	npcSetArmorValueNative53E300(owner, value)
}

//export sub_53E3A0
func sub_53E3A0(owner, item *nox_object_t) C.int {
	return C.int(npcArmorDequip53E3A0(asObjectS(owner), asObjectS(item), recalculateEquipmentArmor53E300))
}

//export sub_53E430
func sub_53E430(owner, item *nox_object_t, report, broadcast C.int) C.int {
	return C.int(playerArmorDequip53E430(asObjectS(owner), asObjectS(item), report, broadcast))
}

func playerArmorDequip53E430(owner, item *server.Object, report, broadcast C.int) int {
	if !item.ObjClass.Has(object.ClassArmor) || !item.Flags().Has(object.FlagEquipped) {
		return 0
	}
	if owner.ObjClass.Has(object.ClassMonster) {
		return npcArmorDequip53E3A0(owner, item, recalculateEquipmentArmor53E300)
	}
	if !owner.ObjClass.Has(object.ClassPlayer) || !inventoryHasEquipment(owner, item) {
		return 0
	}
	update := owner.UpdateDataPlayer()
	if update == nil || update.Player == nil {
		return 0
	}
	item.ObjFlags &^= object.FlagNoAutoDrop | object.FlagEquipped
	update.Player.ArmorEquip &^= uint32(nox_xxx_unitArmorInventoryEquipFlags_415C70(asObjectC(item)))
	reportEquipmentRemoval(update, item, item, report, broadcast)
	recalculateEquipmentArmor53E300(owner)
	Nox_xxx_itemApplyDisengageEffect_4F3030(item, owner)
	if uint32(item.ObjSubClass)&2 != 0 && update.State >= server.PlayerState15 && update.State <= server.PlayerState17 {
		Nox_xxx_playerSetState_4FA020(owner, server.PlayerState13)
	}
	return 1
}

//export nox_xxx_NPCEquipArmor_53E520
func nox_xxx_NPCEquipArmor_53E520(owner, item *nox_object_t) C.int {
	return C.int(npcArmorEquip53E520(asObjectS(owner), asObjectS(item), recalculateEquipmentArmor53E300, dequipShieldConflictingWeapons53E600))
}

//export sub_53E600
func sub_53E600(owner *nox_object_t) {
	dequipShieldConflictingWeapons53E600(asObjectS(owner))
}

func dequipShieldConflictingWeapons53E600(owner *server.Object) {
	if owner == nil {
		return
	}
	for item := owner.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.Flags().Has(object.FlagEquipped) && item.ObjClass.HasAny(equipmentWeaponClasses) &&
			uint32(nox_xxx_weaponInventoryEquipFlags_415820(asObjectC(item)))&shieldConflictingWeaponMask != 0 {
			playerDequipWeapon53A140(owner, item, 1, 1)
		}
	}
}

//export nox_xxx_playerEquipArmor_53E650
func nox_xxx_playerEquipArmor_53E650(owner, item *nox_object_t, report, playSound C.int) C.int {
	return C.int(playerArmorEquip53E650(asObjectS(owner), asObjectS(item), report, playSound != 0))
}

func playerArmorEquip53E650(owner, item *server.Object, report C.int, playSound bool) int {
	if !item.ObjClass.Has(object.ClassArmor) || item.Flags().Has(object.FlagEquipped) {
		return 0
	}
	if owner.ObjClass.Has(object.ClassMonster) {
		return npcArmorEquip53E520(owner, item, recalculateEquipmentArmor53E300, dequipShieldConflictingWeapons53E600)
	}
	if !owner.ObjClass.Has(object.ClassPlayer) {
		return 0
	}
	update := owner.UpdateDataPlayer()
	if update == nil || update.Player == nil {
		return 0
	}
	equipped := objectFieldSameArmor53E7B0(owner, item)
	if !Nox_xxx_playerClassCanUseItem_57B3D0(item, update.Player.PlayerClass()) {
		return rejectEquipment(owner, "armor.c:ArmorEquipClassFail", playSound)
	}
	if playerCheckStrengthCall4F3180(owner, item) == 0 {
		return rejectEquipment(owner, "armor.c:ArmorEquipStrengthFail", playSound)
	}
	if equipped != nil {
		playerArmorDequip53E430(owner, equipped, 1, 1)
	}
	item.ObjFlags |= object.FlagEquipped
	player := update.Player
	flags := uint32(nox_xxx_unitArmorInventoryEquipFlags_415C70(asObjectC(item)))
	player.ArmorEquip |= flags
	nox_xxx_netReportEquip_4D8540(C.int(update.Player.PlayerInd), asObjectC(item), report)
	if sub_53E2D0(asObjectC(item)) == 0 {
		item.ObjFlags |= object.FlagNoAutoDrop
	}
	recalculateEquipmentArmor53E300(owner)
	Nox_xxx_itemApplyEngageEffect_4F2FF0(item, owner)
	if uint32(item.ObjSubClass)&2 != 0 {
		dequipShieldConflictingWeapons53E600(owner)
	}
	return 1
}
