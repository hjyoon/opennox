package legacy

/*
#include "client__gui__guiinv.h"
#include "client__gui__guimsg.h"
#include "GAME2_1.h"

typedef const nox_modifier_attrs_t inventory_const_modifier_attrs_t;
*/
import "C"

import "github.com/opennox/opennox/v1/common/memmap"

func clientInventoryString(key string, line int32) *wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(`C:\NoxPost\src\Client\Gui\guiinv.c`), line)
}

func clientInventoryIsSpecialThing(thingType C.uint32_t) bool {
	switch uint32(thingType) {
	case uint32(legacyAddressableWords.dword_5d4594_1062560), memmap.Uint32(0x5D4594, 1049728), memmap.Uint32(0x5D4594, 1049724),
		uint32(legacyAddressableWords.dword_5d4594_1062556), uint32(legacyAddressableWords.dword_5d4594_1062564):
		return true
	}
	return false
}

//export nox_xxx_spritePickup_461660
func nox_xxx_spritePickup_461660(code, thingType C.int, attrs *C.inventory_const_modifier_attrs_t) C.int {
	if clientInventoryIsSpecialThing(C.uint32_t(thingType)) {
		return 1
	}
	cell, row, column := clientInventoryStackPickup(code, thingType)
	if cell != nil {
		if cell.field_0.flags28&0x10 != 0 {
			sub_472310()
		}
	} else {
		var pos C.int2
		if sub_4617C0(code, thingType, attrs, &pos) == 0 {
			nox_xxx_printCentered_445490(clientInventoryString("InventoryFull", 682))
			return 0
		}
		row, column = int(pos.field_4), int(pos.field_0)
		drawable := inventoryGridCell(row, column).field_0
		if drawable.flags28&0x10 != 0 {
			sub_472310()
		}
		if drawable.flags28&0x03001000 != 0 {
			legacyWords.dword_5d4594_1062516 = 0
			if row >= 3 {
				legacyWords.dword_5d4594_1062516 = C.uint32_t(10 * (5*row - 10))
			}
		}
	}
	if thing := nox_get_thing(int32(thingType)); thing != nil && thing.pri_class&0x01001000 != 0 {
		sub_4673F0(C.int(column), C.int(row))
	}
	return 1
}

//export sub_4617C0
func sub_4617C0(code, thingType C.int, attrs *C.inventory_const_modifier_attrs_t, pos *C.int2) C.int {
	for row := range inventoryRows - 1 {
		for column := range inventoryColumns {
			cell := inventoryGridCell(row, column)
			if cell.field_140 != 0 {
				continue
			}
			drawable := nox_new_drawable_for_thing(int32(thingType))
			cell.field_0 = drawable
			if drawable == nil {
				nox_xxx_printCentered_445490(clientInventoryString("DrawablesExhausted", 550))
				return 0
			}
			drawable.flags30 |= 0x40000000
			cell.field_4, cell.field_140 = C.uint32_t(code), 1
			if drawable.flags28&0x13001000 != 0 {
				item := drawableItemData(drawable)
				item.Field_108, item.Field_109, item.Field_110, item.Field_111 = attrs.modifiers[0], attrs.modifiers[1], attrs.modifiers[2], attrs.modifiers[3]
				item.Field_112_0, item.Field_112_2 = int16(attrs.field_16), int16(attrs.field_16>>16)
			}
			if pos != nil {
				pos.field_0, pos.field_4 = C.int(column), C.int(row)
			}
			if sub_461930() != 0 && clientInventoryAlternateCell() == nil &&
				(drawable.flags28&0x1000 != 0 || drawable.flags28&0x1000000 != 0 && drawable.flags29&2 == 0) {
				projectile := GetServer().S().Modif.Nox_xxx_getProjectileClassById413250(int(drawable.field_27))
				player := Get_dword_8531A0_2576()
				if projectile == nil || player != nil && projectile.Classes62&(byte(1)<<player.PlayerClass()) != 0 {
					nox_xxx_clientSetAltWeapon_461550(cell)
					cell.field_136 = 1
				}
			}
			return 1
		}
	}
	return 0
}

