#include "client__gui__guiinv.h"
#include "client__gui__window.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME5_2.h"
#include "client__drawable__drawable.h"
#include "client__gui__guimsg.h"
#include "client__gui__tooltip.h"
#include "common__magic__speltree.h"
#include "common__object__modifier.h"
#include "common__strman.h"
#include "input_common.h"
#include "operators.h"

extern uintptr_t dword_8531A0_2576;
extern uint32_t dword_5d4594_1049844;
extern uint32_t dword_5d4594_1049808;
extern uint32_t dword_5d4594_1062484;
extern uint32_t dword_5d4594_1062556;
extern uint32_t dword_5d4594_1062564;
extern uint32_t dword_5d4594_1062560;
extern uint32_t dword_5d4594_1049804;
extern uint32_t dword_5d4594_1062496;
extern nox_drawable* dword_5d4594_1063120;
extern uint32_t dword_5d4594_1062488;
extern uint32_t dword_5d4594_1062516;
extern uint32_t dword_5d4594_1049856;
extern uint32_t dword_5d4594_1049800_inventory_click_row_index;
extern uint32_t dword_5d4594_1049796_inventory_click_column_index;
extern uint32_t dword_5d4594_1062512;
extern uint32_t dword_5d4594_1049864;
extern nox_drawable* dword_5d4594_1063116;
extern uintptr_t dword_5d4594_1062480;
extern uintptr_t array_5D4594_1049872[9];
extern int nox_win_width;
extern int nox_win_height;

extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_red_2589776;
extern uint32_t nox_color_blue_2650684;
extern uint32_t nox_color_cyan_2649820;
extern uint32_t nox_color_yellow_2589772;
extern uint32_t nox_color_violet_2598268;
extern uint32_t nox_color_black_2650656;
extern uint32_t nox_color_orange_2614256;

nox_window* nox_win_unk5 = 0;
nox_window* dword_5d4594_1062452 = 0;
nox_window* nox_inventory_window = 0;
nox_window* nox_inventory_current_weapon_window = 0;
nox_window* nox_inventory_identify_window = 0;
nox_window* nox_inventory_scroll_window = 0;
nox_window* nox_inventory_journal_button = 0;
nox_window* nox_inventory_stats_button = 0;
nox_window* nox_inventory_scroll_up_button = 0;
nox_window* nox_inventory_scroll_down_button = 0;
nox_window* nox_inventory_close_button = 0;
void* nox_inventory_font = 0;
nox_video_bag_image_t* nox_inventory_images[NOX_INV_IMG_COUNT];
nox_things_imageRef_t* nox_inventory_extra_lives_anim = 0;
nox_inventory_cell_t nox_client_inventory_grid_1050020[NOX_INVENTORY_CELLS_MAX] = {0};

