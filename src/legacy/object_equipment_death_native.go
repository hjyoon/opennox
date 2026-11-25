package legacy

/*
#include "server__object__die__die.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

const equipmentDeathSource = `C:\NoxPost\src\Server\Object\die\Die.c`

func equipmentDeathPosition(item *server.Object) *types.Pointf {
	if item.InvHolder != nil {
		return &item.InvHolder.PosVec
	}
	return &item.PosVec
}

func equipmentDeathMessage(holder *server.Object, format, name *wchar2_t) {
	if holder == nil || !holder.Class().Has(object.ClassPlayer) {
		return
	}
	update := holder.UpdateData
	var text [256]wchar2_t
	formatWideStrings(text[:], format, name, nil)
	sendLineMessageText(update, &text[0])
}

func armorDeathPlural(item *server.Object) bool {
	lang := nox_strman_get_lang_code()
	if lang != 0 && lang != 1 {
		return false
	}
	definition := (*C.obj_412ae0_t)(nox_xxx_equipClothFindDefByTT_413270(int32(item.TypeInd)))
	if definition == nil || definition.field_2 == nil {
		return false
	}
	length := nox_wcslen(definition.field_2)
	if length == 0 {
		return false
	}
	last := *wideStringUnit(definition.field_2, uintptr(length-1))
	return last == 'S' || last == 's'
}

//export nox_xxx_dieArmor_54E170_obj_die
func nox_xxx_dieArmor_54E170_obj_die(raw *C.nox_object_t) {
	equipmentArmorDeath(asObjectS(raw))
}

func equipmentArmorDeath(item *server.Object) {
	raw := asObjectC(item)
	plural := armorDeathPlural(item)
	holder, position := item.InvHolder, equipmentDeathPosition(item)
	material := object.Material(item.Material)
	key, line := "ArmorDieGeneric", int32(1579)
	// The original generic cue used the low 32 bits of the position pointer.
	cue := sound.ID(int32(uint32(uintptr(unsafe.Pointer(position)))))
	for _, choice := range [...]struct {
		material object.Material
		key      string
		line     int32
		cue      sound.ID
	}{
		{object.MaterialMetal, "ArmorDieMetal", 1538, 806},
		{object.MaterialWood, "ArmorDieWood", 1549, 812},
		{object.MaterialAnimalHide, "ArmorDieHide", 1560, 809},
		{object.MaterialCloth, "ArmorDieCloth", 1571, 815},
	} {
		if material.Has(choice.material) {
			key, line, cue = choice.key, choice.line, choice.cue
			if plural {
				key += "Plural"
				line -= 2
			}
			break
		}
	}
	format := nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(equipmentDeathSource), line)
	name := nox_xxx_itemGetName_4E77E0_obj_util(raw)
	equipmentDeathMessage(holder, format, name)
	GetServer().S().Audio.EventPos(cue, *position, 0, 0)
	nox_xxx_delayedDeleteObject_4E5CC0(raw)
}

//export nox_xxx_dieWeapon_54E370_obj_die
func nox_xxx_dieWeapon_54E370_obj_die(raw *C.nox_object_t) {
	equipmentWeaponDeath(asObjectS(raw))
}

func equipmentWeaponDeath(item *server.Object) {
	raw := asObjectC(item)
	holder, position := item.InvHolder, equipmentDeathPosition(item)
	material := object.Material(item.Material)
	key, line := "WeaponDieGeneric", int32(1640)
	var name *wchar2_t
	var cue sound.ID
	switch {
	case material.Has(object.MaterialMetal):
		key, line, cue = "WeaponDieMetal", 1626, 818
		name = nox_xxx_itemGetName_4E77E0_obj_util(raw)
	case material.Has(object.MaterialWood):
		key, line, cue = "WeaponDieWood", 1633, 819
		name = nox_xxx_itemGetName_4E77E0_obj_util(raw)
	default:
		name = sub_415B60(raw)
	}
	format := nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(equipmentDeathSource), line)
	equipmentDeathMessage(holder, format, name)
	if cue != 0 {
		GetServer().S().Audio.EventPos(cue, *position, 0, 0)
	}
	nox_xxx_delayedDeleteObject_4E5CC0(raw)
}
