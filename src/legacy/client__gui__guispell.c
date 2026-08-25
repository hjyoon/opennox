#include "client__gui__guispell.h"
#include "client__gui__window.h"
#include "common__strman.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3_2.h"
#include "client__gui__gamewin__gamewin.h"
#include "client__gui__guimsg.h"
#include "common__magic__speltree.h"
#include "common__random.h"
#include "operators.h"

extern uintptr_t dword_8531A0_2576;
extern nox_window* dword_5d4594_1049516;
extern nox_window* dword_5d4594_1049524;
extern uint32_t dword_5d4594_1049536;
extern uint32_t dword_5d4594_1049484;
extern uint32_t dword_5d4594_1047552;
extern nox_window* dword_5d4594_1049512;
extern uint32_t dword_5d4594_1047548;
extern nox_window* dword_5d4594_1049520;
extern nox_window* dword_5d4594_1049508;
extern nox_window* dword_5d4594_1049500;
extern nox_window* dword_5d4594_1049504;
extern void* nox_xxx_aClosewoodengat_587000_133480;
extern int nox_win_width;
extern int nox_win_height;
extern uint32_t nox_color_white_2523948;

//----- (0045DEB0) --------------------------------------------------------
int nox_xxx_spellPutInBox_45DEB0(int* a1, int a2, int a3, int a4) {
	int v4;       // ebx
	wchar2_t* v5;  // eax
	uint32_t* v6; // ecx
	int v7;       // eax

	v4 = nox_xxx_spellBoxPointToWnd_45DE60(a1, a3, a4);
	if (v4 >= 0) {
		if (a1 != getMemIntPtr(0x5D4594, 1047940)) {
			nox_xxx_spellKeyPackSetSpell_45DC40(a1, a2, v4);
			return 1;
		}
		if (nox_xxx_spellCanUseInTrap_424BF0(a2)) {
			v6 = (uint32_t*)a1[51];
			v7 = 0;
			while (*v6 != a2) {
				++v7;
				v6 += 2;
				if (v7 >= 3) {
				nox_xxx_spellKeyPackSetSpell_45DC40(a1, a2, v4);
				return 1;
				}
			}
			v5 = nox_strman_loadString_40F1D0("OneSpellPerTrap", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 504);
		} else {
			v5 = nox_strman_loadString_40F1D0("RestrictedTrapSpell", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c",
											  496);
		}
		nox_xxx_printCentered_445490(v5);
		nox_xxx_clientPlaySoundSpecial_452D80(925, 100);
	}
	return 0;
}

