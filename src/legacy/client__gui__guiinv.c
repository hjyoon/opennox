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
extern nox_window* nox_inventory_window;
extern nox_window* nox_inventory_identify_window;
extern nox_window* nox_inventory_scroll_up_button;
extern nox_window* nox_inventory_scroll_down_button;
extern void* nox_inventory_font;

extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_red_2589776;
extern uint32_t nox_color_blue_2650684;
extern uint32_t nox_color_cyan_2649820;
extern uint32_t nox_color_yellow_2589772;
extern uint32_t nox_color_violet_2598268;
extern uint32_t nox_color_black_2650656;
extern uint32_t nox_color_orange_2614256;

extern nox_inventory_cell_t nox_client_inventory_grid_1050020[NOX_INVENTORY_CELLS_MAX];

//----- (00461660) --------------------------------------------------------
int nox_xxx_spritePickup_461660(int net_code, int thing_type, const nox_modifier_attrs_t* attrs) {
	nox_inventory_cell_t* v3; // eax
	wchar2_t* v4; // eax
	int v6;      // ecx
	int v7;      // edx
	nox_thing* v8; // eax
	int2 a4;     // [esp+8h] [ebp-8h]

	if (thing_type != dword_5d4594_1062560 && thing_type != *getMemU32Ptr(0x5D4594, 1049728) &&
		thing_type != *getMemU32Ptr(0x5D4594, 1049724) && thing_type != dword_5d4594_1062556 &&
		thing_type != dword_5d4594_1062564) {
		v3 = sub_461970(net_code, thing_type);
		if (v3) {
			if (v3->field_0->flags28 & 0x10) {
				sub_472310();
			}
		} else {
			if (!sub_4617C0(net_code, thing_type, attrs, &a4)) {
				v4 = nox_strman_loadString_40F1D0("InventoryFull", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 682);
				nox_xxx_printCentered_445490(v4);
				return 0;
			}
			v6 = a4.field_0;
			v7 = a4.field_4;
			if (nox_client_inventory_grid_1050020[a4.field_4 + NOX_INVENTORY_ROW_COUNT * a4.field_0].field_0->flags28 & 0x10) {
				sub_472310();
				v7 = a4.field_4;
				v6 = a4.field_0;
			}
			if (nox_client_inventory_grid_1050020[v7 + NOX_INVENTORY_ROW_COUNT * v6].field_0->flags28 & 0x3001000) {
				dword_5d4594_1062516 = 0;
				if (v7 >= 3) {
					dword_5d4594_1062516 = 10 * (5 * v7 - 10);
				}
			}
		}
		v8 = nox_get_thing(thing_type);
		if (v8) {
			if (v8->pri_class & 0x1001000) {
				sub_4673F0(a4.field_0, a4.field_4);
			}
		}
	}
	return 1;
}

//----- (004617C0) --------------------------------------------------------
typedef struct {
	void* name;
	uint32_t type_index;
	void* description;
	uint8_t colors[24];
	uint32_t effectiveness;
	uint32_t material;
	uint32_t primary_enchant;
	uint32_t secondary_enchant;
	uint32_t durability;
	uint32_t field_56;
	uint16_t required_strength;
	uint8_t classes;
	uint8_t field_63;
} nox_projectile_class_header_t;
_Static_assert(offsetof(nox_projectile_class_header_t, classes) == (sizeof(void*) == 4 ? 62 : 74),
	"wrong native projectile class mask offset!");

int sub_4617C0(int net_code, int thing_type, const nox_modifier_attrs_t* attrs, int2* position) {
	for (int row = 0; row < 20; row++) {
		for (int column = 0; column < NOX_INVENTORY_COL_COUNT; column++) {
			nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[
				row + NOX_INVENTORY_ROW_COUNT * column];
			if (cell->field_140) {
				continue;
			}
			nox_drawable* drawable = nox_new_drawable_for_thing(thing_type);
			cell->field_0 = drawable;
			if (!drawable) {
				wchar2_t* message = nox_strman_loadString_40F1D0(
					"DrawablesExhausted", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 550);
				nox_xxx_printCentered_445490(message);
				return 0;
			}
			drawable->flags30 |= 0x40000000u;
			cell->field_4 = (uint32_t)net_code;
			cell->field_140 = 1;
			if (drawable->flags28 & 0x13001000) {
				for (int i = 0; i < 4; i++) {
					drawable->item_modifiers[i] = attrs->modifiers[i];
				}
				drawable->item_field_112_0 = (int16_t)attrs->field_16;
				drawable->item_field_112_2 = (int16_t)(attrs->field_16 >> 16);
			}
			if (position) {
				position->field_0 = column;
				position->field_4 = row;
			}
			if (sub_461930() && !dword_5d4594_1062480 &&
				(((drawable->flags28 & 0x1000000) && !(drawable->flags29 & 2)) ||
				 (drawable->flags28 & 0x1000))) {
				nox_projectile_class_header_t* projectile =
					(nox_projectile_class_header_t*)nox_xxx_getProjectileClassById_413250(drawable->field_27);
				nox_playerInfo* player = (nox_playerInfo*)dword_8531A0_2576;
				if (!projectile || (player && ((uint8_t)(1u << player->info.playerClass) & projectile->classes))) {
					nox_xxx_clientSetAltWeapon_461550(cell);
					cell->field_136 = 1;
				}
			}
			return 1;
		}
	}
	return 0;
}