//----- (004627F0) --------------------------------------------------------
uintptr_t sub_4627F0(uint32_t* a1) {
	int v2;           // ebx
	int v3;           // eax
	uintptr_t result; // eax
	wchar2_t* v5;      // eax
	nox_window* v6;   // eax
	wchar2_t* v7;      // eax
	wchar2_t* v8;      // eax
	nox_window* v9;   // eax
	nox_window* v10;  // esi
	wchar2_t* v11;     // eax
	wchar2_t* v12;     // eax
	wchar2_t* v13;     // eax
	wchar2_t* v14;     // eax
	nox_drawable* v15; // eax
	int v16;          // ecx
	void* v17;        // eax
	double v18;       // st7
	void* v19;        // ecx
	int v20;          // eax
	wchar2_t* v21;     // eax
	wchar2_t* v22;     // eax
	nox_playerInfo* v23; // ecx
	int v24;          // eax
	int v25;          // ecx
	nox_modifier_t* v26; // edi
	void* v27;        // eax
	double v28;       // st7
	int v29;          // eax
	double v30;       // st7
	double v31;       // st7
	double v32;       // st7
	int v33;          // eax
	wchar2_t* v34;     // eax
	wchar2_t* v35;     // eax
	wchar2_t* v36;     // eax
	wchar2_t* v37;     // eax
	wchar2_t* v38;     // eax
	int v39;          // ecx
	void* v40;        // eax
	wchar2_t* v41;     // eax
	wchar2_t* v42;     // eax
	void** v43;       // edi
	void* v44;        // ecx
	wchar2_t* v45;     // eax
	int v46;          // ebx
	void* v47;        // ecx
	wchar2_t* v48;     // eax
	wchar2_t* v49;    // eax
	nox_window* v50;  // eax
	wchar2_t* v51;     // [esp-Ch] [ebp-444h]
	wchar2_t* v52;     // [esp-Ch] [ebp-444h]
	wchar2_t* v53;     // [esp-Ch] [ebp-444h]
	wchar2_t* v54;     // [esp-4h] [ebp-43Ch]
	wchar2_t* v55;     // [esp-4h] [ebp-43Ch]
	int v56;          // [esp+0h] [ebp-438h]
	double v57;       // [esp+0h] [ebp-438h]
	double v58;       // [esp+0h] [ebp-438h]
	double v59;       // [esp+0h] [ebp-438h]
	double v60;       // [esp+0h] [ebp-438h]
	double v61;       // [esp+0h] [ebp-438h]
	double v62;       // [esp+0h] [ebp-438h]
	int v63;          // [esp+4h] [ebp-434h]
	int v64;          // [esp+4h] [ebp-434h]
	float v65;        // [esp+4h] [ebp-434h]
	int v66;          // [esp+4h] [ebp-434h]
	int v67;          // [esp+4h] [ebp-434h]
	float v68;        // [esp+18h] [ebp-420h]
	float v69;        // [esp+1Ch] [ebp-41Ch]
	float v70;        // [esp+20h] [ebp-418h]
	float v71;        // [esp+24h] [ebp-414h]
	float v72;        // [esp+28h] [ebp-410h]
	int v73;          // [esp+2Ch] [ebp-40Ch]
	int2 v74;         // [esp+30h] [ebp-408h]
	wchar2_t v75[256]; // [esp+38h] [ebp-400h]
	wchar2_t v76[256]; // [esp+238h] [ebp-200h]

	v73 = 1;
	nox_point mpos = nox_client_getMousePos_4309F0();
	nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
	v2 = 0;
	nox_client_drawSetColor_434460(nox_color_black_2650656);
	nox_client_drawRectFilledOpaque_49CE30(*a1 + 11, a1[1] + 15, 200, 200);
	sub_463370(nox_inventory_window, &mpos, &v74);
	if (nox_xxx_pointInRect_4281F0(&v74, (int4*)getMemAt(0x587000, 136352)) || nox_xxx_pointInRect_4281F0(&v74, (int4*)getMemAt(0x587000, 136368))) {
		if (nox_xxx_pointInRect_4281F0(&v74, (int4*)getMemAt(0x587000, 136368)) && (v74.field_4 - 13) / 50 != 1) {
			nox_client_setCursorType_477610(7);
			goto LABEL_14;
		}
	} else {
		if (nox_xxx_pointInRect_4281F0(&v74, (int4*)getMemAt(0x587000, 136336))) {
			nox_client_setCursorType_477610(0);
			goto LABEL_14;
		}
		if (!sub_478030()) {
			nox_client_setCursorType_477610(7);
			goto LABEL_14;
		}
		if (!sub_479870() || (LOBYTE(v3) = sub_479880(&v74), !v3)) {
			nox_client_setCursorType_477610(7);
			goto LABEL_14;
		}
	}
	nox_client_setCursorType_477610(6);
LABEL_14:
	result = (uintptr_t)dword_5d4594_1063116;
	if (!dword_5d4594_1063116) {
		if (dword_5d4594_1063120) {
			dword_5d4594_1063120 = 0;
			v5 = nox_strman_loadString_40F1D0("thing.db:IdentifyDescription", 0,
											  "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2529);
			nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1063124), v5);
			v6 = nox_xxx_wndGetChildByID_46B0C0(nox_inventory_identify_window, 9156);
			result = nox_window_call_field_94(v6, 16399, 0, 0);
		}
		return result;
	}
	if (dword_5d4594_1063120 == dword_5d4594_1063116) {
		return result;
	}
	dword_5d4594_1063120 = dword_5d4594_1063116;
	v7 = nox_strman_loadString_40F1D0("IdentifyItem", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2545);
	nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1063124), L"%s ", v7);
	v8 = nox_xxx_clientAskInfoMb_4BF050(dword_5d4594_1063116);
	nox_wcsncpy(v75, v8, sizeof(v75)/2);
	if (!nox_wcscmp(v75, (const wchar2_t*)getMemAt(0x5D4594, 1063652))) {
		dword_5d4594_1063120 = 0;
	}
	nox_wcscat((wchar2_t*)getMemAt(0x5D4594, 1063124), v75);
	v9 = nox_xxx_wndGetChildByID_46B0C0(nox_inventory_identify_window, 9151);
	sub_46AEE0(v9, (const wchar2_t*)getMemAt(0x5D4594, 1063124));
	v10 = nox_xxx_wndGetChildByID_46B0C0(nox_inventory_identify_window, 9156);
	nox_window_call_field_94(v10, 16399, 0, 0);
	if (nox_common_gameFlags_check_40A5C0(2048)) {
		if (dword_5d4594_1063116->field_73_2) {
			sub_4633B0(dword_5d4594_1063116, &v71, &v68);
			v63 = (int)v68;
			v56 = (int)v71;
			v11 =
				nox_strman_loadString_40F1D0("IdentifyDurability", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2575);
			nox_swprintf(v75, v11, v56, v63);
		} else {
			v12 = nox_strman_loadString_40F1D0("IdentifyDurabilityIndestructable", 0,
											   "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2583);
			nox_wcsncpy(v75, v12, sizeof(v75)/2);
		}
	} else {
		switch (sub_57B190(dword_5d4594_1063116->field_73_1, dword_5d4594_1063116->field_73_2)) {
		case 0:
			v13 = nox_strman_loadString_40F1D0("IdentifyDurabilityNoDamage", 0,
											   "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2595);
			nox_swprintf(v75, v13);
			break;
		case 1:
			v52 = nox_strman_loadString_40F1D0("IdentifyDurabilitySlight", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c",
											   2599);
			nox_swprintf(v75, v52);
			break;
		case 2:
			v53 = nox_strman_loadString_40F1D0("IdentifyDurabilityModerate", 0,
											   "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2603);
			nox_swprintf(v75, v53);
			break;
		case 3:
			v13 = nox_strman_loadString_40F1D0("IdentifyDurabilitySevere", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c",
											   2607);
			nox_swprintf(v75, v13);
			break;
		case 4:
			v51 = nox_strman_loadString_40F1D0("IdentifyDurabilityIndestructable", 0,
											   "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2591);
			nox_swprintf(v75, v51);
			break;
		default:
			break;
		}
	}
	nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
	nox_window_call_field_94(v10, 16397, (uintptr_t)getMemAt(0x5D4594, 1063656), -1);
	v64 = dword_5d4594_1063116->field_74_3;
	v14 = nox_strman_loadString_40F1D0("IdentifyWeight", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2620);
	nox_swprintf(v75, v14, v64);
	nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
	nox_window_call_field_94(v10, 16397, (uintptr_t)getMemAt(0x5D4594, 1063660), -1);
	v15 = dword_5d4594_1063116;
	v16 = dword_5d4594_1063116->flags28;
	if (!(v16 & 0x2000000)) {
		if (!(v16 & 0x1001000)) {
			goto LABEL_72;
		}
		v23 = (nox_playerInfo*)dword_8531A0_2576;
		v27 = dword_5d4594_1063116->item_modifiers[0];
		v69 = 1.0;
		if (!*getMemU32Ptr(0x5D4594, 1063644)) {
			*getMemU32Ptr(0x5D4594, 1063644) = nox_xxx_getTTByNameSpriteMB_44CFC0("ArcherArrow");
			v24 = nox_xxx_getTTByNameSpriteMB_44CFC0("ArcherBolt");
			v23 = (nox_playerInfo*)dword_8531A0_2576;
			*getMemU32Ptr(0x5D4594, 1063648) = v24;
			v15 = dword_5d4594_1063116;
		}
		if (!v23 || !(v15->flags29 & 2)) {
			goto LABEL_50;
		}
		v25 = v23->field_4;
		if (v25 & 4) {
			v26 = nox_xxx_getProjectileClassById_413250(*getMemIntPtr(0x5D4594, 1063644));
			v2 = 4;
		} else {
			if (!(v25 & 8)) {
				goto LABEL_50;
			}
			v26 = nox_xxx_getProjectileClassById_413250(*getMemIntPtr(0x5D4594, 1063648));
			v2 = 8;
		}
		if (v26) {
			goto LABEL_51;
		}
		v15 = dword_5d4594_1063116;
	LABEL_50:
		v26 = nox_xxx_getProjectileClassById_413250(v15->field_27);
	LABEL_51:
		v71 = sub_4626C0(dword_5d4594_1063116);
		v72 = sub_462700(dword_5d4594_1063116);
		if (v27 && nox_modifier_effect_getAttackFunc(v27) == (void*)nox_xxx_effectDamageMultiplier_4E04C0) {
			v69 = nox_modifier_effect_getAttackFloat(v27);
		}
		v28 = nox_xxx_calcBoltDamage_4EF1E0(v23->info.field_2239, v26);
		v29 = nox_xxx_boltDamageModifierType_4EF1E0(v26);
		v70 = v28 * v69 + v71 + v72;
		if (v29 == *getMemU32Ptr(0x5D4594, 1063648) && nox_common_gameFlags_check_40A5C0(2048)) {
			v30 = nox_xxx_gamedataGetFloat_419D40((void*)getMemAt(0x587000, 137632));
		} else {
			v30 = (double)(int32_t)nox_xxx_boltDamageModifierMinimum_4EF1E0(v26);
		}
		v68 = v30 * v69;
		v31 = v70 - v68 - v72 - v71;
		v69 = v31;
		if (v31 < 0.0) {
			v32 = v70 - v69;
			v69 = 0.0;
			v70 = v32;
		}
		v33 = dword_5d4594_1063116->flags29;
		if (v33 & 0xC) {
			v34 =
				nox_strman_loadString_40F1D0("WeaponDamageLabelNA", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2770);
			nox_swprintf(v75, v34);
		} else if (!(v33 & 2) || v2) {
			v58 = v70;
			v55 = nox_strman_loadString_40F1D0("WeaponDamageLabel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2776);
			nox_swprintf(v75, v55, v58);
		} else {
			v57 = v70;
			v54 = nox_strman_loadString_40F1D0("WeaponDamageLabelUnknownPlus", 0,
											   "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2773);
			nox_swprintf(v75, v54, v57);
		}
		nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
		nox_wcsncpy(v75, L"  ", sizeof(v75)/2);
		v59 = v68;
		v35 = nox_strman_loadString_40F1D0("BaseDamageLabel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2785);
		nox_swprintf(v76, v35, v59);
		nox_wcscat(v75, v76);
		nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
		nox_wcsncpy(v75, L"  ", sizeof(v75)/2);
		v60 = v69;
		v36 = nox_strman_loadString_40F1D0("StrengthDamageLabel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2792);
		nox_swprintf(v76, v36, v60);
		nox_wcscat(v75, v76);
		nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
		if (v72 > 0.0) {
			nox_wcsncpy(v75, L"  ", sizeof(v75)/2);
			v61 = v72;
			v37 = nox_strman_loadString_40F1D0("FireDamageLabel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2801);
			nox_swprintf(v76, v37, v61);
			nox_wcscat(v75, v76);
			nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
		}
		if (v71 > 0.0) {
			nox_wcsncpy(v75, L"  ", sizeof(v75)/2);
			v62 = v71;
			v38 = nox_strman_loadString_40F1D0("ElectricalDamageLabel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c",
											   2811);
			nox_swprintf(v76, v38, v62);
			nox_wcscat(v75, v76);
			nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
		}
		nox_window_call_field_94(v10, 16397, (uintptr_t)getMemAt(0x5D4594, 1063668), -1);
		goto LABEL_71;
	}
	v17 = nox_xxx_equipClothFindDefByTT_413270(dword_5d4594_1063116->field_27);
	v18 = 1.0;
	v19 = dword_5d4594_1063116->item_modifiers[0];
	if (v19 && nox_modifier_effect_getDefendFunc(v19) == (void*)sub_4E0370) {
		v18 = nox_modifier_effect_getDefendFloat(v19);
	}
	v65 = v18 * nox_modifier_getArmor(v17) * 1000.0 + 0.5;
	v20 = nox_float2int(v65);
	if (dword_5d4594_1063116->flags29 & 2) {
		v21 = nox_strman_loadString_40F1D0("ArmorValueLabelNA", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2647);
		nox_swprintf(v75, v21);
	} else {
		v66 = v20;
		v22 = nox_strman_loadString_40F1D0("ArmorValueLabel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2649);
		nox_swprintf(v75, v22, v66);
	}
	nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
	nox_window_call_field_94(v10, 16397, (uintptr_t)getMemAt(0x5D4594, 1063664), -1);
LABEL_71:
	v15 = dword_5d4594_1063116;
LABEL_72:
	v39 = v15->flags28;
	if (v39 & 0x13001000) {
		if (v39 & 0x11001000) {
			v40 = nox_xxx_getProjectileClassById_413250(v15->field_27);
		} else {
			v40 = nox_xxx_equipClothFindDefByTT_413270(v15->field_27);
		}
		if (v40) {
			v15 = dword_5d4594_1063116;
			v43 = dword_5d4594_1063116->item_modifiers;
			if (dword_5d4594_1063116->flags28 & 0x10000000) {
				goto LABEL_91;
			}
			v44 = v43[2];
			if (v44 && nox_modifier_effect_getIdentDescription(v44)) {
				v45 = nox_strman_loadString_40F1D0("IdentifySpecialAttributes", 0,
												   "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2851);
				nox_swprintf(v75, v45);
				nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
				v46 = 0;
				nox_wcsncpy(v75, L"  ", sizeof(v75)/2);
				nox_swprintf(v76, nox_modifier_effect_getIdentDescription(v44));
				nox_wcscat(v75, v76);
				nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
				v15 = dword_5d4594_1063116;
			} else {
				v46 = v73;
			}
			v47 = v43[3];
			if (v47 && nox_modifier_effect_getIdentDescription(v47)) {
				if (v46 == 1) {
					v48 = nox_strman_loadString_40F1D0("IdentifySpecialAttributes", 0,
													   "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2868);
					nox_swprintf(v75, v48);
					nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
					v73 = 0;
					v46 = 0;
				}
				nox_wcsncpy(v75, L"  ", sizeof(v75)/2);
				nox_swprintf(v76, nox_modifier_effect_getIdentDescription(v47));
				nox_wcscat(v75, v76);
				nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
				v15 = dword_5d4594_1063116;
			}
			if (v46) {
				goto LABEL_91;
			}
			nox_window_call_field_94(v10, 16397, (uintptr_t)getMemAt(0x5D4594, 1063672), -1);
		} else {
			v41 = nox_strman_loadString_40F1D0("IdentifySpecialAttributes", 0,
											   "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2835);
			nox_swprintf(v75, v41);
			nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
			v42 = nox_strman_loadString_40F1D0("IdentifyUnknown", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2837);
			nox_swprintf(v75, v42);
			nox_window_call_field_94(v10, 16397, (uintptr_t)v75, -1);
		}
		v15 = dword_5d4594_1063116;
	}
LABEL_91:
	v49 = nox_get_thing_desc(v15->field_27);
	if (v49) {
		nox_window_call_field_94(v10, 16397, (uintptr_t)v49, -1);
	}
	v67 = nox_get_thing_pretty_image(dword_5d4594_1063116->field_27);
	v50 = nox_xxx_wndGetChildByID_46B0C0(nox_inventory_identify_window, 9155);
	return nox_xxx_wndSetIcon_46AE60(v50, v67);
}
//----- (004649B0) --------------------------------------------------------
int sub_4649B0(nox_drawable* drawable, int column, int row) {
	if (!drawable || !sub_464B40(column, row)) {
		return 0;
	}
	nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[row + NOX_INVENTORY_ROW_COUNT * column];
	uint8_t count = cell->field_140;
	if (count) {
		if (drawable->flags28 & 0x4000000) {
			return 0;
		}
		if (!cell->field_0 || cell->field_0->field_27 != drawable->field_27) {
			return 0;
		}
	}
	if (count >= 0x20u) {
		return 0;
	}
	if (!count) {
		cell->field_0 = nox_new_drawable_for_thing(drawable->field_27);
		if (!cell->field_0) {
			wchar2_t* message = nox_strman_loadString_40F1D0(
				"DrawablesExhausted", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 898);
			nox_xxx_printCentered_445490(message);
			return 0;
		}
		cell->field_0->flags30 |= 0x40000000u;
		for (size_t i = 0; i < 4; i++) {
			cell->field_0->item_modifiers[i] = drawable->item_modifiers[i];
		}
		cell->field_0->item_field_112_0 = drawable->item_field_112_0;
		cell->field_0->item_field_112_2 = drawable->item_field_112_2;
		cell->field_0->field_113 = drawable->field_113;
		cell->field_0->field_73_1 = drawable->field_73_1;
		cell->field_0->field_73_2 = drawable->field_73_2;
	}
	(&cell->field_4)[cell->field_140++] = drawable->field_32;
	cell->field_132 = 0;
	for (size_t slot = 0; slot < 9; slot++) {
		for (nox_drawable* equipped = (nox_drawable*)array_5D4594_1049872[slot]; equipped;
			 equipped = equipped->field_92) {
			if (equipped->field_32 == drawable->field_32) {
				cell->field_132 = 1;
				if (cell->field_136) {
					nox_xxx_clientSetAltWeapon_461550(NULL);
					cell->field_136 = 0;
				}
				return 1;
			}
		}
	}
	return 1;
}

//----- (00464BD0) --------------------------------------------------------
int sub_464BD0(nox_window* win, int a2, uintptr_t a3, uintptr_t a4) {
	(void)win;
	(void)a4;
	int v4;          // eax
	int v5;          // eax
	int v6;          // eax
	int v7;          // eax
	int v8;          // eax
	int v9;          // esi
	int v10;         // edi
	int v14;         // eax
	int v15;         // eax
	int v16;         // esi
	int v17;         // edi
	nox_drawable* v19; // ecx
	int v20;         // eax
	int v21;         // eax
	int v26;         // eax
	uint32_t* v28;   // esi
	wchar2_t* v29;    // eax
	int v30;         // eax
	int v31;         // eax
	int v32;         // esi
	int v33;         // edi
	int v34;         // eax
	int v36;         // eax
	int v37;         // eax
	int v38;         // eax
	nox_drawable* v40; // ecx
	int v42;         // edx
	int v45;         // esi
	int v47;         // eax
	int v48;         // esi
	const nox_modifier_attrs_t* v49; // edi
	wchar2_t* v50;    // eax
	int2 v51;        // [esp-24h] [ebp-7Ch]
	int v52;         // [esp-1Ch] [ebp-74h]
	int v53;         // [esp-18h] [ebp-70h]
	int v54;         // [esp-8h] [ebp-60h]
	int v55;         // [esp-4h] [ebp-5Ch]
	int2 v56;        // [esp+8h] [ebp-50h]
	int2 v57;        // [esp+10h] [ebp-48h]
	int2 v58;        // [esp+18h] [ebp-40h]
	int2 v59;        // [esp+20h] [ebp-38h]
	nox_drawable* dragged;

	v59.field_4 = a3 >> 16;
	v59.field_0 = (unsigned short)a3;
	sub_463370(nox_inventory_window, &v59, &v56);
	if (sub_45D9B0() || getMemByte(0x5D4594, 1049868) != 2) {
		return 1;
	}
	switch (a2) {
	case 5:
		if (nox_xxx_playerAnimCheck_4372B0()) {
			return 1;
		}
		if (dword_5d4594_1049864 == 5) {
			v8 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136352));
			if (v8) {
				v9 = (v56.field_0 - 314) / 50;
				v10 = (dword_5d4594_1062512 + v56.field_4 - 13) / 50;
				if (!sub_464B40(v9, v10)) {
					return 1;
				}
				int v11 = v10 + NOX_INVENTORY_ROW_COUNT * v9;
				if (nox_client_inventory_grid_1050020[v11].field_140) {
					nox_drawable* dr = nox_client_inventory_grid_1050020[v11].field_0;
					dword_5d4594_1063116 = dr;
					dr->field_32 = nox_client_inventory_grid_1050020[v11].field_4;
				} else {
					dword_5d4594_1063116 = 0;
				}
				return 1;
			}
			if (sub_478030()) {
				if (sub_479870()) {
					LOBYTE(v14) = sub_479880(&v56);
					if (v14) {
						dword_5d4594_1063116 = sub_4798A0(&v56);
						return 1;
					}
				}
			}
		} else if (dword_5d4594_1049864 == 6) {
			v15 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136352));
			if (v15) {
				v16 = (v56.field_0 - 314) / 50;
				v17 = (dword_5d4594_1062512 + v56.field_4 - 13) / 50;
				if (sub_464B40(v16, v17)) {
					int v18 = v17 + NOX_INVENTORY_ROW_COUNT * v16;
					if (nox_client_inventory_grid_1050020[v18].field_140) {
						v19 = nox_client_inventory_grid_1050020[v18].field_0;
						v20 = nox_client_inventory_grid_1050020[v18].field_4;
						if (v19) {
							v19->field_32 = v20;
							nox_xxx_trade_4657B0(v20);
							return 1;
						}
					}
				}
			}
		} else if (getMemByte(0x5D4594, 1049870) != 1 ||
				   (v21 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136336)), v21 != 1)) {
			if (getMemByte(0x5D4594, 1049869) || nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136384)) ||
				nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136400))) {
				if (nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136384)) == 1) {
					return 0;
				}
				if (nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136400)) == 1) {
					return 0;
				}
			} else {
				nox_xxx_wndSetCaptureMain_46ADC0(nox_inventory_window);
				if (sub_479590() == 3) {
					nox_xxx_clientTradeMB_4657E0(&v56);
				} else {
					sub_4658A0(nox_inventory_window, &v56);
				}
				dragged = nox_client_inventory_get_dragged();
				if (dragged) {
					nox_xxx_cursorSetDraggedItem_477690(dragged);
					nox_xxx_setKeybTimeout_4160D0(0);
					*(int2*)getMemAt(0x5D4594, 1062572) = v56;
					nox_xxx_clientPlaySoundSpecial_452D80(791, 100);
					return 1;
				}
			}
		}
		return 1;
	case 7:
		if (nox_xxx_playerAnimCheck_4372B0() || dword_5d4594_1049864 == 6) {
			return 1;
		}
		if (!getMemByte(0x5D4594, 1049869)) {
			v26 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136368));
			if (v26) {
				if ((v56.field_4 - 13) / 50 == 1) {
					if (dword_5d4594_1049864 != 5) {
						sub_465CA0();
						return 1;
					}
					sub_462740();
					return 1;
				}
			}
		}
		// fallthrough
	case 6:
		if (nox_xxx_playerAnimCheck_4372B0() || dword_5d4594_1049864 == 6) {
			return 1;
		}
		int v43 = 0;
		if (dword_5d4594_1049864 == 5) {
			if (nox_xxx_cursorGetTypePrev_477630() == 7) {
				sub_462740();
				return 1;
			}
		} else {
			nox_xxx_wndClearCaptureMain_46ADE0(nox_inventory_window);
		}
		if (dword_5d4594_1049864 == 4) {
			v58 = v59;
			sub_473970(&v58, &v58);
			v28 = nox_drawable_find_49ABF0(&v58, 20);
			if (v28) {
				v57.field_0 = nox_win_width / 2;
				v57.field_4 = nox_win_height / 2;
				sub_473970(&v57, &v57);
				if ((v57.field_0 - v28[3]) * (v57.field_0 - v28[3]) + (v57.field_4 - v28[4]) * (v57.field_4 - v28[4]) <=
					5625) {
					dword_5d4594_1049864 = 0;
					return 1;
				}
				v29 = nox_strman_loadString_40F1D0("ObjectTooFar", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 3858);
			} else {
				v29 = nox_strman_loadString_40F1D0("NoObject", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 3869);
			}
			nox_xxx_printCentered_445490(v29);
			dword_5d4594_1049864 = 0;
			return 1;
		}
		dragged = nox_client_inventory_get_dragged();
		if (!dragged) {
			return 1;
		}
		if (!nox_xxx_wndPointInWnd_46AAB0(nox_inventory_window, v59.field_0, v59.field_4) ||
			(v30 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136384)), v30) ||
			(v31 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136400)), v31)) {
			v58 = v59;
			sub_473970(&v58, &v57);
			if (dword_5d4594_1049856 == 1) {
				if (!sub_4C12C0()) {
					nox_xxx_clientDrop_465BE0(&v57);
				}
			} else {
				v47 = dword_5d4594_1049800_inventory_click_row_index +
					  14 * dword_5d4594_1049796_inventory_click_column_index +
					  7 * dword_5d4594_1049796_inventory_click_column_index;
				v48 = nox_client_inventory_grid_1050020[v47].field_140;
				if (nox_client_inventory_grid_1050020[v47].field_140) {
					v49 = 0;
					nox_xxx_wndClearCaptureMain_46ADE0(nox_inventory_window);
					if (dragged->flags28 & 0x13001000) {
						v49 = (const nox_modifier_attrs_t*)dragged->item_modifiers;
					}
					sub_4C05F0(0, 0);
					v53 = dragged->field_27;
					v52 = dragged->field_32;
					v51 = v58;
					v50 = nox_strman_loadString_40F1D0("DropLabel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 4148);
					nox_gui_itemAmountDialog_4C0430(v50, v51.field_0, v51.field_4, v52, v53, v49, v48 + 1, 0,
												sub_465CD0, 0);
				} else if (!sub_4C12C0()) {
					nox_xxx_clientDrop_465BE0(&v57);
				}
			}
			if (dword_5d4594_1049856) {
				goto LABEL_121;
			}
			v55 = dword_5d4594_1049800_inventory_click_row_index;
			v54 = dword_5d4594_1049796_inventory_click_column_index;
			sub_4649B0(dragged, v54, v55);
			goto LABEL_121;
		}
		v32 = *getMemU32Ptr(0x5D4594, 1062572) - v56.field_0;
		v33 = *getMemU32Ptr(0x5D4594, 1062576) - v56.field_4;
		if (!nox_xxx_checkKeybTimeout_4160F0(0, gameFPS() / 3u) && v32 * v32 + v33 * v33 < 100) {
			v34 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136352));
			if (!v34) {
				goto LABEL_121;
			}
			if (!sub_4C12C0()) {
				if (dragged->flags28 & 0x3001000) {
					int v35 = dword_5d4594_1049800_inventory_click_row_index +
							  NOX_INVENTORY_ROW_COUNT * dword_5d4594_1049796_inventory_click_column_index;
					if (nox_client_inventory_grid_1050020[v35].field_136) {
						nox_xxx_clientSetAltWeapon_461550(0);
						nox_client_inventory_grid_1050020[v35].field_136 = 0;
					} else if (nox_client_inventory_grid_1050020[v35].field_132) {
						nox_xxx_clientDequip_464B70(dragged);
					} else {
						nox_xxx_clientKeyEquip_465C30(*(int*)&dword_5d4594_1049796_inventory_click_column_index,
													  *(int*)&dword_5d4594_1049800_inventory_click_row_index);
					}
				} else {
					nox_xxx_clientUse_465C70(dragged);
				}
			}
			sub_4649B0(dragged, *(int*)&dword_5d4594_1049796_inventory_click_column_index,
					   *(int*)&dword_5d4594_1049800_inventory_click_row_index);
			goto LABEL_121;
		}
		v36 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136336));
		if (v36 && !getMemByte(0x5D4594, 1049870)) {
			if (!dword_5d4594_1049856) {
				nox_xxx_clientEquip_4623B0(dragged);
				sub_4649B0(dragged, *(int*)&dword_5d4594_1049796_inventory_click_column_index,
						   *(int*)&dword_5d4594_1049800_inventory_click_row_index);
			}
			goto LABEL_121;
		}
		v37 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136352));
		if (!v37) {
			v55 = dword_5d4594_1049800_inventory_click_row_index;
			v54 = dword_5d4594_1049796_inventory_click_column_index;
			sub_4649B0(dragged, v54, v55);
			goto LABEL_121;
		}
		v38 = dragged->field_27;
		if (v38 == dword_5d4594_1062560 || v38 == *getMemU32Ptr(0x5D4594, 1049728) ||
			v38 == *getMemU32Ptr(0x5D4594, 1049724) || v38 == dword_5d4594_1062556 || v38 == dword_5d4594_1062564) {
			sub_4649B0(dragged, *(int*)&dword_5d4594_1049796_inventory_click_column_index,
					   *(int*)&dword_5d4594_1049800_inventory_click_row_index);
			goto LABEL_121;
		}
		dword_5d4594_1049804 = (v56.field_0 - 314) / 50;
		dword_5d4594_1049808 = (dword_5d4594_1062512 + v56.field_4 - 13) / 50;
		if (!sub_464B40((v56.field_0 - 314) / 50, (dword_5d4594_1062512 + v56.field_4 - 13) / 50)) {
			goto LABEL_121;
		}
		if (dword_5d4594_1049856) {
			int v39 = dword_5d4594_1049808 + NOX_INVENTORY_ROW_COUNT * dword_5d4594_1049804;
			if (nox_client_inventory_grid_1050020[v39].field_140 &&
				(v40 = nox_client_inventory_grid_1050020[v39].field_0) != NULL &&
				(((v40->flags28 & 0x2000000) && (dragged->flags28 & 0x2000000) &&
				  v40->flags29 == dragged->flags29) ||
				 ((v40->flags28 & 0x1001000) && (dragged->flags28 & 0x1001000)))) {
				v42 = nox_client_inventory_grid_1050020[v39].field_4;
				*getMemU32Ptr(0x5D4594, 1049860) = 1;
				v40->field_32 = v42;
				nox_xxx_clientEquip_4623B0(v40);
			} else {
				*getMemU32Ptr(0x5D4594, 1049860) = 1;
				nox_xxx_clientDequip_464B70(dragged);
			}
			goto LABEL_121;
		}
		if (nox_client_inventory_grid_1050020[dword_5d4594_1049800_inventory_click_row_index +
								NOX_INVENTORY_ROW_COUNT * dword_5d4594_1049796_inventory_click_column_index]
				.field_140) {
			v55 = dword_5d4594_1049800_inventory_click_row_index;
			v54 = dword_5d4594_1049796_inventory_click_column_index;
			sub_4649B0(dragged, v54, v55);
			goto LABEL_121;
		}
		if (!sub_4649B0(dragged, *(int*)&dword_5d4594_1049804, *(int*)&dword_5d4594_1049808)) {
			sub_4649B0(dragged, *(int*)&dword_5d4594_1049796_inventory_click_column_index,
					   *(int*)&dword_5d4594_1049800_inventory_click_row_index);
			goto LABEL_121;
		}
		nox_xxx_clientPlaySoundSpecial_452D80(792, 100);
		v43 = dword_5d4594_1049800_inventory_click_row_index +
				  NOX_INVENTORY_ROW_COUNT * dword_5d4594_1049796_inventory_click_column_index;
		v45 = nox_client_inventory_grid_1050020[v43].field_136;
		if (v45) {
			int v46 = dword_5d4594_1049808 + NOX_INVENTORY_ROW_COUNT * dword_5d4594_1049804;
			nox_client_inventory_grid_1050020[v46].field_136 = v45;
			nox_client_inventory_grid_1050020[v43].field_136 = 0;
			dword_5d4594_1062480 = &nox_client_inventory_grid_1050020[v46];
		}
		sub_461B50();
	LABEL_121:
		nox_xxx_cursorResetDraggedItem_4776A0();
		if (!dword_5d4594_1049856) {
			nox_xxx_spriteDelete_45A4B0(dragged);
		}
		nox_client_inventory_set_dragged(NULL);
		dword_5d4594_1049856 = 0;
		return 1;
	case 8:
		return 1;
	case 9:
		if (dword_5d4594_1049864 == 5) {
			sub_462740();
			return 1;
		}
		return 0;
	case 19:
		if (nox_xxx_playerAnimCheck_4372B0()) {
			return 1;
		}
		v6 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136384));
		if (v6) {
			if (dword_5d4594_1049864 == 5) {
				return 1;
			}
			return 0;
		}
		v7 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136400));
		if (v7) {
			if (dword_5d4594_1049864 == 5) {
				return 1;
			}
			return 0;
		}
		nox_window_call_field_94(nox_inventory_window, 16391, (uintptr_t)nox_inventory_scroll_up_button, 0);
		return 1;
	case 20:
		if (nox_xxx_playerAnimCheck_4372B0()) {
			return 1;
		}
		v4 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136384));
		if (v4) {
			if (dword_5d4594_1049864 == 5) {
				return 1;
			}
			return 0;
		}
		v5 = nox_xxx_pointInRect_4281F0(&v56, (int4*)getMemAt(0x587000, 136400));
		if (v5) {
			if (dword_5d4594_1049864 == 5) {
				return 1;
			}
			return 0;
		}
		nox_window_call_field_94(nox_inventory_window, 16391, (uintptr_t)nox_inventory_scroll_down_button, 0);
		return 1;
	default:
		if (dword_5d4594_1049864 == 5) {
			return 1;
		}
		return 0;
	}
}