//----- (0045E040) --------------------------------------------------------
void nox_client_buildTrap_45E040() {
	uint32_t** v0;    // edx
	uint32_t* v1;     // ecx
	int v2;           // edi
	int i;            // eax
	uint32_t* result; // eax
	wchar2_t* v5;      // eax
	int* v6;          // ecx
	int v7;           // esi
	char v8;          // al
	int v9[5];        // [esp+8h] [ebp-14h]

	v0 = *(uint32_t***)getMemAt(0x5D4594, 1048144);
	v1 = *(uint32_t**)getMemAt(0x5D4594, 1048144);
	v2 = 0;
	for (i = 0; i < 3; ++i) {
		if (*v1) {
			break;
		}
		v1 += 2;
	}
	if (i == 3) {
		if (dword_8531A0_2576) {
			if (*(uint32_t*)(dword_8531A0_2576 + 3832)) {
				v5 = nox_strman_loadString_40F1D0("TrapError", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1145);
				nox_xxx_printCentered_445490(v5);
				nox_xxx_clientPlaySoundSpecial_452D80(925, 100);
			}
		}
	} else {
		*getMemU8Ptr(0x5D4594, 1049488) = 0;
		v6 = v9;
		v7 = 3;
		do {
			result = *v0;
			if (*v0) {
				*v6 = (int)result;
				++v2;
				++v6;
			}
			v0 += 2;
			--v7;
		} while (v7);
		if (v2 < 5) {
			v8 = getMemByte(0x5D4594, 1047924);
			v9[v2] = 34;
			nox_xxx_clientSendSpell_45DB20((char*)v9, v2 + 1, v8);
			*getMemU32Ptr(0x5D4594, 1047916) = 0;
			*getMemU32Ptr(0x5D4594, 1049480) = 0;
		}
	}
}
//----- (0045E190) --------------------------------------------------------
int nox_xxx_quickBarCreate_45E190() {
	int v0;             // ebp
	int v1;             // ebx
	int v2;             // edi
	unsigned char* v4;  // esi
	nox_video_bag_image_t* v5;  // eax
	nox_video_bag_image_t* v6;  // eax
	nox_video_bag_image_t* v7;  // eax
	nox_video_bag_image_t* v8;  // eax
	wchar2_t* v9;        // eax
	wchar2_t* v10;       // eax
	nox_video_bag_image_t* v11; // eax
	nox_video_bag_image_t* v12; // eax
	nox_video_bag_image_t* v13; // eax
	nox_video_bag_image_t* v14; // eax
	wchar2_t* v15;       // eax
	wchar2_t* v16;       // eax
	nox_video_bag_image_t* v17; // eax
	nox_video_bag_image_t* v18; // eax
	nox_video_bag_image_t* v19; // eax
	nox_video_bag_image_t* v20; // eax
	nox_video_bag_image_t* v21; // eax
	int v22;            // eax
	nox_window* v23;    // eax
	nox_window* v24;    // esi
	nox_video_bag_image_t* v25; // eax
	wchar2_t* v26;       // eax
	nox_window* v27;    // esi
	nox_video_bag_image_t* v28; // eax
	wchar2_t* v29;       // eax
	nox_video_bag_image_t* v30; // eax
	nox_video_bag_image_t* v31; // eax
	nox_video_bag_image_t* v32; // eax
	nox_video_bag_image_t* v33; // eax
	nox_video_bag_image_t* v34; // eax
	nox_video_bag_image_t* v35; // eax
	wchar2_t* v36;       // eax
	nox_window* v37;    // esi
	nox_video_bag_image_t* v38; // eax
	wchar2_t* v39;       // eax
	nox_window* v40;    // esi
	nox_video_bag_image_t* v41; // eax
	wchar2_t* v42;       // eax
	nox_window* v43;    // esi
	nox_video_bag_image_t* v44; // eax
	wchar2_t* v45;       // eax
	nox_video_bag_image_t* v46; // eax
	nox_window* v47;    // eax
	int v48;            // eax
	int v51;            // edi
	int v52;            // ecx
	nox_window* v53;    // esi
	nox_video_bag_image_t* v54; // eax
	nox_video_bag_image_t* v55; // eax
	nox_video_bag_image_t* v56; // eax
	nox_video_bag_image_t* v57; // eax
	nox_window* v58;    // esi
	nox_window* v59;    // esi
	int v60;            // eax
	bool v61;           // zf
	int j;              // eax
	int v63;            // [esp-8h] [ebp-58h]
	unsigned short v64; // [esp+Ch] [ebp-44h]
	int v65;            // [esp+10h] [ebp-40h]
	int v66;            // [esp+14h] [ebp-3Ch]
	int v67;            // [esp+18h] [ebp-38h]
	int v68;            // [esp+20h] [ebp-30h]
	int v69;            // [esp+20h] [ebp-30h]
	int i;              // [esp+24h] [ebp-2Ch]
	unsigned char* v71; // [esp+28h] [ebp-28h]
	char v72[32];       // [esp+30h] [ebp-20h]

	v0 = 0;
	*getMemU32Ptr(0x5D4594, 1047916) = 0;
	*getMemU8Ptr(0x5D4594, 1047920) = 0;
	sub_416170(5);
	*getMemU32Ptr(0x5D4594, 1047924) = 0;
	*getMemU32Ptr(0x5D4594, 1047928) = 0;
	*getMemU32Ptr(0x5D4594, 1049480) = 0;
	*getMemU8Ptr(0x5D4594, 1049488) = 0;
	v68 = nox_xxx_guiFontHeightMB_43F320(0);
	dword_5d4594_1047548 = (nox_win_width - 320) / 2;
	dword_5d4594_1047552 = nox_win_height - 74;
	*getMemU32Ptr(0x5D4594, 1049684) = nox_xxx_guiFontPtrByName_43F360("small");
	*getMemU32Ptr(0x587000, 133656) = dword_5d4594_1047548;
	v1 = dword_5d4594_1047548 + 69;
	*getMemU32Ptr(0x587000, 133660) = dword_5d4594_1047552 - 17;
	v2 = dword_5d4594_1047552 + 32;
	*getMemU32Ptr(0x587000, 133664) = dword_5d4594_1047548 + 320;
	*getMemU32Ptr(0x587000, 133668) = nox_win_height;
	if (!dword_8531A0_2576) {
		return 0;
	}
	if (*(uint8_t*)(dword_8531A0_2576 + 2251)) {
		nox_xxx_quickBarInitWindow_4601F0(nox_xxx_aClosewoodengat_587000_133480, dword_5d4594_1047548 + 69,
										  dword_5d4594_1047552 + 32, 5, 0, nox_xxx_quickBarWnd_45EF50,
										  nox_xxx_quickBarWarriorDraw_45FDE0);
	} else {
		nox_xxx_quickBarInitWindow_4601F0(nox_xxx_aClosewoodengat_587000_133480, dword_5d4594_1047548 + 69,
										  dword_5d4594_1047552 + 32, 5, 0, nox_xxx_quickBarWnd_45EF50,
										  nox_xxx_quickBarDrawFn_45FBD0);
	}
	v4 = getMemAt(0x5D4594, 1048964);
	do {
		v2 -= 60;
		nox_xxx_quickBarInitWindow_4601F0(v4, v1, v2, 5, 0, nox_xxx_quickBarWnd_45EF50,
										  nox_xxx_quickBarWarriorDraw_45FDE0);
		nox_window_set_hidden(nox_quickbar_root(v4), 1);
		v4[200] = 0;
		if (sizeof(void*) == 4) {
			*((uint32_t*)v4 + 51) = (uint32_t)(uintptr_t)v4;
		}
		v4 -= 256;
	} while (v4 >= (unsigned char*)getMemAt(0x5D4594, 1048196));
	if (*(uint8_t*)(dword_8531A0_2576 + 2251)) {
		dword_5d4594_1049504 =
			nox_window_new(0, 1160, dword_5d4594_1047548 + 260, dword_5d4594_1047552, 45, 66, 0);
		nox_xxx_wndSetOffsetMB_46AE40(dword_5d4594_1049504, -263, 0);
		dword_5d4594_1049536 = nox_win_height - 74;
		nox_xxx_wndSetWindowProc_46B300(dword_5d4594_1049504, nox_xxx_quickbar_45F8D0);
		dword_5d4594_1049520 = nox_window_new(dword_5d4594_1049504, 1032, 9, 33, 32, 32, 0);
		nox_window_set_all_funcs(dword_5d4594_1049520, nox_xxx_quickbarTrapButtonProc_45F7A0,
								 nox_xxx_quickbarDrawFn_460000, 0);
		dword_5d4594_1049500 = nox_window_new(dword_5d4594_1049504, 1160, 0, 19, 12, 12, 0);
		nox_xxx_wndSetWindowProc_46B300(dword_5d4594_1049500, nox_xxx_quickbarTrapProc_45FB90);
		nox_xxx_wndSetOffsetMB_46AE40(dword_5d4594_1049500, -265, -23);
		v5 = nox_xxx_gLoadImg_42F970("QuickBarTrapButton");
		nox_xxx_wndSetIcon_46AE60(dword_5d4594_1049500, v5);
		v6 = nox_xxx_gLoadImg_42F970("QuickBarTrapButtonLit");
		nox_xxx_wndSetIconLit_46AEA0(dword_5d4594_1049500, v6);
	} else {
		dword_5d4594_1049504 =
			nox_window_new(0, 1672, dword_5d4594_1047548 + 260, dword_5d4594_1047552, 45, 66, 0);
		nox_xxx_wndSetOffsetMB_46AE40(dword_5d4594_1049504, -263, 0);
		dword_5d4594_1049536 = nox_win_height - 74;
		nox_xxx_wndSetWindowProc_46B300(dword_5d4594_1049504, nox_xxx_quickbar_45F8D0);
	}
	if (*(uint8_t*)(dword_8531A0_2576 + 2251)) {
		if (*(uint8_t*)(dword_8531A0_2576 + 2251) == 1) {
			if ((!dword_8531A0_2576 || !*(uint32_t*)(dword_8531A0_2576 + 3832)) &&
				(!nox_common_gameFlags_check_40A5C0(0x2000) || nox_common_gameFlags_check_40A5C0(4096) ||
				 nox_xxx_isQuest_4D6F50() || sub_4D6F70())) {
				v17 = nox_xxx_gLoadImg_42F970("QuickBarWarriorRight");
				nox_xxx_wndSetIcon_46AE60(dword_5d4594_1049504, v17);
				v18 = nox_xxx_gLoadImg_42F970("QuickBarWarriorRight");
				nox_xxx_wndSetIconLit_46AEA0(dword_5d4594_1049504, v18);
				nox_xxx_wndClearFlag_46AD80(dword_5d4594_1049520, 8);
				nox_xxx_wndClearFlag_46AD80(dword_5d4594_1049500, 8);
				nox_xxx_wndSetDrawFn_46B340(dword_5d4594_1049500, nox_xxx_quickbarButtonBookDraw_45EF30);
			} else {
				v13 = nox_xxx_gLoadImg_42F970("QuickBarTrap");
				nox_xxx_wndSetIcon_46AE60(dword_5d4594_1049504, v13);
				v14 = nox_xxx_gLoadImg_42F970("QuickBarTrapHit");
				nox_xxx_wndSetIconLit_46AEA0(dword_5d4594_1049504, v14);
				v15 = nox_strman_loadString_40F1D0("ToolTipLayTrap", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c",
												   1805);
				nox_xxx_wndWddSetTooltip_46B000(&dword_5d4594_1049520->draw_data, v15);
				nox_xxx_wnd_46AD60(dword_5d4594_1049520, 8);
				v16 = nox_strman_loadString_40F1D0("ToolTipTrapConstruct", 0,
												   "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1808);
				nox_xxx_wndWddSetTooltip_46B000(&dword_5d4594_1049500->draw_data, v16);
				nox_xxx_wnd_46AD60(dword_5d4594_1049500, 8);
			}
		} else if (*(uint8_t*)(dword_8531A0_2576 + 2251) == 2) {
			if ((!dword_8531A0_2576 || !*(uint32_t*)(dword_8531A0_2576 + 3832)) &&
				(!nox_common_gameFlags_check_40A5C0(0x2000) || nox_common_gameFlags_check_40A5C0(4096) ||
				 nox_xxx_isQuest_4D6F50() || sub_4D6F70())) {
				v11 = nox_xxx_gLoadImg_42F970("QuickBarWarriorRight");
				nox_xxx_wndSetIcon_46AE60(dword_5d4594_1049504, v11);
				v12 = nox_xxx_gLoadImg_42F970("QuickBarWarriorRight");
				nox_xxx_wndSetIconLit_46AEA0(dword_5d4594_1049504, v12);
				nox_xxx_wndClearFlag_46AD80(dword_5d4594_1049520, 8);
				nox_xxx_wndClearFlag_46AD80(dword_5d4594_1049500, 8);
				nox_xxx_wndSetDrawFn_46B340(dword_5d4594_1049500, nox_xxx_quickbarButtonBookDraw_45EF30);
			} else {
				v7 = nox_xxx_gLoadImg_42F970("QuickBarBomber");
				nox_xxx_wndSetIcon_46AE60(dword_5d4594_1049504, v7);
				v8 = nox_xxx_gLoadImg_42F970("QuickBarBomberHit");
				nox_xxx_wndSetIconLit_46AEA0(dword_5d4594_1049504, v8);
				v9 = nox_strman_loadString_40F1D0("ToolTipSummonBomber", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c",
												  1838);
				nox_xxx_wndWddSetTooltip_46B000(&dword_5d4594_1049520->draw_data, v9);
				nox_xxx_wnd_46AD60(dword_5d4594_1049520, 8);
				v10 = nox_strman_loadString_40F1D0("ToolTipTrapConstruct", 0,
												   "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1841);
				nox_xxx_wndWddSetTooltip_46B000(&dword_5d4594_1049500->draw_data, v10);
				nox_xxx_wnd_46AD60(dword_5d4594_1049500, 8);
			}
		}
	} else {
		v19 = nox_xxx_gLoadImg_42F970("QuickBarWarriorRight");
		nox_xxx_wndSetIcon_46AE60(dword_5d4594_1049504, v19);
		v20 = nox_xxx_gLoadImg_42F970("QuickBarWarriorRight");
		nox_xxx_wndSetIconLit_46AEA0(dword_5d4594_1049504, v20);
	}
	if (*(uint8_t*)(dword_8531A0_2576 + 2251)) {
		void* trap_data = getMemAt(0x5D4594, 1047940);
		nox_xxx_quickBarInitWindow_4601F0(trap_data, dword_5d4594_1047548 + 122,
										  dword_5d4594_1047552 - 17, 3, 21, nox_xxx_quickBarWnd_45EF50,
										  nox_xxx_quickBarWarriorDraw_45FDE0);
		nox_window* trap_root = nox_quickbar_root(trap_data);
		v21 = nox_xxx_gLoadImg_42F970("QuickBarTrapTray");
		nox_xxx_wndSetIcon_46AE60(trap_root, v21);
		nox_xxx_wndSetOffsetMB_46AE40(trap_root, -40, -20);
		v22 = *getMemU32Ptr(0x5D4594, 1048192);
		LOBYTE(v22) = getMemByte(0x5D4594, 1048192) | 1;
		*getMemU32Ptr(0x5D4594, 1048192) = v22;
		nox_window_set_hidden(trap_root, 1);
		dword_5d4594_1049484 = 0;
		v23 = nox_window_new(trap_root, 1032, 20, -7, 110, v68, 0);
		nox_window_set_all_funcs(v23, 0, nox_xxx_quickbarDraw_45FAC0, 0);
		v24 = nox_window_new(trap_root, 1032, 15, 12, 10, 14, 0);
		nox_xxx_wndSetIcon_46AE60(v24, 0);
		v25 = nox_xxx_gLoadImg_42F970("QuickBarTrapTrayUpLit");
		nox_xxx_wndSetIconLit_46AEA0(v24, v25);
		nox_window_set_all_funcs(v24, nox_xxx_quickbarTrapUpDownProc_45F630, nox_xxx_quickbarTrapUpDownDraw_45F6F0, 0);
		nox_xxx_wndSetOffsetMB_46AE40(v24, -55, -32);
		v26 = nox_strman_loadString_40F1D0("ToolTipPrevTrap", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1883);
		nox_xxx_wndWddSetTooltip_46B000(&v24->draw_data, v26);
		v24->field_92 = 3;
		v27 = nox_window_new(trap_root, 1032, 15, 32, 10, 14, 0);
		nox_xxx_wndSetIcon_46AE60(v27, 0);
		v28 = nox_xxx_gLoadImg_42F970("QuickBarTrapTrayDownLit");
		nox_xxx_wndSetIconLit_46AEA0(v27, v28);
		nox_window_set_all_funcs(v27, nox_xxx_quickbarTrapUpDownProc_45F630, nox_xxx_quickbarTrapUpDownDraw_45F6F0, 0);
		nox_xxx_wndSetOffsetMB_46AE40(v27, -55, -52);
		v29 = nox_strman_loadString_40F1D0("ToolTipNextTrap", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1892);
		nox_xxx_wndWddSetTooltip_46B000(&v27->draw_data, v29);
		v27->field_92 = 4;
		dword_5d4594_1049508 = nox_window_new(0, 1032, dword_5d4594_1047548 - 1, dword_5d4594_1047552 + 26, 61, 48, 0);
		nox_xxx_wndSetOffsetMB_46AE40(dword_5d4594_1049508, 1, -26);
		v30 = nox_xxx_gLoadImg_42F970("QuickBarSpellSetBase");
		nox_xxx_wndSetIcon_46AE60(dword_5d4594_1049508, v30);
		v31 = nox_xxx_gLoadImg_42F970("QuickBarSpellSetBase");
		nox_xxx_wndSetIconLit_46AEA0(dword_5d4594_1049508, v31);
		dword_5d4594_1049508->field_92 = 5;
		nox_window_set_all_funcs(dword_5d4594_1049508, nox_xxx_quickbar_45F8D0,
								 nox_xxx_quickbarTrapUpDownDraw_45F6F0, 0);
	} else {
		dword_5d4594_1049508 = nox_window_new(0, 1032, dword_5d4594_1047548 - 1, dword_5d4594_1047552 + 26, 61, 48, 0);
		nox_xxx_wndSetOffsetMB_46AE40(dword_5d4594_1049508, 1, -26);
		v32 = nox_xxx_gLoadImg_42F970("QuickBarWarriorLeft");
		nox_xxx_wndSetIcon_46AE60(dword_5d4594_1049508, v32);
		v33 = nox_xxx_gLoadImg_42F970("QuickBarWarriorLeft");
		nox_xxx_wndSetIconLit_46AEA0(dword_5d4594_1049508, v33);
		nox_window_set_all_funcs(dword_5d4594_1049508, nox_xxx_quickbar_45F8D0,
								 nox_xxx_quickbarTrapUpDownDraw_45F6F0, 0);
	}
	dword_5d4594_1049524 = nox_window_new(dword_5d4594_1049508, 1160, 0, 9, 29, 30, 0);
	v34 = nox_xxx_gLoadImg_42F970("SpellbookButton");
	nox_xxx_wndSetIcon_46AE60(dword_5d4594_1049524, v34);
	v35 = nox_xxx_gLoadImg_42F970("SpellbookButtonLit");
	nox_xxx_wndSetIconLit_46AEA0(dword_5d4594_1049524, v35);
	nox_window* spellbook_button = nox_window_new(dword_5d4594_1049524, 1064, 1, 2, 28, 28, 0);
	nox_window_set_all_funcs(spellbook_button, nox_xxx_quickbarButtonBookWnd_45F450,
							 nox_xxx_quickbarButtonBookDraw_45EF30, nox_xxx_quickbarButtonBook_45F3F0);
	v36 = nox_strman_loadString_40F1D0("OpenSpellBookTT", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1931);
	nox_xxx_wndWddSetTooltip_46B000(&spellbook_button->draw_data, v36);
	if (*(uint8_t*)(dword_8531A0_2576 + 2251)) {
		v37 = nox_window_new(dword_5d4594_1049508, 1032, 30, 0, 15, 19, 0);
		nox_xxx_wndSetOffsetMB_46AE40(v37, -29, -26);
		nox_xxx_wndSetIcon_46AE60(v37, 0);
		v38 = nox_xxx_gLoadImg_42F970("QuickBarSpellSetUpLit");
		nox_xxx_wndSetIconLit_46AEA0(v37, v38);
		nox_window_set_all_funcs(v37, nox_xxx_quickbarTrapUpDownProc_45F630, nox_xxx_quickbarTrapUpDownDraw_45F6F0, 0);
		v37->field_92 = 0;
		v39 = nox_strman_loadString_40F1D0("ToolTipPrevSpellSet", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1943);
		nox_xxx_wndWddSetTooltip_46B000(&v37->draw_data, v39);
		v40 = nox_window_new(dword_5d4594_1049508, 1032, 30, 29, 15, 19, 0);
		nox_xxx_wndSetOffsetMB_46AE40(v40, -29, -55);
		nox_xxx_wndSetIcon_46AE60(v40, 0);
		v41 = nox_xxx_gLoadImg_42F970("QuickBarSpellSetDownLit");
		nox_xxx_wndSetIconLit_46AEA0(v40, v41);
		nox_window_set_all_funcs(v40, nox_xxx_quickbarTrapUpDownProc_45F630, nox_xxx_quickbarTrapUpDownDraw_45F6F0, 0);
		v40->field_92 = 1;
		v42 = nox_strman_loadString_40F1D0("ToolTipNextSpellSet", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1953);
		nox_xxx_wndWddSetTooltip_46B000(&v40->draw_data, v42);
		v43 = nox_window_new(dword_5d4594_1049508, 1032, 48, 16, 13, 17, 0);
		nox_xxx_wndSetOffsetMB_46AE40(v43, -47, -42);
		nox_xxx_wndSetIcon_46AE60(v43, 0);
		v44 = nox_xxx_gLoadImg_42F970("QuickBarSpellSetMaxLit");
		nox_xxx_wndSetIconLit_46AEA0(v43, v44);
		nox_window_set_all_funcs(v43, nox_xxx_quickbarTrapUpDownProc_45F630, nox_xxx_quickbarTrapUpDownDraw_45F6F0, 0);
		v43->field_92 = 2;
		v45 = nox_strman_loadString_40F1D0("ToolTipAllSpellSets", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1963);
		nox_xxx_wndWddSetTooltip_46B000(&v43->draw_data, v45);
		dword_5d4594_1049516 = nox_window_new(0, 1032, 0, 0, 1, 1, 0);
		nox_window_set_all_funcs(dword_5d4594_1049516, sub_45EF40, sub_45F8F0, 0);
		dword_5d4594_1049512 =
			nox_window_new(0, 1152, dword_5d4594_1047548, dword_5d4594_1047552, 2, 2, 0);
		v46 = nox_xxx_gLoadImg_42F970("QuickBarTitle");
		nox_xxx_wndSetIcon_46AE60(dword_5d4594_1049512, v46);
		v47 = nox_window_new(dword_5d4594_1049512, 8, 115, 6, 101, 14, 0);
		nox_window_set_all_funcs(v47, 0, sub_45F9B0, 0);
	}
	v64 = 0;
	while (1) {
		v65 = 0;
		v66 = 0;
		v69 = 5;
		v48 = v64 << 8;
		v71 = getMemAt(0x5D4594, 1048196 + v48);
		nox_window* row_root = nox_quickbar_root(v71);
		if (!row_root) {
			return 0;
		}
		v51 = row_root->off_x + 10;
		v52 = row_root->off_y + 5;
		v67 = v51;
		for (i = v52;; v52 = i) {
			int slot = 5 - v69;
			v53 = nox_window_new(0, 1160, v51, v52, 30, 10, 0);
			v63 = v0 + 1;
			if (*(uint8_t*)(dword_8531A0_2576 + 2251)) {
				nox_sprintf(v72, "QuickBarNugget%d", v63);
				v54 = nox_xxx_gLoadImg_42F970(v72);
				nox_xxx_wndSetIcon_46AE60(v53, v54);
				*(uint32_t*)&v72[strlen(v72)] = *getMemU32Ptr(0x587000, 134996);
				v55 = nox_xxx_gLoadImg_42F970(v72);
				nox_xxx_wndSetIconLit_46AEA0(v53, v55);
				v51 = v67;
			} else {
				nox_sprintf(v72, "QuickBarWarriorNugget%d", v63);
				v56 = nox_xxx_gLoadImg_42F970(v72);
				nox_xxx_wndSetIcon_46AE60(v53, v56);
				v57 = nox_xxx_gLoadImg_42F970(v72);
				nox_xxx_wndSetIconLit_46AEA0(v53, v57);
			}
			nox_xxx_wndSetOffsetMB_46AE40(v53, -70 - v65, -23);
			nox_window_set_all_funcs(v53, nox_xxx_quickbar_45F8D0, 0, 0);
			nox_quickbar_set_nugget(v71, slot, v53);
			nox_xxx_updateSpellIconDir_45DE10(v0, v71);
			if (*(uint8_t*)(dword_8531A0_2576 + 2251)) {
				v58 = nox_window_new(v53, 1032, 12, 0, 10, 10, 0);
				nox_window_set_all_funcs(v58, sub_45F520, nox_xxx_quickbarButtonBookDraw_45EF30, sub_45F480);
				v58->field_92 = v0 | (v64 << 16);
				nox_quickbar_set_nugget_child(v71, slot, v58);
			}
			if (v64 == 4) {
				v59 = nox_window_new(v53, 1032, 13, 39, 10, 10, 0);
				nox_window_set_all_funcs(v59, 0, sub_45F5D0, 0);
				v59->field_92 = v0++;
			} else {
				nox_window_set_hidden(v53, 1);
			}
			v60 = *getMemU32Ptr(0x587000, 133488 + v66);
			v51 += v60;
			v67 = v51;
			v61 = v69 == 1;
			v65 += v60;
			v66 += 4;
			--v69;
			if (v61) {
				break;
			}
		}
		if (++v64 >= 5u) {
			break;
		}
		v0 = 0;
	}
	if (!*(uint8_t*)(dword_8531A0_2576 + 2251)) {
		for (j = 0; j < 120; j += 24) {
			*getMemU32Ptr(0x5D4594, 1047764 + 24*1 + j) = *getMemU32Ptr(0x587000, 133536 + j);
			*getMemU32Ptr(0x5D4594, 1047764 + 24*1 + 4 + j) = *getMemU32Ptr(0x587000, 133536 + 4 + j);
			*getMemU32Ptr(0x5D4594, 1047764 + 24*1 + 8 + j) = *getMemU32Ptr(0x587000, 133536 + 8 + j);
			*getMemU32Ptr(0x5D4594, 1047764 + 24*1 + 12 + j) = *getMemU32Ptr(0x587000, 133536 + 12 + j);
			*getMemU32Ptr(0x5D4594, 1047764 + 24*1 + 16 + j) = *getMemU32Ptr(0x587000, 133536 + 16 + j);
			*getMemU32Ptr(0x5D4594, 1047764 + 24*1 + 20 + j) = 0;
		}
	}
	nox_xxx_clientUpdateButtonRow_45E110(0);
	return 1;
}

//----- (0045F3F0) --------------------------------------------------------
int nox_xxx_quickbarButtonBook_45F3F0() {
	wchar2_t* v0; // eax

	if (sub_45CFC0()) {
		v0 = nox_strman_loadString_40F1D0("CloseSpellbookTT", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 901);
	} else {
		v0 = nox_strman_loadString_40F1D0("OpenSpellbookTT", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 905);
	}
	nox_xxx_cursorSetTooltip_4776B0(v0);
	return 1;
}

//----- (0045F480) --------------------------------------------------------
int sub_45F480(int a1) {
	wchar2_t* v1; // eax

	if (sub_45F500(*(unsigned short*)(a1 + 368),
				   (int)getMemAt(0x5D4594, 1048196 + 256 * ((unsigned short)(*(uint32_t*)(a1 + 368) >> 16))))) {
		v1 = nox_strman_loadString_40F1D0("ToolTipCastOnMe", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 941);
	} else {
		v1 = nox_strman_loadString_40F1D0("ToolTipCastAtOther", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 945);
	}
	nox_xxx_cursorSetTooltip_4776B0(v1);
	return 1;
}

//----- (0045F9B0) --------------------------------------------------------
int sub_45F9B0(uint32_t* a1) {
	uint32_t* v1;    // esi
	wchar2_t* v2;     // eax
	char* v3;        // esi
	int v5;          // [esp-8h] [ebp-5Ch]
	int v6;          // [esp+0h] [ebp-54h]
	int v7;          // [esp+4h] [ebp-50h]
	int v8;          // [esp+8h] [ebp-4Ch]
	int v9;          // [esp+Ch] [ebp-48h]
	int v10;         // [esp+10h] [ebp-44h]
	wchar2_t v11[32]; // [esp+14h] [ebp-40h]

	if (!*getMemU32Ptr(0x5D4594, 1049476)) {
		v1 = a1;
		nox_client_wndGetPosition_46AA60(a1, &v6, &a1);
		nox_window_get_size((int)v1, &v7, &v9);
		v5 = *(unsigned char*)((uint32_t)nox_xxx_aClosewoodengat_587000_133480 + 200) + 1;
		v2 = nox_strman_loadString_40F1D0("SpellSet", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1276);
		nox_swprintf(v11, v2, v5);
		nox_xxx_drawGetStringSize_43F840(0, v11, &v8, &v10, 0);
		v3 = (char*)a1 - nox_xxx_guiFontHeightMB_43F320(0);
		a1 = &v3[nox_xxx_guiFontHeightMB_43F320(*getMemIntPtr(0x5D4594, 1049684)) + 1];
		v6 += (v7 - v8) / 2;
		nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
		nox_xxx_drawSetColor_4343E0(*getMemIntPtr(0x852978, 4));
		nox_draw_drawStringHL_43F730(0, (short*)v11, v6, (int)a1);
	}
	return 1;
}

//----- (0045FAC0) --------------------------------------------------------
int nox_xxx_quickbarDraw_45FAC0(uint32_t* a1) {
	uint32_t* v1;   // esi
	wchar2_t* v2;    // eax
	int v4;         // [esp-4h] [ebp-58h]
	int v5;         // [esp+4h] [ebp-50h]
	int v6;         // [esp+8h] [ebp-4Ch]
	int v7;         // [esp+Ch] [ebp-48h]
	int v8;         // [esp+10h] [ebp-44h]
	wchar2_t v9[32]; // [esp+14h] [ebp-40h]

	v1 = a1;
	nox_client_wndGetPosition_46AA60(a1, &a1, &v7);
	nox_window_get_size((int)v1, &v5, &v8);
	v4 = getMemByte(0x5D4594, 1048140) + 1;
	v2 = nox_strman_loadString_40F1D0("TrapSet", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1299);
	nox_swprintf(v9, v2, v4);
	nox_xxx_drawGetStringSize_43F840(0, v9, &v6, 0, 0);
	a1 = (uint32_t*)((char*)a1 + (v5 - v6) / 2);
	nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
	nox_xxx_drawSetColor_4343E0(*getMemIntPtr(0x852978, 4));
	nox_draw_drawStringHL_43F730(0, (short*)v9, (int)a1, v7);
	return 1;
}

//----- (0045FBD0) --------------------------------------------------------
int nox_xxx_quickBarDrawFn_45FBD0(nox_window* win, nox_window_data* draw_data) {
	if (!win) {
		return 0;
	}
	void* quickbar = nox_quickbar_data_for_window(win);
	if (!quickbar) {
		return 0;
	}
	int slot = -1;
	for (int i = 0; i < 5; ++i) {
		if (nox_quickbar_button(quickbar, i) == win) {
			slot = i;
			break;
		}
	}
	unsigned int row_index = *((unsigned char*)quickbar + 200);
	if (slot < 0 || row_index >= 5) {
		return 0;
	}
	uint32_t* ability_slot = (uint32_t*)((unsigned char*)quickbar + 40 * row_index + 8 * slot);
	uint32_t ability = *ability_slot;
	if (!ability) {
		return 1;
	}

	unsigned int x_left;
	unsigned int y_top;
	nox_client_wndGetPosition_46AA60(win, &x_left, &y_top);
	int highlighted = gameFrame() > 0xAu && slot == *getMemU32Ptr(0x587000, 133484) &&
		(unsigned int)(gameFrame() - *getMemU32Ptr(0x5D4594, 1049540)) < 0xA;
	nox_video_bag_image_t* icon = nox_xxx_spellGetAbilityIcon_425310(ability, highlighted);
	wchar2_t* title = nox_xxx_abilityGetName_0_425260(ability);
	nox_xxx_wndWddSetTooltip_46B000(draw_data ? draw_data : &win->draw_data, title);
	if (icon) {
		nox_client_drawImageAt_47D2C0(icon, x_left, y_top);
	} else {
		nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
		wchar2_t* no_icon = nox_strman_loadString_40F1D0(
			"NoIcon", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1388);
		nox_xxx_drawString_43F6E0(0, (short*)no_icon, x_left + 6,
								 y_top + nox_xxx_guiFontHeightMB_43F320(0) + 2);
	}

	int state_offset = 24 * ability;
	if (*getMemU32Ptr(0x5D4594, 1047764 + 12 + state_offset) ||
		!*getMemU32Ptr(0x5D4594, 1047764 + 8 + state_offset) || nox_xxx_playerAnimCheck_4372B0()) {
		unsigned int cooldown = nox_xxx_abilityCooldown_4252D0(ability);
		unsigned int frames_per_step = cooldown / gameFPS();
		if (frames_per_step && !highlighted) {
			nox_client_drawRectFilledAlpha_49CF10(
				x_left, y_top, 34,
				34 - (gameFrame() - *getMemU32Ptr(0x5D4594, 1047764 + 20 + state_offset)) / frames_per_step);
		}
	}
	return 1;
}

//----- (0045FDE0) --------------------------------------------------------
int nox_xxx_quickBarWarriorDraw_45FDE0(int yTop) {
	int v1;             // edi
	int v2;             // esi
	unsigned char* v3;  // eax
	int* v4;            // ebx
	uint32_t* v5;       // eax
	int v7;             // eax
	int v8;             // ecx
	int v9;             // ebp
	unsigned char v10;  // al
	wchar2_t* v11;       // eax
	short* v12;         // eax
	int v13;            // ebp
	int v14;            // edi
	int* v15;           // edi
	int v16;            // [esp-Ch] [ebp-24h]
	int v17;            // [esp-8h] [ebp-20h]
	int xLeft;          // [esp+Ch] [ebp-Ch]
	unsigned char* v19; // [esp+10h] [ebp-8h]
	int v20;            // [esp+14h] [ebp-4h]

	v1 = yTop;
	v2 = 0;
	v3 = *(unsigned char**)(yTop + 368);
	v19 = v3;
	v4 = (int*)*((uint32_t*)v3 + 51);
	v5 = v3 + 212;
	do {
		if (yTop == *v5) {
			break;
		}
		++v2;
		++v5;
	} while (v2 < 5);
	if (v2 == 5) {
		return 0;
	}
	nox_client_wndGetPosition_46AA60((uint32_t*)yTop, &xLeft, &yTop);
	if (v4[2 * v2]) {
		if (gameFrame() <= 0xAu || v2 != *getMemU32Ptr(0x587000, 133484) ||
			(unsigned int)(gameFrame() - *getMemU32Ptr(0x5D4594, 1049540)) >= 0xA) {
			v7 = nox_xxx_spellIcon_424A90(v4[2 * v2]);
		} else {
			v7 = nox_xxx_spellIconHighlight_424AB0(v4[2 * v2]);
		}
		v8 = v4[2 * v2];
		v9 = v7;
		v10 = getMemByte(0x5D4594, 1049544 + v8);
		if ((char)v10 > 0) {
			*getMemU8Ptr(0x5D4594, 1049544 + v8) = v10 - 1;
		}
		v11 = (wchar2_t*)nox_xxx_spellTitle_424930(v4[2 * v2]);
		nox_xxx_wndWddSetTooltip_46B000((wchar2_t*)(v1 + 36), v11);
		if (v9) {
			nox_client_drawImageAt_47D2C0(v9, xLeft, yTop);
		} else {
			nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
			v17 = nox_xxx_guiFontHeightMB_43F320(0) + yTop + 2;
			v16 = xLeft + 6;
			v12 = (short*)nox_strman_loadString_40F1D0("NoIcon", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1491);
			nox_xxx_drawString_43F6E0(0, v12, v16, v17);
		}
		v13 = nox_xxx_cliGetMana_470DD0();
		v14 = nox_xxx_spellManaCost_4249A0(v4[2 * v2], 1);
		v20 = v14;
		if (v19 == getMemAt(0x5D4594, 1047940) && v2 > 0) {
			v15 = v4;
			v19 = (unsigned char*)v2;
			do {
				if (*v15) {
					v13 -= nox_xxx_spellManaCost_4249A0(*v15, 2);
				}
				v15 += 2;
				--v19;
			} while (v19);
			v14 = v20;
		}
		if (!nox_xxx_spellIsEnabled_424B70(v4[2 * v2]) || nox_xxx_playerAnimCheck_4372B0() ||
			nox_client_drawable_testBuff_4356C0(*getMemIntPtr(0x852978, 8), 29)) {
			nox_client_drawRectFilledAlpha_49CF10(xLeft, yTop, 30, 30);
			return 1;
		}
		if (v13 < v14 && v14) {
			nox_client_drawRectFilledAlpha_49CF10(xLeft, yTop, 30, 30 * (v14 - v13) / v14);
			return 1;
		}
	} else {
		nox_xxx_wndWddSetTooltip_46B000((wchar2_t*)(v1 + 36), (wchar2_t*)getMemAt(0x5D4594, 1049716));
	}
	return 1;
}

//----- (00460070) --------------------------------------------------------
int sub_460070() {
	int result;  // eax
	char* v1;    // eax
	char* v2;    // eax
	wchar2_t* v3; // eax
	wchar2_t* v4; // eax
	char* v5;    // eax
	char* v6;    // eax
	wchar2_t* v7; // eax
	wchar2_t* v8; // eax

	result = dword_8531A0_2576;
	if (dword_8531A0_2576) {
		if (*(uint8_t*)(dword_8531A0_2576 + 2251) == 1) {
			v5 = nox_xxx_gLoadImg_42F970("QuickBarTrap");
			nox_xxx_wndSetIcon_46AE60(*(int*)&dword_5d4594_1049504, (int)v5);
			v6 = nox_xxx_gLoadImg_42F970("QuickBarTrapHit");
			nox_xxx_wndSetIconLit_46AEA0(*(int*)&dword_5d4594_1049504, (int)v6);
			v7 = nox_strman_loadString_40F1D0("ToolTipLayTrap", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1544);
			nox_xxx_wndWddSetTooltip_46B000((wchar2_t*)(dword_5d4594_1049520 + 36), v7);
			nox_xxx_wnd_46AD60(*(int*)&dword_5d4594_1049520, 8);
			v8 = nox_strman_loadString_40F1D0("ToolTipTrapConstruct", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c",
											  1547);
			nox_xxx_wndWddSetTooltip_46B000((wchar2_t*)(dword_5d4594_1049500 + 36), v8);
			nox_xxx_wnd_46AD60(*(int*)&dword_5d4594_1049500, 8);
			result = nox_xxx_wndSetDrawFn_46B340(*(int*)&dword_5d4594_1049500, 0);
		} else {
			result = *(unsigned char*)(dword_8531A0_2576 + 2251) - 2;
			if (*(uint8_t*)(dword_8531A0_2576 + 2251) == 2) {
				v1 = nox_xxx_gLoadImg_42F970("QuickBarBomber");
				nox_xxx_wndSetIcon_46AE60(*(int*)&dword_5d4594_1049504, (int)v1);
				v2 = nox_xxx_gLoadImg_42F970("QuickBarBomberHit");
				nox_xxx_wndSetIconLit_46AEA0(*(int*)&dword_5d4594_1049504, (int)v2);
				v3 = nox_strman_loadString_40F1D0("ToolTipSummonBomber", 0, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c",
												  1559);
				nox_xxx_wndWddSetTooltip_46B000((wchar2_t*)(dword_5d4594_1049520 + 36), v3);
				nox_xxx_wnd_46AD60(*(int*)&dword_5d4594_1049520, 8);
				v4 = nox_strman_loadString_40F1D0("ToolTipTrapConstruct", 0,
												  "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 1562);
				nox_xxx_wndWddSetTooltip_46B000((wchar2_t*)(dword_5d4594_1049500 + 36), v4);
				nox_xxx_wnd_46AD60(*(int*)&dword_5d4594_1049500, 8);
				result = nox_xxx_wndSetDrawFn_46B340(*(int*)&dword_5d4594_1049500, 0);
			}
		}
	}
	return result;
}

//----- (00460EC0) --------------------------------------------------------
uint32_t* nox_xxx_quickbarAddTrap_460EC0(int a1) {
	uint32_t* result; // eax
	int v2;           // esi
	int v3;           // eax
	int v4;           // [esp-24h] [ebp-24h]
	int v5;           // [esp-20h] [ebp-20h]
	int v6;           // [esp-1Ch] [ebp-1Ch]
	char v7;          // [esp-14h] [ebp-14h]
	char v8;          // [esp-10h] [ebp-10h]

	if (!a1) {
		return (uint32_t*)sub_460070();
	}
	result = *(uint32_t**)&dword_5d4594_1049500;
	if (!(*(uint8_t*)(dword_5d4594_1049500 + 4) & 8)) {
		v2 = 50;
		dword_5d4594_1049536 = nox_win_height + 1;
		do {
			v8 = nox_common_randomIntMinMax_415FF0(4, 6, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 2732);
			v7 = nox_common_randomIntMinMax_415FF0(3, 6, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 2731);
			v6 = nox_common_randomIntMinMax_415FF0(-20, -5, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 2729);
			v5 = nox_common_randomIntMinMax_415FF0(-5, 5, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 2728);
			v4 = nox_common_randomIntMinMax_415FF0(0, 20, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 2727) +
				 *(uint32_t*)(dword_5d4594_1049504 + 20) + 10;
			v3 = nox_common_randomIntMinMax_415FF0(0, 20, "C:\\NoxPost\\src\\Client\\Gui\\guispell.c", 2726);
			result = nox_client_newScreenParticle_431540(0, v3 + *(uint32_t*)(dword_5d4594_1049504 + 16) + 10, v4, v5,
														 v6, 1, v7, v8, 2, 1);
			--v2;
		} while (v2);
	}
	return result;
}