//----- (00461A80) --------------------------------------------------------
void sub_461A80(int a1) {
	uint32_t stack_index = 0;
	nox_inventory_cell_t* cell = nox_inventory_find_cell_native_461EF0(a1, &stack_index);
	if (cell) {
		int refresh = cell->field_0 && (cell->field_0->flags28 & 0x10);
		sub_461E60(cell, stack_index);
		cell->field_132 = 0;
		sub_461B50();
		nox_drawable* equipped = sub_461F90(a1);
		if (equipped) {
			nox_xxx_spriteDelete_45A4B0((uint64_t*)equipped);
		}
		if (refresh) {
			sub_472310();
		}
	} else {
		nox_drawable* dragged = nox_client_inventory_get_dragged();
		if (dragged && dragged->field_32 == (uint32_t)a1) {
			if (dragged->flags28 & 0x10) {
				sub_472310();
			}
			nox_xxx_spriteDelete_45A4B0((uint64_t*)dragged);
			nox_client_inventory_set_dragged(NULL);
			dword_5d4594_1049856 = 0;
			nox_xxx_cursorResetDraggedItem_4776A0();
		} else {
			wchar2_t* message = nox_strman_loadString_40F1D0(
				"DroppedNotFound", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1439);
			nox_xxx_printCentered_445490(message);
		}
	}
}

//----- (00462040) --------------------------------------------------------
void sub_462040(int net_code) {
	uint32_t stack_index = 0;
	nox_inventory_cell_t* source_cell = nox_inventory_find_cell_native_461EF0(net_code, &stack_index);
	(void)stack_index;
	nox_drawable* source = source_cell ? source_cell->field_0 : NULL;
	if (!source_cell) {
		source = nox_client_inventory_get_dragged();
	}
	// GAME.EXE checks field_32 only for the transient dragged drawable. An
	// inventory drawable is identified by the cell's stacked net-code array;
	// requiring its otherwise unrelated field_32 rejects every normal equip.
	if (!source || (!source_cell && source->field_32 != (uint32_t)net_code)) {
		wchar2_t* message = nox_strman_loadString_40F1D0(
			"EquippedNotFound", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1605);
		nox_xxx_printCentered_445490(message);
		return;
	}
	int slot = sub_4622E0(source);
	if (slot == 9) {
		wchar2_t* message = nox_strman_loadString_40F1D0(
			"TooManyEquipped", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1701);
		nox_xxx_printCentered_445490(message);
		return;
	}
	nox_drawable* equipped = nox_new_drawable_for_thing(source->field_27);
	if (!equipped) {
		wchar2_t* message = nox_strman_loadString_40F1D0(
			"DrawablesExhausted", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1619);
		nox_xxx_printCentered_445490(message);
		return;
	}
	equipped->field_32 = (uint32_t)net_code;
	equipped->flags30 |= 0x40000000u;
	for (int i = 0; i < 4; i++) {
		equipped->item_modifiers[i] = source->item_modifiers[i];
	}
	equipped->item_field_112_0 = source->item_field_112_0;
	equipped->item_field_112_2 = source->item_field_112_2;
	equipped->field_113 = source->field_113;
	equipped->field_73_1 = source->field_73_1;
	equipped->field_73_2 = source->field_73_2;
	sub_4623E0(equipped, slot);
	if (source_cell) {
		source_cell->field_132 = 1;
		if (source_cell->field_136) {
			nox_xxx_clientSetAltWeapon_461550(NULL);
			source_cell->field_136 = 0;
		}
	}
	if ((equipped->flags28 & 0x1000000) && (equipped->flags29 & 0xC)) {
		nox_inventory_cell_t* preferred = NULL;
		if (dword_5d4594_1062488) {
			preferred = nox_inventory_find_cell_native_461EF0((int)dword_5d4594_1062488, NULL);
		}
		if (preferred && preferred->field_0) {
			preferred->field_0->field_32 = preferred->field_4;
			nox_xxx_clientEquip_4623B0(preferred->field_0);
		} else {
			for (int row = 0; row < NOX_INVENTORY_ROW_COUNT - 1; row++) {
				int found = 0;
				for (int column = 0; column < NOX_INVENTORY_COL_COUNT; column++) {
					nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[
						row + NOX_INVENTORY_ROW_COUNT * column];
					if (cell->field_140 && cell->field_0 &&
						(cell->field_0->flags28 & 0x1000000) && cell->field_0->flags29 == 2) {
						cell->field_0->field_32 = cell->field_4;
						nox_xxx_clientEquip_4623B0(cell->field_0);
						found = 1;
						break;
					}
				}
				if (found) {
					break;
				}
			}
		}
	}
	if (slot == 0) {
		dword_5d4594_1062488 = equipped->field_32;
	}
	if (equipped->item_field_112_0 >= 0) {
		sub_470D90(equipped->item_field_112_0, equipped->item_field_112_2);
	}
	if (dword_5d4594_1062496) {
		nox_inventory_cell_t* alternate =
			nox_inventory_find_cell_native_461EF0((int)dword_5d4594_1062496, NULL);
		if (alternate) {
			alternate->field_136 = 1;
			nox_xxx_clientSetAltWeapon_461550(alternate);
			dword_5d4594_1062496 = 0;
		}
	}
}

