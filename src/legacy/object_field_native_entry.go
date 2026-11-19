package legacy

/*
#include "GAME3_3.h"
#include "GAME4_3.h"
#include "GAME5_2.h"
typedef const nox_object_t object_field_const_object_t;
typedef const nox_modifier_attrs_t object_field_const_attrs_t;
typedef const nox_color3_t object_field_const_color_t;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//export sub_5408A0
func sub_5408A0(owner *C.nox_object_t) C.int {
	return C.int(objectFieldMonsterAction5408A0(asObjectS((*nox_object_t)(owner))))
}

//export nox_xxx_unitNeedSync_4E44F0
func nox_xxx_unitNeedSync_4E44F0(owner *C.nox_object_t) {
	asObjectS((*nox_object_t)(owner)).NeedSync()
}

//export sub_4E4500
func sub_4E4500(owner *C.nox_object_t, value, mask, mode C.int) *C.int {
	return (*C.int)(objectFieldMask1404E4500(asObjectS((*nox_object_t)(owner)), uint32(value), uint32(mask), mode != 0))
}

//export nox_xxx_unitSetOnOff_4E4670
func nox_xxx_unitSetOnOff_4E4670(ptr *nox_object_t, enabled C.int) *C.int {
	obj := asObjectS(ptr)
	obj.SetOnOff(enabled != 0)
	return (*C.int)(objectSyncFieldsEnd4E4500(obj))
}

//export nox_xxx_unitRaise_4E46F0
func nox_xxx_unitRaise_4E46F0(ptr *nox_object_t, z C.float) {
	asObjectS(ptr).Raise(float32(z))
}

//export nox_xxx_servMarkObjAnimFrame_4E4880
func nox_xxx_servMarkObjAnimFrame_4E4880(ptr *nox_object_t, frame C.int) *C.int {
	obj := asObjectS(ptr)
	obj.MarkAnimFrame(uint32(frame))
	return (*C.int)(objectSyncFieldsEnd4E4500(obj))
}

//export nox_xxx_setUnitBuffFlags_4E48F0
func nox_xxx_setUnitBuffFlags_4E48F0(ptr *nox_object_t, flags C.uint) *C.int {
	obj := asObjectS(ptr)
	obj.SetBuffFlags(uint32(flags), func(player *server.Player, value uint32) {
		Nox_xxx_playerResetProtectionCRC_56F7D0(player.ProtUnitBuffs, int(value))
	})
	return (*C.int)(objectSyncFieldsEnd4E4500(obj))
}

//export nox_xxx_modifSetItemAttrs_4E4990
func nox_xxx_modifSetItemAttrs_4E4990(ptr *nox_object_t, attrPtr *C.object_field_const_attrs_t) C.uintptr_t {
	obj := asObjectS(ptr)
	attrs := (*server.ModifierInitData)(unsafe.Pointer(attrPtr))
	forced := obj.Class().Has(object.ClassWand) && uint32(obj.SubClass())&0x047f0000 != 0
	if !forced {
		var hasModifiers bool
		for _, modifier := range attrs.Modifiers {
			if modifier != nil {
				hasModifiers = true
			}
		}
		if !hasModifiers {
			return C.uintptr_t(uintptr(unsafe.Pointer(&attrs.Field16)))
		}
	}
	cache := memmap.PtrUint32(0x5D4594, 1564960)
	if *cache == 0 {
		*cache = uint32(GetServer().S().Types.IndByID("TeamBase"))
	}
	teamBase := *cache
	if !obj.SetModifierAttrs(attrs, teamBase) {
		return C.uintptr_t(teamBase)
	}
	return C.uintptr_t(uintptr(objectSyncFieldsEnd4E4500(obj)))
}

//export nox_xxx_objectGetMass_4E4A70
func nox_xxx_objectGetMass_4E4A70(ptr *C.object_field_const_object_t) C.double {
	obj := asObjectS((*nox_object_t)(unsafe.Pointer(ptr)))
	return C.double(obj.Mass)
}

//export nox_xxx_objectGetBuffs_4E4A80
func nox_xxx_objectGetBuffs_4E4A80(owner *C.object_field_const_object_t) C.uint {
	return C.uint(asObjectS((*nox_object_t)(unsafe.Pointer(owner))).Buffs)
}

//export nox_xxx_setNPCColor_4E4A90
func nox_xxx_setNPCColor_4E4A90(ptr *nox_object_t, index C.uchar, color *C.object_field_const_color_t) *C.int {
	obj := asObjectS(ptr)
	obj.Nox_xxx_setNPCColor_4E4A90(byte(index), (*server.Color3)(unsafe.Pointer(color)))
	return (*C.int)(objectSyncFieldsEnd4E4500(obj))
}

//export nox_xxx_npcSetItemEquipFlags_4E4B20
func nox_xxx_npcSetItemEquipFlags_4E4B20(ptr, item *nox_object_t, equipped C.int) *C.int {
	obj := asObjectS(ptr)
	obj.SetNPCItemEquipFlags(asObjectS(item), equipped == 1, objectNPCWeaponEquipFlags, objectNPCArmorEquipFlags)
	return (*C.int)(objectSyncFieldsEnd4E4500(obj))
}

//export nox_xxx_armorHaveSameSubclass_53E7B0
func nox_xxx_armorHaveSameSubclass_53E7B0(owner *C.nox_object_t, item *C.object_field_const_object_t) *C.nox_object_t {
	return (*C.nox_object_t)(asObjectC(objectFieldSameArmor53E7B0(asObjectS((*nox_object_t)(owner)), asObjectS((*nox_object_t)(unsafe.Pointer(item))))))
}

//export nox_xxx_netGetUnitCodeCli_578B00
func nox_xxx_netGetUnitCodeCli_578B00(drawable *C.nox_drawable) C.uint {
	return C.uint(objectFieldDrawableCode578B00((*client.Drawable)(unsafe.Pointer(drawable))))
}

//export nox_xxx_netClearHighBit_578B30
func nox_xxx_netClearHighBit_578B30(word C.uint) C.int {
	// GAME.EXE clears AH's high bit and preserves all other argument bits.
	return C.int(uint32(word) &^ uint32(0x8000))
}

//export nox_xxx_netTestHighBit_578B70
func nox_xxx_netTestHighBit_578B70(word C.uint) C.uint {
	return C.uint((uint32(word) >> 15) & 1)
}
