#include "client__gui__guibrief.h"
#include "quest_briefing_draw_44f300.h"
#include "client__gui__window.h"
#include "common__strman.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"

extern uintptr_t dword_8531A0_2576;
extern uint32_t dword_587000_122956;
extern uint32_t nox_xxx_aSpellphoneme_3_587000_123008;
extern uint32_t dword_5d4594_832480;
extern uintptr_t dword_5d4594_832520;
extern uintptr_t dword_5d4594_832500;
extern uintptr_t dword_5d4594_832528;
extern uintptr_t dword_5d4594_832524;
extern uintptr_t dword_5d4594_832512;
extern uintptr_t dword_5d4594_832496;
extern uintptr_t dword_5d4594_832516;
extern uintptr_t dword_5d4594_832508;
extern uintptr_t dword_5d4594_832504;
extern uintptr_t dword_5d4594_832492;
extern uintptr_t dword_5d4594_832532;
extern uintptr_t dword_5d4594_832536;
extern nox_window* nox_wnd_briefing_831232;
extern uint32_t dword_5d4594_832476;
extern uintptr_t dword_5d4594_832484;
extern int nox_win_width;
extern int nox_win_height;

extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_blue_2650684;
extern uint32_t nox_color_green_2614268;
extern uint32_t nox_color_black_2650656;
extern uint32_t nox_color_orange_2614256;

//----- (0044E410) --------------------------------------------------------
wchar2_t* sub_44E410() {
	int v0;          // esi
	int v1;          // edi
	int i;           // ebp
	int v3;          // ebx
	int v4;          // esi
	wchar2_t* result; // eax
	char* dialog;
	int v7;          // [esp+10h] [ebp-44h]
	char v8[64];     // [esp+14h] [ebp-40h]

	v0 = 0;
	v7 = 0;
	v1 = 0;
	for (i = 1;; ++i) {
		while (1) {
			if (v0) {
				if (v0 == 1) {
					v3 = v1 ? (v1 != 1) + 2 : 4;
				} else {
					v3 = v0 + 3;
				}
			} else {
				v3 = 1;
			}
			const char* class_name = getMemPtr(0x587000, 122944 + 4 * v1);
			nox_sprintf(v8, "Briefing:%sChapterBegin%d", class_name, i);
			v4 = 32 * (v1 + v0 + 10 * v1);
			setMemPtr(0x5D4594, 831300 + v4, nox_xxx_gLoadImg_42F970(&v8[9]));
			setMemPtr(0x5D4594, 831304 + v4,
				nox_strman_loadString_40F1D0(v8, &dialog, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 1221));
			setMemPtr(0x5D4594, 831308 + v4, dialog);
			*getMemU32Ptr(0x5D4594, 831312 + v4) = v3;
			nox_sprintf(v8, "Briefing:%sChapterLoss%d", class_name, i);
			setMemPtr(0x5D4594, 831316 + v4, nox_xxx_gLoadImg_42F970(&v8[9]));
			++v1;
			setMemPtr(0x5D4594, 831320 + v4,
				nox_strman_loadString_40F1D0(v8, &dialog, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 1227));
			setMemPtr(0x5D4594, 831324 + v4, dialog);
			*getMemU32Ptr(0x5D4594, 831328 + v4) = v3;
			if (v1 >= 3) {
				break;
			}
			v0 = v7;
		}
		v7 = i;
		if (i >= 11) {
			break;
		}
		v0 = i;
		v1 = 0;
	}
	setMemPtr(0x5D4594, 831264, nox_xxx_gLoadImg_42F970("CreditsImage"));
	result = nox_strman_loadString_40F1D0("Nox:Credits", &dialog, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c",
										  1233);
	setMemPtr(0x5D4594, 831268, result);
	setMemPtr(0x5D4594, 831272, dialog);
	return result;
}

