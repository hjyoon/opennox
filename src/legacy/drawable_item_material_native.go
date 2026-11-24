package legacy

/*
#include <string.h>
#include "GAME3_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_4B9470
func sub_4B9470(material **C.nox_flag_const_char_t) C.int {
	if material == nil {
		return 0
	}
	for offset := uintptr(177488); ; offset += 8 {
		name := (*C.char)(*memmap.PtrPtr(0x587000, offset))
		if name == nil {
			return 0
		}
		if C.strcmp((*C.char)(unsafe.Pointer(*material)), name) == 0 {
			return C.int(memmap.Int32(0x587000, offset+4))
		}
	}
}

//export sub_4B94E0
func sub_4B94E0(dr *nox_drawable) C.int {
	if dr == nil || dr.flags28&0x10000000 == 0 {
		return 0
	}
	return sub_4B9470((**C.nox_flag_const_char_t)(drawableItemData(dr).Field_109))
}

func setEquipmentMaterial(slot int32, color uint32) {
	nox_draw_setMaterial_4340A0(slot, int32(byte(color)), int32(byte(color>>8)), int32(byte(color>>16)))
}

func applyEquipmentMaterials(def unsafe.Pointer, drawable *nox_drawable) {
	if def == nil {
		return
	}
	for slot := C.int(1); slot <= 6; slot++ {
		setEquipmentMaterial(int32(slot), uint32(nox_modifier_getColorRGB(def, slot)))
	}
	if drawable == nil {
		return
	}
	item := drawableItemData(drawable)
	modifiers := [...]unsafe.Pointer{item.Field_108, item.Field_109, item.Field_110, item.Field_111}
	for i, modifier := range modifiers {
		if modifier != nil {
			color := uint32(nox_modifier_effect_getColorRGB(modifier))
			slot := nox_modifier_getColorSlot(def, C.int(i))
			setEquipmentMaterial(int32(slot), color)
		}
	}
}

//export sub_4B95D0
func sub_4B95D0(dr *nox_drawable) *C.uint32_t {
	def := nox_xxx_getProjectileClassById_413250(int32(dr.field_27))
	applyEquipmentMaterials(def, dr)
	return (*C.uint32_t)(def)
}

//export sub_4B9650
func sub_4B9650(thing C.int) *C.uint32_t {
	def := nox_xxx_getProjectileClassById_413250(int32(thing))
	applyEquipmentMaterials(def, nil)
	return (*C.uint32_t)(def)
}

//export sub_4B96F0
func sub_4B96F0(dr *nox_drawable) *C.uint32_t {
	def := nox_xxx_equipClothFindDefByTT_413270(int32(dr.field_27))
	applyEquipmentMaterials(def, dr)
	return (*C.uint32_t)(def)
}
