package legacy

/*
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_3.h"
#include "player_try_equip_4f2f70.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/server"
)

const shieldConflictingWeaponMask = 0x7ffe40c

func inventoryHasEquipment(owner, item *server.Object) bool {
	for candidate := owner.InvFirstItem; candidate != nil; candidate = candidate.InvNextItem {
		if candidate == item {
			return true
		}
	}
	return false
}

func reportEquipmentRemoval(update *server.PlayerUpdateData, local, remote *server.Object, report, broadcast C.int) {
	if report != 0 {
		nox_xxx_netReportDequip_4D8590(C.int(update.Player.PlayerInd), C.nox_equipment_report_object_t(asObjectC(local)))
	}
	if broadcast != 0 {
		nox_xxx_netReportDequip_4D84C0(255, C.nox_equipment_report_object_t(asObjectC(remote)))
	}
}

func rejectEquipment(owner *server.Object, reason strman.ID, playSound bool) int {
	world := GetServer().S()
	world.NetPriMsgToPlayer(owner, reason, 0)
	if playSound {
		world.Audio.EventObj(925, owner, 2, owner.NetCode)
	}
	return 0
}

//export nox_xxx_playerDequipWeapon_53A140
func nox_xxx_playerDequipWeapon_53A140(owner, item *nox_object_t, report, broadcast C.int) C.int {
	return C.int(playerDequipWeapon53A140(asObjectS(owner), asObjectS(item), report, broadcast))
}

func playerDequipWeapon53A140(owner, item *server.Object, report, broadcast C.int) int {
	if owner == nil || item == nil {
		return 0
	}
	flags := uint32(nox_xxx_weaponInventoryEquipFlags_415820(asObjectC(item)))
	if owner.Class().Has(object.ClassMonster) {
		return npcWeaponDequipNative53A030(owner, item)
	}
	if !owner.Class().Has(object.ClassPlayer) || !item.Class().HasAny(equipmentWeaponClasses) {
		return 0
	}
	update := owner.UpdateDataPlayer()
	if update == nil || update.Player == nil {
		return 0
	}
	if update.State == server.PlayerState1 {
		Nox_xxx_playerSetState_4FA020(owner, server.PlayerState13)
	}
	if update.EquippedWeapon == nil || update.EquippedWeapon != item && flags != 2 {
		return 0
	}
	if flags&0xC != 0 {
		dequipEquippedQuiver53A0F0(owner, report, broadcast)
	}
	sub_4FEB60(asObjectC(owner), asObjectC(item))
	if flags == 2 {
		item.ObjFlags &^= object.FlagEquipped
		update.Player.WeaponEquip &^= 2
		reportEquipmentRemoval(update, item, item, report, broadcast)
	} else if equipped := update.EquippedWeapon; equipped != nil {
		equipped.ObjFlags &^= object.FlagEquipped
		update.Player.WeaponEquip &^= uint32(nox_xxx_weaponInventoryEquipFlags_415820(asObjectC(equipped)))
		reportEquipmentRemoval(update, equipped, item, report, broadcast)
		update.EquippedWeapon = nil
	}
	Nox_xxx_itemApplyDisengageEffect_4F3030(item, owner)
	if unsafe.Sizeof(uintptr(0)) == 4 && Get_gameex_flags()&2 != 0 {
		restoreGameExShield32(owner, update)
	}
	return 1
}

func restoreGameExShield32(owner *server.Object, update *server.PlayerUpdateData) {
	// GameEx stores the remembered shield in the original 32-bit player slot.
	remembered := (**server.Object)(unsafe.Add(unsafe.Pointer(update.Player), 2500))
	for item := owner.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.Class().Has(object.ClassArmor) && item.Flags().Has(object.FlagEquipped) &&
			nox_xxx_unitArmorInventoryEquipFlags_415C70(asObjectC(item))&0x3000000 != 0 {
			*remembered = item
		}
	}
	secondary := *(**nox_object_t)(unsafe.Pointer(&update.Field27))
	if secondary != nil && uint32(nox_xxx_weaponInventoryEquipFlags_415820(secondary))&shieldConflictingWeaponMask != 0 {
		return
	}
	shield := *remembered
	if shield == nil || uint8(shield.ObjFlags) != 16 {
		shield = nil
		for item := owner.InvFirstItem; item != nil; item = item.InvNextItem {
			if item.Class().Has(object.ClassArmor) && item.ObjFlags == 16 {
				shield = item
				break
			}
		}
	}
	if shield != nil {
		Nox_xxx_playerTryEquip_4F2F70(owner, shield)
	}
}

//export nox_xxx_playerEquipWeapon_53A420
func nox_xxx_playerEquipWeapon_53A420(owner, item *nox_object_t, report, playSound C.int) C.int {
	return C.int(playerEquipWeapon53A420(asObjectS(owner), asObjectS(item), report, playSound != 0))
}

func playerEquipWeapon53A420(owner, item *server.Object, report C.int, playSound bool) int {
	if owner == nil || item == nil {
		return 0
	}
	flags := uint32(nox_xxx_weaponInventoryEquipFlags_415820(asObjectC(item)))
	if !item.Class().HasAny(equipmentWeaponClasses) || item.Flags().Has(object.FlagEquipped) {
		return 0
	}
	if owner.Class().Has(object.ClassMonster) {
		return npcWeaponEquipNative53A2C0(owner, item)
	}
	if !owner.Class().Has(object.ClassPlayer) {
		return 0
	}
	update := owner.UpdateDataPlayer()
	if update == nil || update.Player == nil {
		return 0
	}
	if nox_xxx_probablyWarcryCheck_4FC3E0(asObjectC(owner), 2) != 0 ||
		nox_xxx_probablyWarcryCheck_4FC3E0(asObjectC(owner), 1) != 0 {
		return 0
	}
	if !Nox_xxx_playerClassCanUseItem_57B3D0(item, update.Player.PlayerClass()) {
		return rejectEquipment(owner, "weapon.c:WeaponEquipClassFail", playSound)
	}
	if playerCheckStrengthCall4F3180(owner, item) == 0 {
		return rejectEquipment(owner, "weapon.c:WeaponEquipStrengthFail", playSound)
	}
	if !inventoryHasEquipment(owner, item) {
		return 0
	}
	if update.State == server.PlayerState1 {
		Nox_xxx_playerSetState_4FA020(owner, server.PlayerState13)
	}
	if flags == 2 {
		if update.Player.WeaponEquip&0xC == 0 && sub_53A680(asObjectC(owner)) == 0 {
			return rejectEquipment(owner, "weapon.c:BowNotFound", playSound)
		}
		dequipEquippedQuiver53A0F0(owner, 1, 1)
	}
	if equipped := update.EquippedWeapon; equipped != nil && flags != 2 && playerDequipWeapon53A140(owner, equipped, 1, 1) == 0 {
		return 0
	}
	item.ObjFlags |= object.FlagEquipped
	update.Player.WeaponEquip |= flags
	nox_xxx_netReportEquip_4D8540(C.int(update.Player.PlayerInd), asObjectC(item), report)
	if flags != 2 {
		update.EquippedWeapon = item
	}
	reportEquippedWeaponCharges(update, item, flags)
	Nox_xxx_itemApplyEngageEffect_4F2FF0(item, owner)
	if flags&shieldConflictingWeaponMask != 0 {
		dequipEquippedShields53A3D0(owner)
	}
	return 1
}

func reportEquippedWeaponCharges(update *server.PlayerUpdateData, item *server.Object, flags uint32) {
	data := item.UseData.Ptr
	if data == nil {
		return
	}
	switch {
	case item.Class().Has(object.ClassWand) && uint32(item.ObjSubClass)&0x47f0000 != 0:
		wand := (*server.WandUseData)(data)
		Nox_xxx_netReportCharges_4D82B0(update.Player.PlayerInd, item, wand.Charge, wand.MaxCharge)
	case item.Class().Has(object.ClassWeapon):
		ammo := (*server.AmmoUseData)(data)
		if flags&0x82 != 0 {
			Nox_xxx_netReportCharges_4D82B0(update.Player.PlayerInd, item, ammo.Charge1, ammo.Charge0)
		} else if flags&0xC != 0 {
			ammo.Charge0 = 0
		}
	}
}
