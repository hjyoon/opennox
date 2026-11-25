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

void* nox_quickbar_font;

static nox_playerInfo* nox_quickbar_player(void) { return (nox_playerInfo*)dword_8531A0_2576; }

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
	nox_quickbar_font = nox_xxx_guiFontPtrByName_43F360("small");
	if (sizeof(void*) == 4) {
		*getMemU32Ptr(0x5D4594, 1049684) = (uint32_t)(uintptr_t)nox_quickbar_font;
	}
	*getMemU32Ptr(0x587000, 133656) = dword_5d4594_1047548;
	v1 = dword_5d4594_1047548 + 69;
	*getMemU32Ptr(0x587000, 133660) = dword_5d4594_1047552 - 17;
	v2 = dword_5d4594_1047552 + 32;
	*getMemU32Ptr(0x587000, 133664) = dword_5d4594_1047548 + 320;
	*getMemU32Ptr(0x587000, 133668) = nox_win_height;
	nox_playerInfo* player = nox_quickbar_player();
	if (!player) {
		return 0;
	}
	if (player->info.playerClass) {
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
	if (player->info.playerClass) {
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
	if (player->info.playerClass) {
		if (player->info.playerClass == 1) {
			if (!player->spell_lvl[34] &&
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
		} else if (player->info.playerClass == 2) {
			if (!player->spell_lvl[34] &&
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
	if (player->info.playerClass) {
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
	if (player->info.playerClass) {
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
			if (player->info.playerClass) {
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
			if (player->info.playerClass) {
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
	if (!player->info.playerClass) {
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