//----- (00462740) --------------------------------------------------------
int sub_462740() {
	wchar2_t* v0;  // eax
	uint32_t* v1; // eax

	if (wndIsShown_nox_xxx_wndIsShown_46ACC0(nox_inventory_identify_window)) {
		return 0;
	}
	nox_window_set_hidden(nox_inventory_identify_window, 1);
	dword_5d4594_1063116 = 0;
	dword_5d4594_1063120 = 0;
	v0 = nox_strman_loadString_40F1D0("thing.db:IdentifyDescription", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c",
									  2361);
	nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1063124), v0);
	v1 = nox_xxx_wndGetChildByID_46B0C0(nox_inventory_identify_window, 9156);
	nox_window_call_field_94((int)v1, 16399, 0, 0);
	nox_xxx_wndClearCaptureMain_46ADE0(nox_inventory_window);
	dword_5d4594_1049864 = 0;
	nox_client_setCursorType_477610(0);
	return 1;
}

//----- (004627F0) --------------------------------------------------------
uintptr_t sub_4627F0(uint32_t* a1) {
	int v2;           // ebx
	int v3;           // eax
	uintptr_t result; // eax
	wchar2_t* v5;      // eax
	uint32_t* v6;     // eax
	wchar2_t* v7;      // eax
	wchar2_t* v8;      // eax
	uint32_t* v9;     // eax
	uint32_t* v10;    // esi
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
	uintptr_t v23;    // ecx
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
	int v49;          // eax
	uint32_t* v50;    // eax
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
			result = nox_window_call_field_94((int)v6, 16399, 0, 0);
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
	sub_46AEE0((nox_window*)v9, (const wchar2_t*)getMemAt(0x5D4594, 1063124));
	v10 = nox_xxx_wndGetChildByID_46B0C0(nox_inventory_identify_window, 9156);
	nox_window_call_field_94((int)v10, 16399, 0, 0);
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
	nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
	nox_window_call_field_94((int)v10, 16397, (int)getMemAt(0x5D4594, 1063656), -1);
	v64 = dword_5d4594_1063116->field_74_3;
	v14 = nox_strman_loadString_40F1D0("IdentifyWeight", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2620);
	nox_swprintf(v75, v14, v64);
	nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
	nox_window_call_field_94((int)v10, 16397, (int)getMemAt(0x5D4594, 1063660), -1);
	v15 = dword_5d4594_1063116;
	v16 = dword_5d4594_1063116->flags28;
	if (!(v16 & 0x2000000)) {
		if (!(v16 & 0x1001000)) {
			goto LABEL_72;
		}
		v23 = dword_8531A0_2576;
		v27 = dword_5d4594_1063116->item_modifiers[0];
		v69 = 1.0;
		if (!*getMemU32Ptr(0x5D4594, 1063644)) {
			*getMemU32Ptr(0x5D4594, 1063644) = nox_xxx_getTTByNameSpriteMB_44CFC0("ArcherArrow");
			v24 = nox_xxx_getTTByNameSpriteMB_44CFC0("ArcherBolt");
			v23 = dword_8531A0_2576;
			*getMemU32Ptr(0x5D4594, 1063648) = v24;
			v15 = dword_5d4594_1063116;
		}
		if (!v23 || !(v15->flags29 & 2)) {
			goto LABEL_50;
		}
		v25 = *(uint32_t*)(v23 + 4);
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
		v28 = nox_xxx_calcBoltDamage_4EF1E0(*(uint32_t*)(v23 + 2239), v26);
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
		nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
		nox_wcsncpy(v75, L"  ", sizeof(v75)/2);
		v59 = v68;
		v35 = nox_strman_loadString_40F1D0("BaseDamageLabel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2785);
		nox_swprintf(v76, v35, v59);
		nox_wcscat(v75, v76);
		nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
		nox_wcsncpy(v75, L"  ", sizeof(v75)/2);
		v60 = v69;
		v36 = nox_strman_loadString_40F1D0("StrengthDamageLabel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2792);
		nox_swprintf(v76, v36, v60);
		nox_wcscat(v75, v76);
		nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
		if (v72 > 0.0) {
			nox_wcsncpy(v75, L"  ", sizeof(v75)/2);
			v61 = v72;
			v37 = nox_strman_loadString_40F1D0("FireDamageLabel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2801);
			nox_swprintf(v76, v37, v61);
			nox_wcscat(v75, v76);
			nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
		}
		if (v71 > 0.0) {
			nox_wcsncpy(v75, L"  ", sizeof(v75)/2);
			v62 = v71;
			v38 = nox_strman_loadString_40F1D0("ElectricalDamageLabel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c",
											   2811);
			nox_swprintf(v76, v38, v62);
			nox_wcscat(v75, v76);
			nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
		}
		nox_window_call_field_94((int)v10, 16397, (int)getMemAt(0x5D4594, 1063668), -1);
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
	nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
	nox_window_call_field_94((int)v10, 16397, (int)getMemAt(0x5D4594, 1063664), -1);
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
				nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
				v46 = 0;
				nox_wcsncpy(v75, L"  ", sizeof(v75)/2);
				nox_swprintf(v76, nox_modifier_effect_getIdentDescription(v44));
				nox_wcscat(v75, v76);
				nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
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
					nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
					v73 = 0;
					v46 = 0;
				}
				nox_wcsncpy(v75, L"  ", sizeof(v75)/2);
				nox_swprintf(v76, nox_modifier_effect_getIdentDescription(v47));
				nox_wcscat(v75, v76);
				nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
				v15 = dword_5d4594_1063116;
			}
			if (v46) {
				goto LABEL_91;
			}
			nox_window_call_field_94((int)v10, 16397, (int)getMemAt(0x5D4594, 1063672), -1);
		} else {
			v41 = nox_strman_loadString_40F1D0("IdentifySpecialAttributes", 0,
											   "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2835);
			nox_swprintf(v75, v41);
			nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
			v42 = nox_strman_loadString_40F1D0("IdentifyUnknown", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2837);
			nox_swprintf(v75, v42);
			nox_window_call_field_94((int)v10, 16397, (int)v75, -1);
		}
		v15 = dword_5d4594_1063116;
	}
LABEL_91:
	v49 = nox_get_thing_desc(v15->field_27);
	if (v49) {
		nox_window_call_field_94((int)v10, 16397, v49, -1);
	}
	v67 = nox_get_thing_pretty_image(dword_5d4594_1063116->field_27);
	v50 = nox_xxx_wndGetChildByID_46B0C0(nox_inventory_identify_window, 9155);
	return nox_xxx_wndSetIcon_46AE60((int)v50, v67);
}
//----- (00463880) --------------------------------------------------------
void nox_client_makePlayerStatsDlg_463880(int* a1) {
	wchar2_t v77[256];

	int v1 = nox_xxx_guiFontHeightMB_43F320(0);
	int v2 = nox_xxx_guiFontHeightMB_43F320(nox_inventory_font);
	int v68 = v1 - v2;
	float v51 = (double)(v1 - v2) * 0.5 + 0.5;
	int v3 = nox_float2int(v51);
	int v73 = v3;
	int v72 = nox_color_white_2523948;
	int v6 = a1[0];
	int v7 = a1[1];
	int v4 = dword_8531A0_2576;
	if (!v4) {
		return;
	}
	sub_57B350();
	float4 v70a = nox_xxx_plrGetMaxVarsPtr_57B360(*(unsigned char*)(v4 + 2251));
	float4 v71a = nox_xxx_plrGetMaxVarsPtr_57B360(0);
	float* v70 = &v70a;
	float* v71 = &v71a;
	int v8 = v6 + 11;
	int v9 = v7 + 15;
	nox_xxx_drawSetTextColor_434390(v72);
	nox_client_drawSetColor_434460(nox_color_black_2650656);
	nox_client_drawRectFilledOpaque_49CE30(v8, v9, 200, 200);
	int v10 = v8 + 2;
	int v11 = v9 + 2 * v1 + 3;
	int v52 = *(char*)(v4 + 3684);
	wchar2_t* v12 = nox_strman_loadString_40F1D0("StatsLevel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1878);
	nox_swprintf(v77, v12, v52);
	nox_xxx_drawStringWrap_43FAF0(0, v77, v10, v11, 200, 0);
	int v13 = v11 + v1 + 1;
	if (nox_common_gameFlags_check_40A5C0(2048)) {
		int v53 = (long long)nox_xxx_gamedataGetFloatTable_419D70("XPTable", *(char*)(v4 + 3684) + 1);
		int v41 = *getMemU32Ptr(0x5D4594, 1062544);
		wchar2_t* v14 = nox_strman_loadString_40F1D0("StatsEXP", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1886);
		nox_swprintf(v77, v14, v41, v53);
		nox_xxx_drawStringWrap_43FAF0(0, v77, v10, v13, 200, 0);
	}
	int v15 = 2 * v1 + 2 + v13;
	wchar2_t* v16 = nox_strman_loadString_40F1D0("StatsHealth", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1896);
	nox_xxx_drawStringWrap_43FAF0(0, v16, v10, v15, 200, 0);
	nox_client_drawSetColor_434460(nox_color_violet_2598268);
	nox_client_drawRectFilledOpaque_49CE30(v10 + 60, v15, 90, v1);
	float v54 = (double)(int)(90 * *(uint32_t*)(v4 + 2247)) / *v70;
	int v67 = nox_float2int(v54);
	nox_client_drawSetColor_434460(nox_color_red_2589776);
	nox_client_drawRectFilledOpaque_49CE30(v10 + 60, v15, v67, v1);
	v68 = 90 * sub_470CC0();
	float v55 = (double)v68 / *v70;
	v67 = nox_float2int(v55);
	nox_client_drawSetColor_434460(*getMemIntPtr(0x85B3FC, 940));
	nox_client_drawRectFilledOpaque_49CE30(v10 + 60, v15, v67, v1);
	int v56 = nox_float2int(*v70);
	int v42 = *(uint32_t*)(v4 + 2247);
	wchar2_t* v17 = nox_strman_loadString_40F1D0("MinMaxFormat", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1914);
	nox_swprintf(v77, v17, v42, v56);
	nox_xxx_drawGetStringSize_43F840(nox_inventory_font, v77, &v67, 0, 0);
	float v69 = 0;
	LODWORD(v69) = v15 + v73;
	nox_xxx_drawStringWrap_43FAF0(nox_inventory_font, v77, v10 - v67 + 193, v15 + v73, 200, 0);
	int v18 = sub_470CC0();
	nox_swprintf(v77, L"%d", v18);
	nox_xxx_drawStringWrap_43FAF0(nox_inventory_font, v77, v10 + 45, SLODWORD(v69), 200, 0);
	int v19 = v15 + v1 + 1;
	if (*(uint8_t*)(v4 + 2251)) {
		nox_client_drawSetColor_434460(*getMemIntPtr(0x85B3FC, 944));
		nox_client_drawRectFilledOpaque_49CE30(v10 + 60, v19, 90, v1);
		v68 = 90 * *(uint32_t*)(v4 + 2243);
		float v57 = (double)v68 / v70[1];
		v67 = nox_float2int(v57);
		wchar2_t* v20 = nox_strman_loadString_40F1D0("StatsMana", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1941);
		nox_xxx_drawStringWrap_43FAF0(0, v20, v10, v19, 200, 0);
		nox_client_drawSetColor_434460(nox_color_blue_2650684);
		nox_client_drawRectFilledOpaque_49CE30(v10 + 60, v19, v67, v1);
		v68 = 90 * nox_xxx_cliGetMana_470DD0();
		float v58 = (double)v68 / v70[1];
		v67 = nox_float2int(v58);
		nox_client_drawSetColor_434460(nox_color_cyan_2649820);
		nox_client_drawRectFilledOpaque_49CE30(v10 + 60, v19, v67, v1);
		int v59 = nox_float2int(v70[1]);
		int v43 = *(uint32_t*)(v4 + 2243);
		wchar2_t* v21 = nox_strman_loadString_40F1D0("MinMaxFormat", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1952);
		nox_swprintf(v77, v21, v43, v59);
		nox_xxx_drawGetStringSize_43F840(nox_inventory_font, v77, &v67, 0, 0);
		nox_xxx_drawStringWrap_43FAF0(nox_inventory_font, v77, v10 - v67 + 193, v19 + v73, 200, 0);
		int v22 = nox_xxx_cliGetMana_470DD0();
		nox_swprintf(v77, L"%d", v22);
		nox_xxx_drawStringWrap_43FAF0(nox_inventory_font, v77, v10 + 45, v19 + v73, 200, 0);
		v19 += v1 + 1;
	}
	nox_client_drawSetColor_434460(*getMemIntPtr(0x85B3FC, 956));
	nox_client_drawRectFilledOpaque_49CE30(v10 + 60, v19, 90, v1);
	v68 = 90 * *(uint32_t*)(v4 + 2239);
	float v60 = (double)v68 / v70[3];
	v67 = nox_float2int(v60);
	wchar2_t* v23 = nox_strman_loadString_40F1D0("StatsStrength", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1975);
	nox_xxx_drawStringWrap_43FAF0(0, v23, v10, v19, 200, 0);
	nox_client_drawSetColor_434460(*getMemIntPtr(0x5D4594, 2597996));
	nox_client_drawRectFilledOpaque_49CE30(v10 + 60, v19, v67, v1);
	int v61 = nox_float2int(v70[3]);
	int v44 = *(uint32_t*)(v4 + 2239);
	wchar2_t* v24 = nox_strman_loadString_40F1D0("MinMaxFormat", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1982);
	nox_swprintf(v77, v24, v44, v61);
	nox_xxx_drawGetStringSize_43F840(nox_inventory_font, v77, &v67, 0, 0);
	nox_xxx_drawStringWrap_43FAF0(nox_inventory_font, v77, v10 - v67 + 193, v19 + v73, 200, 0);
	nox_swprintf(v77, L"%d", *(uint32_t*)(v4 + 2239));
	nox_xxx_drawStringWrap_43FAF0(nox_inventory_font, v77, v10 + 45, v19 + v73, 200, 0);
	int v25 = v19 + v1 + 1;
	nox_client_drawSetColor_434460(nox_color_orange_2614256);
	nox_client_drawRectFilledOpaque_49CE30(v10 + 60, v25, 90, v1);
	v68 = 90 * *(uint32_t*)(v4 + 2235);
	float v62 = (double)v68 / v70[2] + 0.5;
	v67 = nox_float2int(v62);
	wchar2_t* v26 = nox_strman_loadString_40F1D0("StatsSpeed", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2006);
	nox_xxx_drawStringWrap_43FAF0(0, v26, v10, v25, 200, 0);
	nox_client_drawSetColor_434460(nox_color_yellow_2589772);
	nox_client_drawRectFilledOpaque_49CE30(v10 + 60, v25, v67, v1);
	nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
	v69 = *getMemFloatPtr(0x5D4594, 1063100) / (v71[2] * 0.000001);
	if (getMemByte(0x5D4594, 1062541) & 2) {
		v69 = ((double)v67 + v69) * 1.25 - (double)v67;
	}
	if (getMemByte(0x5D4594, 1062540) & 0x10) {
		v69 = ((double)v67 + v69) * 0.5 - (double)v67;
	}
	if (v69 >= 0.0) {
		if (v69 > 0.0) {
			*(float*)&v68 = COERCE_FLOAT(nox_float2int(v69));
			if (v67 + v68 > 90) {
				v68 = 90 - v67;
			}
			nox_client_drawSetColor_434460(nox_color_yellow_2589772);
			nox_client_drawRectFilledOpaque_49CE30(v67 + v10 + 60, v25, v68, v1);
			nox_xxx_drawSetTextColor_434390(nox_color_blue_2650684);
		}
	} else {
		nox_client_drawSetColor_434460(*getMemIntPtr(0x85B3FC, 944));
		float v45 = -v69;
		int v46 = nox_float2int(v45);
		int v27 = nox_float2int(v69);
		nox_client_drawRectFilledOpaque_49CE30(v67 + v27 + v10 + 60, v25, v46, v1);
	}
	*(float*)&v68 = v69 * 100.0 * 0.011111111;
	float v63 = v70[2] * 100.0 / v71[2];
	int v64 = nox_float2int(v63);
	float v47 = (double)*(int*)(v4 + 2235) * 100.0 / v71[2] + *(float*)&v68 + 0.5;
	int v48 = nox_float2int(v47);
	wchar2_t* v28 = nox_strman_loadString_40F1D0("MinMaxFormat", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2045);
	nox_swprintf(v77, v28, v48, v64);
	int v76 = 0;
	nox_xxx_drawGetStringSize_43F840(nox_inventory_font, v77, &v76, 0, 0);
	LODWORD(v69) = v25 + v73;
	nox_xxx_drawStringWrap_43FAF0(nox_inventory_font, v77, v10 - v76 + 193, v25 + v73, 200, 0);
	nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
	float v65 = (double)*(int*)(v4 + 2235) * 100.0 / v71[2] + *(float*)&v68 + 0.5;
	int v29 = nox_float2int(v65);
	nox_swprintf(v77, L"%d", v29);
	nox_xxx_drawStringWrap_43FAF0(nox_inventory_font, v77, v10 + 45, SLODWORD(v69), 200, 0);
	nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
	int v30 = 2 * v1 + 2 + v25;
	if (nox_strman_get_lang_code() == 6 || nox_strman_get_lang_code() == 8) {
		v10 += 39;
	}
	nox_xxx_drawSetTextColor_434390(v72);
	wchar2_t* v31 = nox_strman_loadString_40F1D0("StatsArmorLabel", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2072);
	nox_wcsncpy(v77, v31, sizeof(v77)/2);
	int v75 = 0;
	nox_xxx_drawGetStringSize_43F840(0, v77, &v75, 0, 0);
	nox_xxx_drawStringWrap_43FAF0(0, v77, v10, v30, 0, 0);
	float v49 = *getMemFloatPtr(0x5D4594, 1062548) * 1000.0 + 0.5;
	int v50 = nox_float2int(v49);
	wchar2_t* v32 = nox_strman_loadString_40F1D0("MinMaxFormat", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2076);
	nox_swprintf(v77, v32, v50, 1000);
	nox_xxx_drawStringWrap_43FAF0(0, v77, v75 + v10 + 5, v30, 0, 0);
	int v74 = v30 + v1 + 1;
	int itemsWeight = 0;
	for (int i = 0; i < NOX_INVENTORY_ROW_COUNT; i++) {
		nox_inventory_cell_t* v71a = &nox_client_inventory_grid_1050020[i];
		unsigned char* v33 = &(v71a->field_140);
		for (int j = 0; j < NOX_INVENTORY_COL_COUNT; j++) {
			if (*v33) {
				itemsWeight += *v33 * *(unsigned char*)(*((uint32_t*)v33 - 35) + 298);
			}
			v33 += NOX_INVENTORY_ROW_COUNT * sizeof(nox_inventory_cell_t);
		}
	}
	nox_xxx_drawSetTextColor_434390(v72);
	wchar2_t* v35 = nox_strman_loadString_40F1D0("DollWeight", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2098);
	nox_xxx_drawGetStringSize_43F840(0, v35, &v67, 0, 0);
	int v36 = v74;
	int v40 = v74;
	int v39 = v10 + v75 - v67;
	wchar2_t* v37 = nox_strman_loadString_40F1D0("DollWeight", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2099);
	nox_xxx_drawStringWrap_43FAF0(0, v37, v39, v40, 0, 0);
	if (itemsWeight > *(unsigned short*)(v4 + 3652)) {
		v72 = *getMemU32Ptr(0x85B3FC, 940);
	}
	nox_xxx_drawSetTextColor_434390(v72);
	int v66 = *(unsigned short*)(v4 + 3652);
	wchar2_t* v38 = nox_strman_loadString_40F1D0("MinMaxFormat", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 2107);
	nox_swprintf(v77, v38, itemsWeight, v66);
	nox_xxx_drawStringWrap_43FAF0(0, v77, v75 + v10 + 5, v36, 0, 0);
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
//----- (00465A30) --------------------------------------------------------
void nox_xxx_cliInventorySpriteUpd_465A30() {
	int inventory_item_idx = dword_5d4594_1049800_inventory_click_row_index +
							 NOX_INVENTORY_ROW_COUNT * dword_5d4594_1049796_inventory_click_column_index;
	if (nox_client_inventory_grid_1050020[inventory_item_idx].field_140) {
		nox_inventory_cell_t* cell = &nox_client_inventory_grid_1050020[inventory_item_idx];
		nox_drawable* source = cell->field_0;
		nox_drawable* dragged = nox_new_drawable_for_thing(source->field_27);
		nox_client_inventory_set_dragged(dragged);
		if (dragged) {
			dragged->flags30 |= 0x40000000u;
			dragged->field_32 = cell->field_4;
			for (size_t i = 0; i < 4; i++) {
				dragged->item_modifiers[i] = source->item_modifiers[i];
			}
			dragged->item_field_112_0 = source->item_field_112_0;
			dragged->item_field_112_2 = source->item_field_112_2;
			dragged->field_113 = source->field_113;
			dragged->field_73_1 = source->field_73_1;
			dragged->field_73_2 = source->field_73_2;
			sub_461E60(cell, 0);
		} else {
			wchar2_t* v2 =
				nox_strman_loadString_40F1D0("DrawablesExhausted", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 1123);
			nox_xxx_printCentered_445490(v2);
		}
	}
}

//----- (00466160) --------------------------------------------------------
int sub_466160() {
	wchar2_t* v0; // eax

	if (getMemByte(0x5D4594, 1049868) == 2) {
		v0 = nox_strman_loadString_40F1D0("CloseInventoryTT", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 410);
	} else {
		v0 = nox_strman_loadString_40F1D0("OpenInventoryTT", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 414);
	}
	nox_xxx_cursorSetTooltip_4776B0(v0);
	return 1;
}

//----- (004661D0) --------------------------------------------------------
int sub_4661D0() {
	wchar2_t* v0; // eax
	wchar2_t* v2; // eax

	if (dword_5d4594_1062480) {
		v0 = nox_xxx_clientAskInfoMb_4BF050(**(wchar2_t***)&dword_5d4594_1062480);
		nox_xxx_cursorSetTooltip_4776B0(v0);
	} else {
		v2 = nox_strman_loadString_40F1D0("ToolTipWeapon2Area", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 3331);
		nox_xxx_cursorSetTooltip_4776B0(v2);
	}
	return 1;
}

//----- (00466660) --------------------------------------------------------
nox_drawable* nox_inventory_prepare_tooltip_drawable_466660(nox_inventory_cell_t* cell) {
	if (!cell || !cell->field_140 || !cell->field_0) {
		return NULL;
	}
	nox_drawable* drawable = cell->field_0;
	drawable->field_32 = cell->field_4;
	return drawable;
}

wchar2_t* sub_466660(int a1, int2* a2) {
	int v2;          // eax
	int v3;          // eax
	wchar2_t* result; // eax
	wchar2_t* v5;     // eax
	int v6;          // eax
	int v7;          // esi
	int v8;          // edi
	int v9;          // edx
	int v10;         // eax
	nox_drawable* v13; // ecx

	v2 = nox_xxx_pointInRect_4281F0(a2, (int4*)getMemAt(0x587000, 136336));
	if (!v2) {
		if (!getMemByte(0x5D4594, 1049869)) {
			v6 = nox_xxx_pointInRect_4281F0(a2, (int4*)getMemAt(0x587000, 136368));
			if (v6) {
				v7 = a2->field_4 - 13;
				v8 = v7 / 50;
				v9 = 20;
				dword_5d4594_1049796_inventory_click_column_index = v7 / 50;
			} else {
				v10 = nox_xxx_pointInRect_4281F0(a2, (int4*)getMemAt(0x587000, 136352));
				if (!v10) {
					return 0;
				}
				v8 = (a2->field_0 - 314) / 50;
				dword_5d4594_1049796_inventory_click_column_index = (a2->field_0 - 314) / 50;
				v9 = (a2->field_4 + dword_5d4594_1062512 - 13) / 50;
			}
			dword_5d4594_1049800_inventory_click_row_index = v9;
			if (sub_464B40(v8, v9)) {
				int v12 = dword_5d4594_1049800_inventory_click_row_index +
						  NOX_INVENTORY_ROW_COUNT * dword_5d4594_1049796_inventory_click_column_index;
				v13 = nox_inventory_prepare_tooltip_drawable_466660(&nox_client_inventory_grid_1050020[v12]);
				if (v13) {
					return nox_xxx_clientAskInfoMb_4BF050((wchar2_t*)v13);
				}
			}
		}
		return 0;
	}
	if (getMemByte(0x5D4594, 1049870) == 1) {
		return 0;
	}
	v3 = sub_465990(a2);
	if (v3 == -1) {
		return nox_strman_loadString_40F1D0("DollRegionError", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 3155);
	}
	v5 = *(wchar2_t**)&array_5D4594_1049872[v3];
	if (v5) {
		result = nox_xxx_clientAskInfoMb_4BF050(v5);
	} else {
		result = nox_strman_loadString_40F1D0("ToolTipDrag", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 3159);
	}
	return result;
}
//----- (004667E0) --------------------------------------------------------
int nox_xxx_inventroryOnHovewerSub_4667E0(int a1, int a2, unsigned int a3) {
	int v3;       // edx
	int v4;       // esi
	int v5;       // ecx
	int v6;       // eax
	wchar2_t* v7;  // eax
	int v9;       // ecx
	wchar2_t* v10; // eax
	int v11;      // eax
	wchar2_t* v12; // eax
	int v13;      // eax
	wchar2_t* v14; // eax
	int2 v15;     // [esp+4h] [ebp-8h]

	v3 = 40;
	v15.field_0 = (unsigned short)a3;
	v15.field_4 = a3 >> 16;
	v4 = 0;
	while (v3 <= (unsigned short)a3) {
		v3 += 35;
		++v4;
	}
	v5 = 0;
	do {
		if ((1 << v5) & *getMemU32Ptr(0x5D4594, 1062540) && v5 != 31) {
			--v4;
		}
		if (v4 < 0) {
			break;
		}
		++v5;
	} while (v5 < 32);
	if (v5 != 32) {
		v6 = nox_xxx_getEnchantSpell_424920(v5);
		v7 = (wchar2_t*)nox_xxx_spellTitle_424930(v6);
		nox_xxx_cursorSetTooltip_4776B0(v7);
		return 1;
	}
	v9 = 0;
	do {
		if ((1 << v9) & getMemByte(0x5D4594, 1062536)) {
			--v4;
		}
		if (v4 < 0) {
			break;
		}
		++v9;
	} while (v9 < 6);
	if (v9 != 6) {
		v10 = sub_413480(1 << v9);
		nox_xxx_cursorSetTooltip_4776B0(v10);
		return 1;
	}
	if (!nox_common_gameFlags_check_40A5C0(4096)) {
		nox_xxx_cursorSetTooltip_4776B0(0);
		return 1;
	}
	v11 = nox_xxx_pointInRect_4281F0(&v15, (int4*)getMemAt(0x5D4594, 1049812));
	if (v11 == 1) {
		v12 = nox_strman_loadString_40F1D0("thing.db:AnkhGUI", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 4385);
		nox_xxx_cursorSetTooltip_4776B0(v12);
		return 1;
	}
	v13 = nox_xxx_pointInRect_4281F0(&v15, (int4*)getMemAt(0x5D4594, 1049828));
	if (v13 == 1 && sub_4BFD30() == 1) {
		v14 = nox_strman_loadString_40F1D0("GeneralPrint:TooltipKeyIcon", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c",
										   4388);
		nox_xxx_cursorSetTooltip_4776B0(v14);
		return 1;
	} else {
		nox_xxx_cursorSetTooltip_4776B0(0);
		return 1;
	}
}
//----- (00466E20) --------------------------------------------------------
int sub_466E20(uint32_t* a1) {
	wchar2_t* v1; // eax

	switch (*a1) {
	case 0x2391:
		v1 = nox_strman_loadString_40F1D0("JournalModeTT", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 424);
		nox_xxx_cursorSetTooltip_4776B0(v1);
		return 1;
	case 0x2392:
		v1 = nox_strman_loadString_40F1D0("InventoryModeTT", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 428);
		nox_xxx_cursorSetTooltip_4776B0(v1);
		return 1;
	case 0x2393:
		v1 = nox_strman_loadString_40F1D0("StatsModeTT", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 432);
		nox_xxx_cursorSetTooltip_4776B0(v1);
		return 1;
	case 0x2394:
		v1 = nox_strman_loadString_40F1D0("PaperDollModeTT", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 436);
		nox_xxx_cursorSetTooltip_4776B0(v1);
		return 1;
	case 0x2397:
		v1 = nox_strman_loadString_40F1D0("CloseInventoryTT", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 440);
		nox_xxx_cursorSetTooltip_4776B0(v1);
		return 1;
	default:
		return 0;
	}
}

//----- (004671E0) --------------------------------------------------------
int nox_xxx_inventoryNameSignInit_4671E0() {
	static const char* const class_names[] = {"Warrior", "Wizard", "Conjurer"};
	char key[100];

	nox_wcscpy((wchar2_t*)getMemAt(0x5D4594, 1062588), (const wchar2_t*)getMemAt(0x5D4594, 1063676));
	nox_playerInfo* player = (nox_playerInfo*)dword_8531A0_2576;
	int level;
	if (nox_common_gameFlags_check_40A5C0(4096) || nox_xxx_isQuest_4D6F50() || sub_4D6F70()) {
		level = dword_5d4594_1049844;
		if (level > NOX_PLAYER_MAX_LEVEL) {
			level = NOX_PLAYER_MAX_LEVEL;
		}
	} else {
		if (!player) {
			return 0;
		}
		level = player->field_3684;
	}
	if (player && player->info.playerClass < sizeof(class_names) / sizeof(class_names[0])) {
		nox_sprintf(key, "experience:%s%d", class_names[player->info.playerClass], level);
		wchar2_t* level_name =
			nox_strman_loadString_40F1D0(key, 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 4763);
		wchar2_t* format = nox_strman_loadString_40F1D0(
			"ElaborateNameFormat", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 4761);
		return nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1062588), format, player->name_final, level_name);
	}
	return level;
}

//----- (00467750) --------------------------------------------------------
int sub_467750(int a1, char a2) {
	if (a1 != 0) {
		uint32_t* v2 = (uint32_t*)sub_461EF0(a1);
		if (v2 != NULL) {
			if (dword_5d4594_1062480) {
				*(uint32_t*)(dword_5d4594_1062480 + 136) = 0;
			}
			dword_5d4594_1062480 = *v2;
			*(uint32_t*)(dword_5d4594_1062480 + 136) = 1;
			return 1;
		}
	} else {
		if (dword_5d4594_1062480) {
			*(uint32_t*)(dword_5d4594_1062480 + 136) = 0;
			dword_5d4594_1062480 = 0;
		}
	}

	if (a2) {
		if (a2 != 1) {
			return 0;
		}
		wchar2_t* v5 =
			nox_strman_loadString_40F1D0("Weapon2CantUse", 0, "C:\\NoxPost\\src\\Client\\Gui\\guiinv.c", 5379);
		nox_xxx_printCentered_445490(v5);
		if (!dword_5d4594_1062484) {
			return 0;
		}
		int* v6 = (int*)sub_461EF0(*(int*)&dword_5d4594_1062484);
		if (v6) {
			nox_xxx_clientSetAltWeapon_461550(*v6);
			return 0;
		}
	}
	dword_5d4594_1062484 = 0;
	return 0;
}
