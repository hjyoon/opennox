package legacy

/*
#include "GAME3_2.h"
#include "cgo_const.h"

typedef const nox_object_t modifier_engage_const_item_t;
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func enchantItemTestInventoryNative4DFBB0(owner *server.Object, flag byte) bool {
	engage := modifierEngagePointerNative4DFBB0(flag)
	if engage == nil || owner == nil {
		return false
	}
	const modifiableClasses = object.ClassWand | object.ClassWeapon | object.ClassArmor | object.ClassFlag
	for item := owner.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.ObjFlags.Has(object.FlagEquipped) || !item.ObjClass.HasAny(modifiableClasses) || item.InitData == nil {
			continue
		}
		data := (*server.ModifierInitData)(item.InitData)
		for _, mod := range data.Modifiers {
			if mod != nil && mod.Engage112 == engage {
				return true
			}
		}
	}
	return false
}

//export nox_xxx_enchantItemTestInventory_4DFBB0
func nox_xxx_enchantItemTestInventory_4DFBB0(owner *nox_object_t, flag C.uint8_t) C.int {
	return C.int(bool2int(enchantItemTestInventoryNative4DFBB0(asObjectS(owner), byte(flag))))
}

func activateModifierFlag(owner *server.Object, flag byte, audio sound.ID) {
	if owner == nil {
		return
	}
	owner.Field110 |= uint32(flag)
	GetServer().S().Audio.EventObj(audio, owner, 0, 0)
}

func deactivateModifierFlag(owner *server.Object, flag byte, audio sound.ID) {
	if owner == nil {
		return
	}
	if !enchantItemTestInventoryNative4DFBB0(owner, flag) {
		owner.Field110 &^= uint32(flag)
	}
	GetServer().S().Audio.EventObj(audio, owner, 0, 0)
}

//export sub_4DFB50
func sub_4DFB50(effect unsafe.Pointer, owner *nox_object_t, item *C.modifier_engage_const_item_t) {
	activateModifierFlag(asObjectS(owner), 8, sound.SoundLightOn)
}

//export sub_4DFB80
func sub_4DFB80(effect unsafe.Pointer, owner *nox_object_t, item *C.modifier_engage_const_item_t) {
	deactivateModifierFlag(asObjectS(owner), 8, sound.SoundLightOff)
}

// The original adds two binary32 operands and stores the result toward zero.
// Recover the residual of the native binary32 addition to direct rounding.
func modifierSpeedAddChop32(left, right float32) float32 {
	sum := float32(left + right)
	if math.Abs(float64(left)) < math.Abs(float64(right)) {
		left, right = right, left
	}
	residual := float32(right - float32(sum-left))
	if sum > 0 && residual < 0 || sum < 0 && residual > 0 {
		return math.Nextafter32(sum, 0)
	}
	return sum
}

func modifierSpeedPlayer(owner *server.Object) *server.Player {
	if owner == nil || !owner.ObjClass.Has(object.ClassPlayer) || owner.UpdateData == nil {
		return nil
	}
	return (*server.PlayerUpdateData)(owner.UpdateData).Player
}

func applyModifierSpeed(owner *server.Object, player *server.Player, delta float32, audio sound.ID) {
	owner.SpeedBonus = modifierSpeedAddChop32(owner.SpeedBonus, delta)
	bits := math.Float32bits(owner.SpeedBonus)
	nox_xxx_netReportStatsSpeed_4D9360(C.int(player.PlayerInd), asObjectC(owner), 0, C.int(int32(bits)))
	GetServer().S().Audio.EventObj(audio, owner, 0, 0)
}

//export nox_xxx_effectSpeedEngage_4DFC30
func nox_xxx_effectSpeedEngage_4DFC30(effect unsafe.Pointer, owner *nox_object_t, _ *C.modifier_engage_const_item_t) {
	unit := asObjectS(owner)
	player := modifierSpeedPlayer(unit)
	if player == nil {
		return
	}
	unit.Field110 |= 16
	applyModifierSpeed(unit, player, float32(nox_modifier_effect_getEngageFloat(effect)), sound.SoundHasteOn)
}

//export nox_xxx_effectSpeedDisengage_4DFCA0
func nox_xxx_effectSpeedDisengage_4DFCA0(effect unsafe.Pointer, owner *nox_object_t, _ *C.modifier_engage_const_item_t) {
	unit := asObjectS(owner)
	player := modifierSpeedPlayer(unit)
	if player == nil {
		return
	}
	if !enchantItemTestInventoryNative4DFBB0(unit, 16) {
		unit.Field110 &^= 16
	}
	applyModifierSpeed(unit, player, -float32(nox_modifier_effect_getEngageFloat(effect)), sound.SoundHasteOff)
}

//export sub_4DFD10
func sub_4DFD10(effect unsafe.Pointer, owner *nox_object_t, item *C.modifier_engage_const_item_t) {
	activateModifierFlag(asObjectS(owner), 1, sound.SoundProtectionFromFireOn)
}

//export nox_xxx_modifFireProtection_4DFD40
func nox_xxx_modifFireProtection_4DFD40(_ unsafe.Pointer, owner *nox_object_t, item *C.modifier_engage_const_item_t) {
	if item != nil {
		deactivateModifierFlag(asObjectS(owner), 1, sound.SoundProtectionFromFireOff)
	}
}

//export nox_xxx_buff_4DFD80
func nox_xxx_buff_4DFD80(effect unsafe.Pointer, owner *nox_object_t, item *C.modifier_engage_const_item_t) {
	activateModifierFlag(asObjectS(owner), 4, sound.SoundProtectionFromElectricityOn)
}

//export sub_4DFDB0
func sub_4DFDB0(effect unsafe.Pointer, owner *nox_object_t, item *C.modifier_engage_const_item_t) {
	deactivateModifierFlag(asObjectS(owner), 4, sound.SoundProtectionFromElectricityOff)
}

//export nox_xxx_checkPoisonProtectEnch_4DFDE0
func nox_xxx_checkPoisonProtectEnch_4DFDE0(effect unsafe.Pointer, owner *nox_object_t, item *C.modifier_engage_const_item_t) {
	activateModifierFlag(asObjectS(owner), 2, sound.SoundProtectionFromPoisonOn)
}

//export sub_4DFE10
func sub_4DFE10(effect unsafe.Pointer, owner *nox_object_t, item *C.modifier_engage_const_item_t) {
	deactivateModifierFlag(asObjectS(owner), 2, sound.SoundProtectionFromPoisonOff)
}

//export sub_4E0140
func sub_4E0140(effect unsafe.Pointer, owner *nox_object_t, item *C.modifier_engage_const_item_t) {
	activateModifierFlag(asObjectS(owner), 32, sound.SoundRegenerationOn)
}

//export sub_4E0170
func sub_4E0170(effect unsafe.Pointer, owner *nox_object_t, item *C.modifier_engage_const_item_t) {
	unit := asObjectS(owner)
	if unit != nil && unit.ObjClass.Has(object.ClassPlayer) {
		deactivateModifierFlag(unit, 32, sound.SoundRegenerationOff)
	}
}
