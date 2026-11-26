package legacy

/*
#include "GAME2_1.h"
#include "GAME3_2.h"
#include "client__gui__guiinv.h"
*/
import "C"

import (
	"math"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const clientInventoryDescriptionSource = "client_inventory_identify_native.go"

type clientInventoryDescription struct {
	window *nox_window
}

func (d clientInventoryDescription) add(text *wchar2_t) {
	nox_window_call_field_94_fnc(d.window, 16397, uintptr(unsafe.Pointer(text)), ^uintptr(0), internCStr(clientInventoryDescriptionSource), 0)
}

func (d clientInventoryDescription) label(key string, line int32) {
	d.add(clientInventoryString(key, line))
}

func (d clientInventoryDescription) text(text string) {
	ptr, free := CWString(text)
	defer free()
	d.add(ptr)
}

func (d clientInventoryDescription) gap(offset uintptr) {
	d.add((*wchar2_t)(memmap.PtrOff(0x5D4594, offset)))
}

func (d clientInventoryDescription) damage(key string, line int32, indent string, value float32) {
	var formatted [256]wchar2_t
	formatFixedWide(&formatted[0], clientInventoryString(key, line), formatValue{doubleBits: math.Float64bits(float64(value))})
	d.text(indent + GoWString(&formatted[0]))
}

func clientInventoryIdentifyCursor() {
	mouse := nox_client_getMousePos_4309F0()
	var point C.int2
	sub_463370(inventoryWindows.panel, &mouse, (*C.uint32_t)(unsafe.Pointer(&point)))
	switch {
	case clientInventoryPointInRect(&point, 0x587000, 136368):
		if (point.field_4-13)/50 != 1 {
			nox_client_setCursorType_477610(7)
			return
		}
	case clientInventoryPointInRect(&point, 0x587000, 136352):
	case clientInventoryPointInRect(&point, 0x587000, 136336):
		nox_client_setCursorType_477610(0)
		return
	case sub_478030() != 0 && sub_479870() != 0 && bool(sub_479880((*C.uint32_t)(unsafe.Pointer(&point)))):
	default:
		nox_client_setCursorType_477610(7)
		return
	}
	nox_client_setCursorType_477610(6)
}

//export sub_4627F0
func sub_4627F0(origin *C.uint32_t) C.uintptr_t {
	pos := unsafe.Slice(origin, 2)
	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_white_2523948()))
	nox_client_drawSetColor_434460(int32(Get_nox_color_black_2650656()))
	nox_client_drawRectFilledOpaque_49CE30(int32(pos[0])+11, int32(pos[1])+15, 200, 200)
	clientInventoryIdentifyCursor()

	selected := legacyPointers.inventoryTooltipItem
	previous := legacyPointers.inventoryPreviousTooltipItem
	if selected == previous {
		return C.uintptr_t(uintptr(unsafe.Pointer(selected)))
	}
	legacyPointers.inventoryPreviousTooltipItem = selected
	root := inventoryWindows.identify
	title := (*wchar2_t)(memmap.PtrOff(0x5D4594, 1063124))
	if selected == nil {
		copyWideString(title, clientInventoryString("thing.db:IdentifyDescription", 2529))
		list := nox_xxx_wndGetChildByID_46B0C0(root, 9156)
		return C.uintptr_t(nox_window_call_field_94_fnc(list, 16399, 0, 0, internCStr(clientInventoryDescriptionSource), 0))
	}
	name := GoWString(nox_xxx_clientAskInfoMb_4BF050(selected))
	if name == GoWString((*wchar2_t)(memmap.PtrOff(0x5D4594, 1063652))) {
		legacyPointers.inventoryPreviousTooltipItem = nil
	}
	header := GoWString(clientInventoryString("IdentifyItem", 2545)) + " " + name
	CWStringCopyTo(title, 260, header)
	sub_46AEE0(nox_xxx_wndGetChildByID_46B0C0(root, 9151), title)
	list := nox_xxx_wndGetChildByID_46B0C0(root, 9156)
	nox_window_call_field_94_fnc(list, 16399, 0, 0, internCStr(clientInventoryDescriptionSource), 0)

	description := clientInventoryDescription{window: list}
	cooperative := noxflags.HasGame(noxflags.GameModeCoop)
	clientInventoryDescribeDurability(description, selected, cooperative)
	description.gap(1063656)
	description.text(clientInventoryStatText("IdentifyWeight", 2620, int(selected.field_74_3), 0))
	description.gap(1063660)

	switch {
	case selected.flags28&0x2000000 != 0:
		clientInventoryDescribeArmor(description, selected)
	case selected.flags28&0x1001000 != 0:
		clientInventoryDescribeWeapon(description, selected, cooperative)
	}
	clientInventoryDescribeEnchantments(description, selected)
	if text := nox_get_thing_desc(int32(selected.field_27)); text != nil {
		description.add(text)
	}
	imageID := uint32(nox_get_thing_pretty_image(int32(selected.field_27)))
	icon := (*nox_video_bag_image_t)(unsafe.Pointer(uintptr(imageID)))
	return C.uintptr_t(nox_xxx_wndSetIcon_46AE60(nox_xxx_wndGetChildByID_46B0C0(root, 9155), icon))
}