//----- (0044E8E0) --------------------------------------------------------
int sub_44E8E0(nox_window* win, nox_window_data* draw) {
	int v2;                  // ebx
	int v3;                  // esi
	wchar2_t* v4;             // eax
	wchar2_t* v5;             // eax
	unsigned char* v6;       // eax
	nox_quest_stats_row_450770* v7; // ebp, native score table
	int v8;                  // esi
	nox_playerInfo* v9;      // ecx, native player identity
	int v10;                 // ebx
	void* v11;               // eax
	signed int v12;          // eax
	int v13;                 // esi
	wchar2_t* v14;            // eax
	int v15;                 // esi
	wchar2_t* v16;            // eax
	int v17;                 // esi
	wchar2_t* v18;            // eax
	int v19;                 // esi
	wchar2_t* v20;            // eax
	wchar2_t* v21;            // eax
	int v22;                 // ebx
	int v23;                 // esi
	int v24;                 // ebp
	wchar2_t* v25;            // eax
	wchar2_t* v26;            // eax
	int v27;                 // esi
	wchar2_t* v28;            // eax
	wchar2_t* v29;            // eax
	int v30;                 // esi
	int v31;                 // ebp
	int result;              // eax
	int v33;                 // ebp
	unsigned short* v34;     // esi
	int v35;                 // ebx
	float v36;               // [esp+0h] [ebp-674h]
	int v37;                 // [esp+0h] [ebp-674h]
	int v38;                 // [esp+14h] [ebp-660h]
	int v39;                 // [esp+18h] [ebp-65Ch]
	int v40;                 // [esp+1Ch] [ebp-658h]
	int v41;                 // [esp+20h] [ebp-654h]
	int v42;                 // [esp+24h] [ebp-650h]
	int v43;                 // [esp+28h] [ebp-64Ch]
	int v44;                 // [esp+2Ch] [ebp-648h]
	int v45;                 // [esp+30h] [ebp-644h]
	int v46;                 // [esp+34h] [ebp-640h]
	int v47;                 // [esp+38h] [ebp-63Ch]
	unsigned char* v48;      // [esp+3Ch] [ebp-638h]
	int v49;                 // [esp+40h] [ebp-634h]
	int v50;                 // [esp+44h] [ebp-630h]
	int v51;                 // [esp+48h] [ebp-62Ch]
	int v52;                 // [esp+4Ch] [ebp-628h]
	int v53;                 // [esp+50h] [ebp-624h]
	wchar2_t WideCharStr[11]; // [esp+54h] [ebp-620h]
	wchar2_t v55[257];        // [esp+72h] [ebp-602h]
	wchar2_t v56[256];        // [esp+274h] [ebp-400h]
	wchar2_t v57[256];        // [esp+474h] [ebp-200h]

	v49 = (nox_win_width - NOX_DEFAULT_WIDTH) / 2;
	v47 = 0;
	v45 = 0;
	v50 = (nox_win_height - NOX_DEFAULT_HEIGHT) / 2;
	nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
	v2 = nox_win_width / 2;
	v3 = nox_win_height / 2;
	v51 = nox_win_width / 2;
	v52 = nox_win_height / 2;
	v4 = nox_strman_loadString_40F1D0("GUIBrief.c:GauntletStatTitle", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c",
									  449);
	nox_wcscpy(&v55[1], v4);
	v5 = nox_strman_loadString_40F1D0("Noxworld.c:Stage", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 451);
	nox_wcscpy(v56, v5);
	nox_wcscat(v56, L" XX1 ");
	nox_itow(*getMemIntPtr(0x5D4594, 831228), WideCharStr, 10);
	nox_wcscat(v56, WideCharStr);
	nox_swprintf(v57, L"%s - %s", &v55[1], v56);
	nox_xxx_drawGetStringSize_43F840(draw->font, v57, &v39, &v38, 0);
	nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
	nox_xxx_drawString_43F6E0(draw->font, (short*)v57, v2 - v39 / 2, v38 + v3 - 240);
	v40 = *getMemU32Ptr(0x587000, 122968) - *getMemU32Ptr(0x587000, 122964);
	v36 = (double)v38 * 1.5;
	v41 = nox_float2int(v36);
	v6 = getMemAt(0x587000, 122964);
	v43 = 0;
	v42 = 0;
	v7 = nox_quest_stats_450770;
	v48 = getMemAt(0x587000, 122964);
	while (1) {
		v8 = *(uint32_t*)v6 + v3 - 240;
		v9 = v7->player;
		v10 = *((uint32_t*)v6 - 1) + v2 - 320;
		if (v7->player) {
			++v43;
			if ((uintptr_t)v9 == dword_8531A0_2576) {
				v47 = v7->coop_secrets;
			} else {
				v45 += v7->coop_secrets;
			}
			nox_xxx_drawSetTextColor_434390(nox_color_orange_2614256);
			nox_swprintf(&v55[1], L"%d) %s", v42 + 1, v7->player->name_final);
			v11 = draw->font;
			v46 = *getMemU32Ptr(0x587000, 122968) - *getMemU32Ptr(0x587000, 122960) + v10 - 16;
			nox_xxx_drawGetStringSize_43F840(v11, &v55[1], &v44, &v53, 0);
			while (v10 + v44 >= v46) {
				v12 = nox_wcslen(&v55[1]);
				if (v12 <= 5) {
					break;
				}
				v55[v12] = 0;
				nox_xxx_drawGetStringSize_43F840(draw->font, &v55[1], &v44, &v53, 0);
			}
			nox_xxx_drawStringWrap_43FAF0(draw->font, &v55[1], v10, v8, v40 - 8, v38);
			v13 = v41 + v41 / 2 + v8;
			nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
			v14 = nox_strman_loadString_40F1D0("GUIBrief.c:GeneratorsDestroyed", 0,
											   "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 529);
			nox_xxx_drawStringWrap_43FAF0(draw->font, v14, v10, v13, *(int*)&dword_5d4594_832476, v38);
			nox_swprintf(&v55[1], L" %d", v7->generators);
			nox_xxx_drawSetTextColor_434390(nox_color_green_2614268);
			nox_xxx_drawStringWrap_43FAF0(draw->font, &v55[1], v10 + dword_5d4594_832476, v13,
									  v40 - dword_5d4594_832476 - 8, v38);
			v15 = v41 + v13;
			nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
			v16 = nox_strman_loadString_40F1D0("GUIBrief.c:numSecretsFound", 0,
											   "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 541);
			nox_xxx_drawStringWrap_43FAF0(draw->font, v16, v10, v15, *(int*)&dword_5d4594_832476, v38);
			nox_swprintf(&v55[1], L" %d", v7->secrets);
			nox_xxx_drawSetTextColor_434390(nox_color_green_2614268);
			nox_xxx_drawStringWrap_43FAF0(draw->font, &v55[1], v10 + dword_5d4594_832476, v15,
									  v40 - dword_5d4594_832476 - 8, v38);
			v17 = v41 + v15;
			nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
			v18 = nox_strman_loadString_40F1D0("GUIBrief.c:Kills", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 553);
			nox_xxx_drawStringWrap_43FAF0(draw->font, v18, v10, v17, *(int*)&dword_5d4594_832476, v38);
			nox_swprintf(&v55[1], L" %d", v7->kills);
			nox_xxx_drawSetTextColor_434390(nox_color_green_2614268);
			nox_xxx_drawStringWrap_43FAF0(draw->font, &v55[1], v10 + dword_5d4594_832476, v17,
									  v40 - dword_5d4594_832476 - 8, v38);
			v19 = v41 + v17;
			nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
			v20 = nox_strman_loadString_40F1D0("GUIBrief.c:TotalScore", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c",
											   565);
			nox_xxx_drawStringWrap_43FAF0(draw->font, v20, v10, v19, *(int*)&dword_5d4594_832476, v38);
			nox_swprintf(&v55[1], L" %d", v7->score);
			nox_xxx_drawSetTextColor_434390(nox_color_blue_2650684);
			nox_xxx_drawStringWrap_43FAF0(draw->font, &v55[1], v10 + dword_5d4594_832476, v19,
									  v40 - dword_5d4594_832476 - 8, v38);
			v6 = v48;
		}
		v6 += 8;
		++v7;
		++v42;
		v48 = v6;
		if ((int)v6 >= (int)getMemAt(0x587000, 123012)) {
			break;
		}
		v2 = v51;
		v3 = v52;
	}
	v21 =
		nox_strman_loadString_40F1D0("GeneralPrint:SecretsTotal", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 583);
	nox_swprintf(&v55[1], v21, *getMemU32Ptr(0x5D4594, 832356));
	nox_xxx_drawGetStringSize_43F840(draw->font, &v55[1], &v39, &v38, 0);
	v22 = v49;
	v23 = v49 - v39 / 2 + 320;
	v24 = v50 + 2 * (150 - v38) + 150 - v38;
	nox_xxx_drawSetTextColor_434390(nox_color_orange_2614256);
	nox_xxx_drawString_43F6E0(draw->font, (short*)&v55[1], v23, v24);
	if (v47) {
		v37 = v47;
		v25 = nox_strman_loadString_40F1D0("GeneralPrint:SecretsFound", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c",
										   593);
		nox_swprintf(&v55[1], v25, v37);
	} else {
		v26 = nox_strman_loadString_40F1D0("GeneralPrint:SecretsNoneFound", 0,
										   "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 596);
		nox_swprintf(&v55[1], v26);
	}
	if (v43 <= 1) {
		nox_wcscpy(v57, &v55[1]);
	} else {
		v27 = v45;
		if (v45) {
			v28 = nox_strman_loadString_40F1D0("GeneralPrint:SecretsFoundByFriends", 0,
											   "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 606);
			if (v28) {
				nox_swprintf(v56, v28, v27);
			}
		} else {
			v29 = nox_strman_loadString_40F1D0("GeneralPrint:SecretsNoneFoundByFriends", 0,
											   "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 612);
			nox_wcscpy(v56, v29);
		}
		nox_swprintf(v57, L"%s - %s", &v55[1], v56);
	}
	nox_xxx_drawGetStringSize_43F840(draw->font, v57, &v39, &v38, 0);
	v30 = v22 - v39 / 2 + 320;
	v31 = v50 + 2 * (225 - v38);
	nox_xxx_drawSetTextColor_434390(nox_color_orange_2614256);
	nox_xxx_drawString_43F6E0(draw->font, (short*)v57, v30, v31);
	result = gameFrame() / 0x1Eu;
	if (gameFrame() % 0x1Eu) {
		if (dword_587000_122956 != 1) {
			return result;
		}
	} else {
		result = 1;
		if (dword_587000_122956 == 1) {
			dword_587000_122956 = gameFrame() % 0x1Eu;
			return result;
		}
		dword_587000_122956 = 1;
	}
	v33 = nox_color_white_2523948;
	v34 =
		nox_strman_loadString_40F1D0("GeneralPrint:QuestSplash12", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 642);
	nox_xxx_drawGetStringSize_43F840(draw->font, v34, &v39, 0, 0);
	v35 = v22 - v39 / 2 + 320;
	nox_xxx_drawSetTextColor_434390(v33);
	return nox_xxx_drawString_43F6E0(draw->font, (short*)v34, v35, v50 + 450);
}

//----- (0044F0F0) --------------------------------------------------------
int sub_44F0F0(nox_window* win, nox_window_data* draw) {
	int v2;             // esi
	int v3;             // ebx
	wchar2_t* v4;        // eax
	int result;         // eax
	int v6;             // esi
	int v7;             // ebx
	unsigned short* v8; // ebp
	int v9;             // esi
	int v10;            // [esp+10h] [ebp-40Ch]
	int v11;            // [esp+14h] [ebp-408h]
	int v12;            // [esp+18h] [ebp-404h]
	wchar2_t v13[256];   // [esp+1Ch] [ebp-400h]
	wchar2_t v14[256];   // [esp+21Ch] [ebp-200h]

	// quest title screen text
	v2 = nox_win_width / 2;
	v3 = nox_win_height / 2;
	nox_xxx_drawSetTextColor_434390(nox_color_white_2523948);
	v4 = nox_strman_loadString_40F1D0("Noxworld.c:Stage", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 668);
	nox_wcscpy(v13, v4);
	nox_wcscat(v13, L" %d");
	nox_swprintf(v14, v13, nox_gui_getQuestStage_450B10());
	nox_xxx_drawGetStringSize_43F840(draw->font, v14, &v10, &v11, 0);
	nox_xxx_drawString_43F6E0(draw->font, (short*)v14, v2 - v10 / 2, v3 + 2 * (v11 - 80) + v11 - 80);
	if (getMemPtr(0x5D4594, 832464)) {
		nox_xxx_drawGetStringSize_43F840(draw->font, getMemPtr(0x5D4594, 832464), &v10, &v11, 0);
		nox_xxx_drawString_43F6E0(draw->font, getMemPtr(0x5D4594, 832464), v2 - v10 / 2,
								  v3 + 2 * (80 - v11) + 80 - v11);
	}
	result = gameFrame() / 0x1Eu;
	if (gameFrame() % 0x1Eu) {
		if (nox_xxx_aSpellphoneme_3_587000_123008 != 1) {
			return result;
		}
	} else {
		result = 1;
		if (nox_xxx_aSpellphoneme_3_587000_123008 == 1) {
			nox_xxx_aSpellphoneme_3_587000_123008 = gameFrame() % 0x1Eu;
			return result;
		}
		nox_xxx_aSpellphoneme_3_587000_123008 = 1;
	}
	v12 = nox_color_white_2523948;
	v6 = (nox_win_width - NOX_DEFAULT_WIDTH) / 2;
	v7 = (nox_win_height - NOX_DEFAULT_HEIGHT) / 2;
	v8 =
		nox_strman_loadString_40F1D0("GeneralPrint:QuestSplash12", 0, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 714);
	nox_xxx_drawGetStringSize_43F840(draw->font, v8, &v10, 0, 0);
	v9 = v6 - v10 / 2 + 320;
	nox_xxx_drawSetTextColor_434390(v12);
	return nox_xxx_drawString_43F6E0(draw->font, (short*)v8, v9, v7 + 462);
}

//----- (0044F300) --------------------------------------------------------
int sub_44F300(nox_window* win, nox_window_data* draw) {
	return nox_client_questBriefingDraw_native_44F300(win, draw);
}

//----- (00450770) --------------------------------------------------------
extern int nox_client_questWinScreen_native_450770(unsigned char* packet);
int nox_xxx_clientQuestWinScreen_450770(const unsigned char* packet) {
	return nox_client_questWinScreen_native_450770((unsigned char*)packet);
}

//----- (00450980) --------------------------------------------------------
int nox_client_showQuestBriefing2_450980(const unsigned char* packet, int a2) {
	int a1 = (int)(uintptr_t)packet;
	char* v2;    // eax
	wchar2_t* v3; // eax
	int result;  // eax

	dword_5d4594_832480 = 0;
	nox_client_resetScreenParticles_431510();
	nox_xxx_bookHideMB_45ACA0(1);
	sub_446780();
	v2 = (char*)nox_xxx_gLoadImg_42F970((const char*)(a1 + 5));
	sub_450AD0(v2);
	if (strlen((const char*)(a1 + 37))) {
		v3 = nox_strman_loadString_40F1D0((char*)(a1 + 37), 0, "C:\\NoxPost\\src\\client\\Gui\\GUIBrief.c", 1714);
		sub_450AF0(v3);
	} else {
		sub_450AF0((wchar2_t*)getMemAt(0x5D4594, 832544));
	}
	nox_gui_setQuestStage_450B00(*(unsigned short*)(a1 + 2));
	if (*(uint8_t*)(a1 + 4) & 2) {
		dword_5d4594_832480 = 1;
	}
	result = a2;
	if (a2) {
		result = nox_client_lockScreenBriefing_450160(254, 1, 2);
	}
	return result;
}

//----- (00450A30) --------------------------------------------------------
extern int nox_client_showQuestBriefing_native_450A30(unsigned char* packet, int show);
int nox_client_showQuestBriefing_450A30(const unsigned char* packet, int a2) {
	return nox_client_showQuestBriefing_native_450A30((unsigned char*)packet, a2);
}