//export sub_461A80
func sub_461A80(code C.int) {
	if cell, index := inventoryLookupNativeSearch(uint32(code)); cell != nil {
		refresh := cell.field_0 != nil && cell.field_0.flags28&0x10 != 0
		sub_461E60(cell, C.uint32_t(index))
		cell.field_132 = 0
		sub_461B50()
		if drawable := sub_461F90(code); drawable != nil {
			nox_xxx_spriteDelete_45A4B0(drawable)
		}
		if refresh {
			sub_472310()
		}
		return
	}
	dragged := nox_client_inventory_get_dragged()
	if dragged == nil || dragged.field_32 != C.uint32_t(code) {
		nox_xxx_printCentered_445490(clientInventoryString("DroppedNotFound", 1439))
		return
	}
	if dragged.flags28&0x10 != 0 {
		sub_472310()
	}
	nox_xxx_spriteDelete_45A4B0(dragged)
	nox_client_inventory_set_dragged(nil)
	legacyWords.dword_5d4594_1049856 = 0
	nox_xxx_cursorResetDraggedItem_4776A0()
}

func clientInventoryCloneDrawable(source *nox_drawable) *nox_drawable {
	drawable := nox_new_drawable_for_thing(int32(source.field_27))
	if drawable != nil {
		drawable.flags30 |= 0x40000000
		*drawableItemData(drawable) = *drawableItemData(source)
		drawable.field_113 = source.field_113
		drawable.field_73_1, drawable.field_73_2 = source.field_73_1, source.field_73_2
	}
	return drawable
}

//export sub_462040
func sub_462040(code C.int) {
	cell, _ := inventoryLookupNativeSearch(uint32(code))
	var source *nox_drawable
	if cell != nil {
		source = cell.field_0
	} else {
		source = nox_client_inventory_get_dragged()
	}
	// The stack owns an item's net code until the sprite is being dragged.
	if source == nil || cell == nil && source.field_32 != C.uint32_t(code) {
		nox_xxx_printCentered_445490(clientInventoryString("EquippedNotFound", 1605))
		return
	}
	slot := sub_4622E0(source)
	if slot == 9 {
		nox_xxx_printCentered_445490(clientInventoryString("TooManyEquipped", 1701))
		return
	}
	equipped := clientInventoryCloneDrawable(source)
	if equipped == nil {
		nox_xxx_printCentered_445490(clientInventoryString("DrawablesExhausted", 1619))
		return
	}
	equipped.field_32 = C.uint32_t(code)
	sub_4623E0(equipped, slot)
	if cell != nil {
		cell.field_132 = 1
		if cell.field_136 != 0 {
			nox_xxx_clientSetAltWeapon_461550(nil)
			cell.field_136 = 0
		}
	}
	if equipped.flags28&0x1000000 != 0 && equipped.flags29&0xc != 0 {
		clientInventoryEquipAmmo()
	}
	if slot == 0 {
		clientItemStorage.equipped_item_code = equipped.field_32
	}
	if item := drawableItemData(equipped); item.Field_112_0 >= 0 {
		sub_470D90(C.int(item.Field_112_0), C.int(item.Field_112_2))
	}
	if code := legacyWords.dword_5d4594_1062496; code != 0 {
		if alternate, _ := inventoryLookupNativeSearch(uint32(code)); alternate != nil {
			alternate.field_136 = 1
			nox_xxx_clientSetAltWeapon_461550(alternate)
			legacyWords.dword_5d4594_1062496 = 0
		}
	}
}

func clientInventoryEquipAmmo() {
	if code := clientItemStorage.equipped_item_code; code != 0 {
		if cell, _ := inventoryLookupNativeSearch(uint32(code)); cell != nil && cell.field_0 != nil {
			cell.field_0.field_32 = cell.field_4
			nox_xxx_clientEquip_4623B0(cell.field_0)
			return
		}
	}
	for row := range inventoryRows - 1 {
		for column := range inventoryColumns {
			cell := inventoryGridCell(row, column)
			drawable := cell.field_0
			if cell.field_140 != 0 && drawable != nil && drawable.flags28&0x1000000 != 0 && drawable.flags29 == 2 {
				drawable.field_32 = cell.field_4
				nox_xxx_clientEquip_4623B0(drawable)
				return
			}
		}
	}
}

//export nox_xxx_cliInventorySpriteUpd_465A30
func nox_xxx_cliInventorySpriteUpd_465A30() {
	cell := inventoryGridCell(int(legacyWords.dword_5d4594_1049800_inventory_click_row_index), int(legacyWords.dword_5d4594_1049796_inventory_click_column_index))
	if cell.field_140 == 0 {
		return
	}
	dragged := clientInventoryCloneDrawable(cell.field_0)
	nox_client_inventory_set_dragged(dragged)
	if dragged == nil {
		nox_xxx_printCentered_445490(clientInventoryString("DrawablesExhausted", 1123))
		return
	}
	dragged.field_32 = cell.field_4
	sub_461E60(cell, 0)
}