func clientInventoryDescribeDurability(d clientInventoryDescription, item *nox_drawable, cooperative bool) {
	if cooperative {
		if item.field_73_2 == 0 {
			d.label("IdentifyDurabilityIndestructable", 2583)
			return
		}
		var current, maximum C.float
		sub_4633B0(item, &current, &maximum)
		d.text(clientInventoryStatText("IdentifyDurability", 2575, int(current), int(maximum)))
		return
	}
	labels := [...]struct {
		key  string
		line int32
	}{
		{"IdentifyDurabilityNoDamage", 2595},
		{"IdentifyDurabilitySlight", 2599},
		{"IdentifyDurabilityModerate", 2603},
		{"IdentifyDurabilitySevere", 2607},
		{"IdentifyDurabilityIndestructable", 2591},
	}
	label := labels[itemHealthBucket57B190(uint16(item.field_73_1), uint16(item.field_73_2))]
	d.label(label.key, label.line)
}

func clientInventoryDescribeArmor(d clientInventoryDescription, item *nox_drawable) {
	def := nox_xxx_equipClothFindDefByTT_413270(int32(item.field_27))
	scale := 1.0
	modifier := drawableItemData(item).Field_108
	if modifier != nil && nox_modifier_effect_getDefendFunc(modifier) == unsafe.Pointer(C.sub_4E0370) {
		scale = float64(nox_modifier_effect_getDefendFloat(modifier))
	}
	armor := clientInventoryStatInt(scale*float64(nox_modifier_getArmor(def))*1000 + 0.5)
	if item.flags29&2 != 0 {
		d.label("ArmorValueLabelNA", 2647)
	} else {
		d.text(clientInventoryStatText("ArmorValueLabel", 2649, armor, 0))
	}
	d.gap(1063664)
}

func clientInventoryIdentifyWeaponModifier(srv *server.Server, item *nox_drawable, pl *server.Player) (*server.Modifier, bool) {
	arrow := memmap.PtrUint32(0x5D4594, 1063644)
	bolt := memmap.PtrUint32(0x5D4594, 1063648)
	if *arrow == 0 {
		*arrow = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("ArcherArrow")))
		*bolt = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("ArcherBolt")))
	}
	knownAmmo := false
	if pl != nil && item.flags29&2 != 0 {
		switch {
		case pl.WeaponEquip&4 != 0:
			knownAmmo = true
			if def := srv.Modif.Nox_xxx_getProjectileClassById413250(int(*arrow)); def != nil {
				return def, true
			}
		case pl.WeaponEquip&8 != 0:
			knownAmmo = true
			if def := srv.Modif.Nox_xxx_getProjectileClassById413250(int(*bolt)); def != nil {
				return def, true
			}
		}
	}
	return srv.Modif.Nox_xxx_getProjectileClassById413250(int(item.field_27)), knownAmmo
}

func clientInventoryDescribeWeapon(d clientInventoryDescription, item *nox_drawable, cooperative bool) {
	srv := GetServer().S()
	pl := Get_dword_8531A0_2576()
	def, knownAmmo := clientInventoryIdentifyWeaponModifier(srv, item, pl)
	if def == nil || pl == nil {
		d.label("WeaponDamageLabelNA", 2770)
		d.gap(1063668)
		return
	}
	lightning := float32(sub_4626C0(item))
	fire := float32(sub_462700(item))
	scale := 1.0
	modifier := drawableItemData(item).Field_108
	if modifier != nil && nox_modifier_effect_getAttackFunc(modifier) == unsafe.Pointer(C.nox_xxx_effectDamageMultiplier_4E04C0) {
		scale = float64(nox_modifier_effect_getAttackFloat(modifier))
	}
	damage := srv.CalcBoltDamage4EF1E0(int32(pl.Info().Field2239()), def)
	total := float32(damage*scale + float64(lightning) + float64(fire))
	minimum := float64(def.DamageMin72)
	if cooperative && def.TypeInd == memmap.Uint32(0x5D4594, 1063648) {
		minimum = float64(nox_xxx_gamedataGetFloat_419D40((*C.char)(memmap.PtrOff(0x587000, 137632))))
	}
	base := float32(minimum * scale)
	strength := total - base - fire - lightning
	if strength < 0 {
		total -= strength
		strength = 0
	}
	switch {
	case item.flags29&0xc != 0:
		d.label("WeaponDamageLabelNA", 2770)
	case item.flags29&2 == 0 || knownAmmo:
		d.damage("WeaponDamageLabel", 2776, "", total)
	default:
		d.damage("WeaponDamageLabelUnknownPlus", 2773, "", total)
	}
	d.damage("BaseDamageLabel", 2785, "  ", base)
	d.damage("StrengthDamageLabel", 2792, "  ", strength)
	if fire > 0 {
		d.damage("FireDamageLabel", 2801, "  ", fire)
	}
	if lightning > 0 {
		d.damage("ElectricalDamageLabel", 2811, "  ", lightning)
	}
	d.gap(1063668)
}

func clientInventoryDescribeEnchantments(d clientInventoryDescription, item *nox_drawable) {
	if item.flags28&0x13001000 == 0 {
		return
	}
	var def unsafe.Pointer
	if item.flags28&0x11001000 != 0 {
		def = nox_xxx_getProjectileClassById_413250(int32(item.field_27))
	} else {
		def = nox_xxx_equipClothFindDefByTT_413270(int32(item.field_27))
	}
	if def == nil {
		d.label("IdentifySpecialAttributes", 2835)
		d.label("IdentifyUnknown", 2837)
		return
	}
	if item.flags28&0x10000000 != 0 {
		return
	}
	data := drawableItemData(item)
	first := nox_modifier_effect_getIdentDescription(data.Field_110)
	second := nox_modifier_effect_getIdentDescription(data.Field_111)
	if first != nil {
		d.label("IdentifySpecialAttributes", 2851)
		d.text("  " + GoWString(first))
	} else if second != nil {
		d.label("IdentifySpecialAttributes", 2868)
	}
	if second != nil {
		d.text("  " + GoWString(second))
	}
	if first != nil || second != nil {
		d.gap(1063672)
	}
}
