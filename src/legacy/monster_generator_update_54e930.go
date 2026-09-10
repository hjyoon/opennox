package legacy

/*
#include "defs.h"
#include "GAME1_1.h"
#include "GAME3_3.h"
#include "GAME4_1.h"
#include "GAME4_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

const npcArmorSpecialFlag53E520 = object.Flags(0x10000000)

func npcArmorValueNative415C00(item *server.Object) float32 {
	if item == nil || uint32(item.ObjClass)&0x2000000 == 0 {
		return 0
	}
	definition := GetServer().S().Modif.Nox_xxx_equipClothFindDefByTT413270(int(item.TypeInd))
	if definition == nil {
		return 0
	}
	value := definition.DamageCoeffOrArmor64
	if item.InitData == nil {
		return value
	}
	effect := item.InitDataModifier().Modifiers[0]
	if effect != nil && effect.Defend76.Fnc != nil {
		// All stock armor modifiers resolve to the pointer-width-safe callbacks
		// handled here. An unknown ABI32 callback is deliberately not entered on
		// wide targets; keeping the base value is safer than truncating objects.
		itemDurabilityApplyDefendNative4E1560(effect, item, nil, item, nil, &value)
	}
	return value
}

func npcSetArmorValueNative53E300(owner *server.Object, value float32) {
	if owner == nil || owner.UpdateData == nil {
		return
	}
	if owner.Class().Has(object.ClassPlayer) {
		*(*float32)(unsafe.Pointer(&owner.UpdateDataPlayer().Field57)) = value
		return
	}
	if owner.Class().Has(object.ClassMonster) {
		*(*float32)(unsafe.Pointer(&owner.UpdateDataMonster().Field518)) = value
	}
}

func npcRecalculateArmorNative53E300(owner *server.Object) {
	value := float32(0)
	for item := owner.FirstItem(); item != nil; item = item.NextItem() {
		if uint32(item.ObjClass)&0x2000000 != 0 && item.Flags().Has(object.FlagEquipped) {
			value += npcArmorValueNative415C00(item)
		}
	}
	if float64(value) > 1.0 {
		value = 1
	}
	npcSetArmorValueNative53E300(owner, value)
}

func npcArmorDequipNative53E3A0(owner, item *server.Object) int {
	if owner == nil || item == nil || uint32(item.ObjClass)&0x2000000 == 0 ||
		!item.Flags().Has(object.FlagEquipped) {
		return 0
	}
	found := false
	for candidate := owner.FirstItem(); candidate != nil; candidate = candidate.NextItem() {
		if candidate == item {
			found = true
			break
		}
	}
	if !found {
		return 0
	}
	item.ObjFlags &^= object.FlagEquipped
	if uint8(owner.ObjSubClass)&0x10 != 0 {
		owner.SetNPCItemEquipFlags(item, false, objectNPCWeaponEquipFlags, objectNPCArmorEquipFlags)
	}
	item.ObjFlags &^= npcArmorSpecialFlag53E520
	npcRecalculateArmorNative53E300(owner)
	Nox_xxx_itemApplyDisengageEffect_4F3030(item, owner)
	return 1
}

func npcArmorEquipNative53E520(owner, item *server.Object) int {
	if owner == nil || item == nil || owner.UpdateData == nil ||
		uint32(item.ObjClass)&0x2000000 == 0 || item.Flags().Has(object.FlagEquipped) {
		return 0
	}
	found := false
	for candidate := owner.FirstItem(); candidate != nil; candidate = candidate.NextItem() {
		if candidate == item {
			found = true
			break
		}
	}
	if !found {
		return 0
	}
	for equipped := owner.FirstItem(); equipped != nil; equipped = equipped.NextItem() {
		if equipped != item && uint32(equipped.ObjClass)&0x2000000 != 0 &&
			equipped.Flags().Has(object.FlagEquipped) && equipped.ObjSubClass == item.ObjSubClass {
			npcArmorDequipNative53E3A0(owner, equipped)
			break
		}
	}
	item.ObjFlags |= object.FlagEquipped
	if uint8(owner.ObjSubClass)&0x10 != 0 {
		owner.SetNPCItemEquipFlags(item, true, objectNPCWeaponEquipFlags, objectNPCArmorEquipFlags)
	}
	if GetServer().S().Armor.Nox_xxx_unitArmorInventoryEquipFlags_415C70(item)&0xC0D != 0 {
		item.ObjFlags |= npcArmorSpecialFlag53E520
	}
	npcRecalculateArmorNative53E300(owner)
	Nox_xxx_itemApplyEngageEffect_4F2FF0(item, owner)
	if uint8(item.ObjSubClass)&2 != 0 {
		for weapon := owner.FirstItem(); weapon != nil; {
			next := weapon.NextItem()
			if weapon.Flags().Has(object.FlagEquipped) && uint32(weapon.ObjClass)&0x1001000 != 0 &&
				GetServer().S().Weapons.Nox_xxx_weaponInventoryEquipFlags_415820(weapon)&0x7FFE40C != 0 {
				npcWeaponDequipNative53A030(owner, weapon)
			}
			weapon = next
		}
	}
	return 1
}

func monsterGeneratorRuntime54E930() server.MonsterGeneratorRuntime54E930 {
	outer := GetServer()
	return server.MonsterGeneratorRuntime54E930{
		QuestStage: func() uint32 {
			return uint32(Nox_game_getQuestStage_4E3CC0())
		},
		QuestGroup:           Nox_xxx_getQuestStage_51A930,
		MapTileAllowTeleport: mapTileAllowTeleport411A90,
		CreateAt: func(object, owner *server.Object, position types.Pointf) {
			outer.CreateObjectAt(object, owner, position)
		},
		ScriptCallback: func(
			callback *server.ScriptCallback,
			caller, trigger *server.Object,
			event server.ScriptEventType,
		) {
			outer.NoxScriptC().ScriptCallback(callback, caller, trigger, event)
		},
		SendSpawnFX: func(values [4]int32, length int16) {
			C.nox_xxx_sendGeneratorSpawnFX_523830(
				(*C.int4)(unsafe.Pointer(&values)),
				C.short(length),
			)
		},
		InventoryPut: func(owner, item *server.Object, report int32) {
			inventoryPutImpl4F3070(owner, item, report != 0)
		},
		EquipWeapon:       npcWeaponEquipNative53A2C0,
		EquipArmor:        npcArmorEquipNative53E520,
		QuestHealthFactor: func() float64 { return float64(C.sub_4E40F0()) },
		FloatToInt:        func(value float32) int32 { return int32(C.nox_float2int(C.float(value))) },
	}
}

var monsterGeneratorUpdateCall54E930 = func(object *server.Object) {
	GetServer().S().MonsterGeneratorUpdate54E930(object, monsterGeneratorRuntime54E930())
}
