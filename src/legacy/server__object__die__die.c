#include "common__object__armrlook.h"
#include "server__object__objutil.h"

#include "GAME1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "common__strman.h"

// TODO: convert table_274080

static int nox_armorDieGenericSound_54E170(float2* pos) {
	// GAME.EXE reuses the original 32-bit position pointer as the generic
	// armor sound ID. Preserve that bit-level behavior on wider hosts.
	return (int)(uint32_t)(uintptr_t)pos;
}

//----- (0054E170) --------------------------------------------------------
void nox_xxx_dieArmor_54E170_obj_die(nox_object_t* obj) {
	int plural = 0;
	int lang = nox_strman_get_lang_code();
	if (lang == 0 || lang == 1) {
		obj_412ae0_t* def = nox_xxx_equipClothFindDefByTT_413270(obj->typ_ind);
		if (def) {
			size_t len = nox_wcslen(def->field_2);
			wchar2_t* description = def->field_2;
			wchar2_t last = description[len - 1];
			if (last == 'S' || last == 's') {
				plural = 1;
			}
		}
	}
	nox_object_t* holder = obj->inv_holder;
	float2* pos = holder ? (float2*)&holder->x : (float2*)&obj->x;
	uint16_t material = obj->material;
	wchar2_t* format;
	int sound;
	if (material & 0x10) {
		if (plural) {
			format = nox_strman_loadString_40F1D0(
				"ArmorDieMetalPlural", 0, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c", 1536);
		} else {
			format = nox_strman_loadString_40F1D0(
				"ArmorDieMetal", 0, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c", 1538);
		}
		sound = 806;
	} else if (material & 8) {
		if (plural) {
			format = nox_strman_loadString_40F1D0(
				"ArmorDieWoodPlural", 0, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c", 1547);
		} else {
			format = nox_strman_loadString_40F1D0(
				"ArmorDieWood", 0, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c", 1549);
		}
		sound = 812;
	} else if (material & 4) {
		if (plural) {
			format = nox_strman_loadString_40F1D0(
				"ArmorDieHidePlural", 0, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c", 1558);
		} else {
			format = nox_strman_loadString_40F1D0(
				"ArmorDieHide", 0, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c", 1560);
		}
		sound = 809;
	} else if (material & 2) {
		if (plural) {
			format = nox_strman_loadString_40F1D0(
				"ArmorDieClothPlural", 0, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c", 1569);
		} else {
			format = nox_strman_loadString_40F1D0(
				"ArmorDieCloth", 0, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c", 1571);
		}
		sound = 815;
	} else {
		format = nox_strman_loadString_40F1D0(
			"ArmorDieGeneric", 0, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c", 1579);
		sound = nox_armorDieGenericSound_54E170(pos);
	}
	wchar2_t* name = nox_xxx_itemGetName_4E77E0_obj_util(obj);
	nox_xxx_netSendLineMessage_4D9EB0(holder, format, name);
	nox_xxx_audCreate_501A30(sound, pos, 0, 0);
	nox_xxx_delayedDeleteObject_4E5CC0(obj);
}

//----- (0054E370) --------------------------------------------------------
wchar2_t* sub_415B60(nox_object_t* a1);
void nox_xxx_dieWeapon_54E370_obj_die(nox_object_t* obj) {
	nox_object_t* holder = obj->inv_holder;
	float2* pos = holder ? (float2*)&holder->x : (float2*)&obj->x;
	uint16_t material = obj->material;
	if (material & 0x10) {
		wchar2_t* name = nox_xxx_itemGetName_4E77E0_obj_util(obj);
		wchar2_t* format = nox_strman_loadString_40F1D0(
			"WeaponDieMetal", 0, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c", 1626);
		nox_xxx_netSendLineMessage_4D9EB0(holder, format, name);
		nox_xxx_audCreate_501A30(818, pos, 0, 0);
		nox_xxx_delayedDeleteObject_4E5CC0(obj);
	} else {
		if (material & 8) {
			wchar2_t* name = nox_xxx_itemGetName_4E77E0_obj_util(obj);
			wchar2_t* format = nox_strman_loadString_40F1D0(
				"WeaponDieWood", 0, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c", 1633);
			nox_xxx_netSendLineMessage_4D9EB0(holder, format, name);
			nox_xxx_audCreate_501A30(819, pos, 0, 0);
		} else {
			wchar2_t* name = sub_415B60(obj);
			wchar2_t* format = nox_strman_loadString_40F1D0(
				"WeaponDieGeneric", 0, "C:\\NoxPost\\src\\Server\\Object\\die\\Die.c", 1640);
			nox_xxx_netSendLineMessage_4D9EB0(holder, format, name);
		}
		nox_xxx_delayedDeleteObject_4E5CC0(obj);
	}
}
