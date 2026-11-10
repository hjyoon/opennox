#ifndef GAME2_BOOK_REFERENCE_H

#define GAME2_BOOK_REFERENCE_H

#include "game2_book_native.h"

// Exact pre-port native C bodies, private identifier renaming only.

#define nox_xxx_guiCursor_477600 game2_book_old_cursor

#define nox_netlist_addToMsgListCli_40EBC0 game2_book_old_send

static int game2_book_old_nox_client_localPlayerHasFlags30(uint32_t flags);

static void game2_book_old_nox_book_anim_start(nox_things_imageRef_t* ref);

static void game2_book_old_nox_book_anim_set_end(nox_things_imageRef_t* ref, void* fnc);

static nox_image_ref_anim_t* game2_book_old_nox_book_anim_data(nox_things_imageRef_t* ref);

static int game2_book_old_nox_xxx_guiSpellSortFn_45ABC0(const void* a1, const void* a2);

static int game2_book_old_nox_xxx_bookSetColor_45AC40();

static void game2_book_old_nox_client_toggleSpellbook_45AC70();

static int game2_book_old_nox_xxx_bookHideMB_45ACA0(int a1);

static int game2_book_old_nox_xxx_guiSpellSortList_45ADF0(int a1);

static void game2_book_old_nox_xxx_book_45B010(int a1);

static int game2_book_old_nox_xxx_bookWndProc_45B070(int a1, int a2);

static int game2_book_old_nox_xxx_bookClickSpell_45B1F0();

static int game2_book_old_nox_xxx_bookClickCreature_45B200();

static int game2_book_old_nox_xxx_book_45B210(int a1, int a2);

static int game2_book_old_nox_xxx_bookChildWndProcMB_45B360(uint32_t* a1, unsigned int a2);

static int game2_book_old_nox_xxx_bookListWndProc_45B5F0(nox_window* a1, unsigned int a2, unsigned int a3);

static int game2_book_old_nox_xxx_bookMoveToPage_45B930(int a1);

static int game2_book_old_nox_xxx_book_45BD30(int a1, int a2);

static int game2_book_old_nox_xxx_bookInit_45B9D0();

static int game2_book_old_nox_xxx_bookDrawIconFn_45CB30(uint32_t* a1);

static int game2_book_old_nox_xxx_bookWndFn_45CC10(uint32_t* a1, int a2, unsigned int a3);

static int game2_book_old_sub_45CFC0();

static void game2_book_old_nox_xxx_netSpellRewardCli_45CFE0(int a1, int a2, int a3, int a4);

static int* game2_book_old_nox_xxx_bookSetForward_45D200(int* a1, int a2, int2* a3);

static void game2_book_old_nox_xxx_abilityReward_45D290(int a1, char* a2, int a3);

static int game2_book_old_sub_45D320(int a1);

static int game2_book_old_sub_45D400(int a1);

static char* game2_book_old_nox_xxx_clientQuestDisableAbility_45D4A0(int a1);

static int game2_book_old_sub_45D500(int a1);

static uint32_t* game2_book_old_sub_45D550(uint32_t* a1);

static void game2_book_old_nox_xxx_bookFillAll_45D570(int a1, int a2);

static int game2_book_old_sub_45D7D0(int* a1, int* a2);

static void game2_book_old_sub_45D810();

static int game2_book_old_sub_45D9B0();

static int game2_book_old_nox_xxx_guiSpellTargetClickSet_45D9D0(int a1);

static int game2_book_old_nox_xxx_guiSpell_45DA10(int a1);

static void game2_book_old_nox_client_invokeSpellSlot_45DA50(int a1);

static void game2_book_old_nox_xxx_clientStoreLastButton_45DAD0(int a1);

static int game2_book_old_nox_xxx_clientSendAbil_45DAF0(int a1);

static int game2_book_old_nox_xxx_clientSendSpell_45DB20(char* a1, int a2, char a3);

static int game2_book_old_sub_45DB90();

static void game2_book_old_nox_xxx_guiSpellTargetClickCheckSend_45DBB0();

static unsigned char* game2_book_old_nox_quickbar_selected_row(void* base);

static unsigned char* game2_book_old_nox_quickbar_trap_selected_row(void);

static unsigned char* game2_book_old_nox_quickbar_selected_slot(void* base, int slot);

static int game2_book_old_nox_client_localPlayerHasFlags30(uint32_t flags) {
	nox_drawable* local_player = getMemPtr(0x852978, 8);
	return local_player && (local_player->flags30 & flags) == flags;
}

static void game2_book_old_nox_book_anim_start(nox_things_imageRef_t* ref) {
	nox_image_ref_anim_t* anim = game2_book_old_nox_book_anim_data(ref);
	if (anim) {
		anim->field_3 = nox_xxx_bookGet_430B40_get_mouse_prev_seq();
	}
}

static void game2_book_old_nox_book_anim_set_end(nox_things_imageRef_t* ref, void* fnc) {
	nox_image_ref_anim_t* anim = game2_book_old_nox_book_anim_data(ref);
	if (anim) {
		anim->on_end = fnc;
	}
}

static nox_image_ref_anim_t* game2_book_old_nox_book_anim_data(nox_things_imageRef_t* ref) {
	return ref ? (nox_image_ref_anim_t*)ref->field_24 : NULL;
}

static int game2_book_old_nox_xxx_guiSpellSortFn_45ABC0(const void* a1, const void* a2) {
	int v2;            // edi
	const wchar2_t* v3; // esi
	const wchar2_t* v4; // eax
	int result;        // eax
	int v6;            // [esp-4h] [ebp-Ch]
	nox_playerInfo* player = (nox_playerInfo*)(uintptr_t)dword_5d4594_1047516;

	v2 = *(uint32_t*)a2;
	if (dword_5d4594_1046868 == 1) {
		v3 = (const wchar2_t*)nox_xxx_guiCreatureGetName_427240(*(uint32_t*)a1);
		v4 = (const wchar2_t*)nox_xxx_guiCreatureGetName_427240(v2);
	} else {
		v6 = *(uint32_t*)a1;
		if (player->info.playerClass) {
			v3 = (const wchar2_t*)nox_xxx_spellTitle_424930(v6);
			v4 = (const wchar2_t*)nox_xxx_spellTitle_424930(v2);
		} else {
			v3 = (const wchar2_t*)nox_xxx_abilityGetName_0_425260(v6);
			v4 = (const wchar2_t*)nox_xxx_abilityGetName_0_425260(v2);
		}
	}
	if (v3 && v4) {
		result = _nox_wcsicmp(v3, v4);
	} else {
		result = 0;
	}
	return result;
}

static int game2_book_old_nox_xxx_bookSetColor_45AC40() {
	int result; // eax

	*getMemU32Ptr(0x5D4594, 1046880) = nox_color_rgb_4344A0(15, 15, 15);
	result = nox_color_rgb_4344A0(115, 100, 100);
	*getMemU32Ptr(0x5D4594, 1046884) = result;
	return result;
}

static void game2_book_old_nox_client_toggleSpellbook_45AC70() {
	if (wndIsShown_nox_xxx_wndIsShown_46ACC0(nox_win_unk1)) {
		nox_xxx_bookShowMB_45AD70(0);
	} else {
		game2_book_old_nox_xxx_bookHideMB_45ACA0(0);
	}
}

static int game2_book_old_nox_xxx_bookHideMB_45ACA0(int a1) {
	nox_window* v1; // eax
	nox_window* v2; // ecx

	if (wndIsShown_nox_xxx_wndIsShown_46ACC0(nox_win_unk1) || dword_5d4594_1047520 == 1) {
		return 0;
	}
	dword_5d4594_1046864 = 0;
	set_dword_5d4594_3799468(1);
	dword_5d4594_1046868 = dword_5d4594_1046872 != 0;
	v1 = nox_xxx_wndGetCaptureMain_46AE00();
	v2 = nox_win_unk1;
	if (v1 == nox_win_unk1) {
		nox_xxx_wndClearCaptureMain_46ADE0(nox_win_unk1);
		v2 = nox_win_unk1;
	}
	nox_window_set_hidden(v2, 1);
	nox_xxx_clientPlaySoundSpecial_452D80(787, 100);
	if (a1) {
		nox_xxx_aNox_cfg_0_587000_132132 = 1;
		dword_5d4594_1046936 = 0;
		nox_window_set_hidden(dword_5d4594_1046952, 1);
	}
	if (nox_xxx_bookGetSpellDnDType_477670() == 1) {
		nox_xxx_bookSpellDnDclear_477660();
	}
	return 1;
}

static int game2_book_old_nox_xxx_guiSpellSortList_45ADF0(int a1) {
	int v1; // edi
	int i;  // esi
	int v3; // eax
	int k;  // esi
	int v5; // eax
	int j;  // esi
	int v7; // eax
	int v9; // [esp+Ch] [ebp-4h]
	nox_playerInfo* player = (nox_playerInfo*)(uintptr_t)dword_5d4594_1047516;

	v1 = 0;
	dword_5d4594_1046656 = nox_xxx_guiFontHeightMB_43F320(0) + 2;
	*getMemU32Ptr(0x5D4594, 1047508) = 0;
	dword_5d4594_1047512 = 0;
	v9 = 2 * (141 / dword_5d4594_1046656) - 2;
	if (dword_5d4594_1046868 == 1) {
		for (i = nox_xxx_bookGetFirstCreMB_427300(); i; i = nox_xxx_bookGetNextCre_427320(i)) {
			if (nox_common_gameFlags_check_40A5C0(0x2000) && !nox_common_gameFlags_check_40A5C0(4096) ||
				player->beast_scroll_lvl[i]) {
				if (nox_xxx_bookCreatureTest_4D70C0(i)) {
					v3 = *getMemU32Ptr(0x5D4594, 1047508);
					v1 = 1;
					*getMemU32Ptr(0x5D4594, 1046960 + 4 * *getMemU32Ptr(0x5D4594, 1047508)) = i;
					*getMemU32Ptr(0x5D4594, 1047508) = v3 + 1;
				}
			}
		}
	} else if (a1) {
		for (j = nox_xxx_spellFirstValid_424AD0(); j; j = nox_xxx_spellNextValid_424AF0(j)) {
			if (j != 34 && !nox_xxx_playerCheckSpellClass_57AEA0(a1, j) &&
				(nox_common_gameFlags_check_40A5C0(0x2000) && !nox_common_gameFlags_check_40A5C0(4096) ||
				 player->spell_lvl[j])) {
				if (nox_xxx_spellHasFlags_424A50(j, 0x15000)) {
					++dword_5d4594_1047512;
				}
				if (!nox_xxx_spellHasFlags_424A50(j, 0x2000)) {
					v7 = *getMemU32Ptr(0x5D4594, 1047508);
					v1 = 1;
					*getMemU32Ptr(0x5D4594, 1046960 + 4 * *getMemU32Ptr(0x5D4594, 1047508)) = j;
					*getMemU32Ptr(0x5D4594, 1047508) = v7 + 1;
				}
			}
		}
	} else {
		for (k = nox_xxx_bookFirstKnownAbil_425330(); k; k = nox_xxx_bookNextKnownAbil_425350(k)) {
			if (nox_common_gameFlags_check_40A5C0(0x2000) && !nox_common_gameFlags_check_40A5C0(4096) ||
				player->spell_lvl[k]) {
				v5 = *getMemU32Ptr(0x5D4594, 1047508);
				v1 = 1;
				*getMemU32Ptr(0x5D4594, 1046960 + 4 * *getMemU32Ptr(0x5D4594, 1047508)) = k;
				*getMemU32Ptr(0x5D4594, 1047508) = v5 + 1;
			}
		}
	}
	*getMemU32Ptr(0x5D4594, 1046940) = (*getMemU32Ptr(0x5D4594, 1047508) - dword_5d4594_1047512) / v9 + 1;
	qsort(getMemAt(0x5D4594, 1046960), *getMemU32Ptr(0x5D4594, 1047508) - dword_5d4594_1047512, 4u,
		  game2_book_old_nox_xxx_guiSpellSortFn_45ABC0);
	return v1;
}

static void game2_book_old_nox_xxx_book_45B010(int a1) {
	unsigned int result; // eax

	result = nox_gui_xxx_check_446360();
	if (result) {
		return;
	}
	result = sub_4AE3D0();
	if (result) {
		return;
	}
	if (a1 || (nox_xxx_get_57AF20() == 0)) {
		dword_5d4594_1046864 = 1;
		dword_5d4594_1046868 = dword_5d4594_1046872 != 0;
		nox_xxx_wndShowModalMB_46A8C0(nox_win_unk1);
		nox_xxx_clientPlaySoundSpecial_452D80(786, 100);
	}
}

static int game2_book_old_nox_xxx_bookWndProc_45B070(int a1, int a2) {
	if (dword_5d4594_1047520 == 1) {
		return 1;
	}
	if (a2 == 10) {
		game2_book_old_nox_client_toggleSpellbook_45AC70();
		return 1;
	}
	if (game2_book_old_nox_client_localPlayerHasFlags30(2) && !sub_478030() && !sub_47A260()) {
		return 1;
	}
	if (a2 != 5) {
		return 0;
	}
	nox_window_set_hidden(dword_5d4594_1046944, 0);
	if (nox_xxx_aNox_cfg_0_587000_132132) {
		if (dword_5d4594_1046936 + 1 < *getMemIntPtr(0x5D4594, 1046940)) {
			++dword_5d4594_1046936;
		} else {
			dword_5d4594_1046932 = 0;
			nox_xxx_aNox_cfg_0_587000_132132 = 0;
			nox_window_set_hidden(dword_5d4594_1046952, 0);
		}
		if (dword_5d4594_1046872) {
			game2_book_old_nox_book_anim_set_end(dword_5d4594_1046924, game2_book_old_nox_xxx_bookClickCreature_45B200);
		} else {
			game2_book_old_nox_book_anim_set_end(dword_5d4594_1046924, game2_book_old_nox_xxx_bookClickSpell_45B1F0);
		}
		dword_5d4594_1046868 = 2;
		game2_book_old_nox_book_anim_start(dword_5d4594_1046924);
		nox_xxx_clientPlaySoundSpecial_452D80(788, 100);
		return 1;
	}
	if (dword_5d4594_1046932 < *getMemIntPtr(0x5D4594, 1047508) - dword_5d4594_1047512 - 1) {
		++dword_5d4594_1046932;
		if (dword_5d4594_1046872) {
			game2_book_old_nox_book_anim_set_end(dword_5d4594_1046924, game2_book_old_nox_xxx_bookClickCreature_45B200);
		} else {
			game2_book_old_nox_book_anim_set_end(dword_5d4594_1046924, game2_book_old_nox_xxx_bookClickSpell_45B1F0);
		}
		dword_5d4594_1046868 = 2;
		game2_book_old_nox_book_anim_start(dword_5d4594_1046924);
		nox_xxx_clientPlaySoundSpecial_452D80(788, 100);
		return 1;
	}
	return 1;
}

static int game2_book_old_nox_xxx_bookClickSpell_45B1F0() {
	int result; // eax

	result = 0;
	dword_5d4594_1046868 = 0;
	dword_5d4594_1046872 = 0;
	return result;
}

static int game2_book_old_nox_xxx_bookClickCreature_45B200() {
	int result; // eax

	result = 1;
	dword_5d4594_1046868 = 1;
	dword_5d4594_1046872 = 1;
	return result;
}

static int game2_book_old_nox_xxx_book_45B210(int a1, int a2) {
	if (dword_5d4594_1047520 == 1) {
		return 1;
	}
	if (a2 == 10) {
		game2_book_old_nox_client_toggleSpellbook_45AC70();
		return 1;
	}
	if (game2_book_old_nox_client_localPlayerHasFlags30(2) && !sub_478030() && !sub_47A260()) {
		return 1;
	}
	if (a2 != 5) {
		return 0;
	}
	nox_window_set_hidden(dword_5d4594_1046948, 0);
	if (!nox_xxx_aNox_cfg_0_587000_132132) {
		if (*(int*)&dword_5d4594_1046932 < *getMemIntPtr(0x5D4594, 1047508) - *(int*)&dword_5d4594_1047512) {
			if (*(int*)&dword_5d4594_1046932 <= 0) {
				nox_xxx_aNox_cfg_0_587000_132132 = 1;
				dword_5d4594_1046936 = *getMemU32Ptr(0x5D4594, 1046940) - 1;
				nox_window_set_hidden(dword_5d4594_1046952, 1);
			} else {
				--dword_5d4594_1046932;
			}
		} else {
			nox_xxx_aNox_cfg_0_587000_132132 = 1;
			dword_5d4594_1046936 = 0;
			nox_window_set_hidden(dword_5d4594_1046952, 1);
		}
	} else {
		if (!dword_5d4594_1046936) {
			return 1;
		}
		--dword_5d4594_1046936;
	}
	if (dword_5d4594_1046872) {
		game2_book_old_nox_book_anim_set_end(dword_5d4594_1046928, game2_book_old_nox_xxx_bookClickCreature_45B200);
	} else {
		game2_book_old_nox_book_anim_set_end(dword_5d4594_1046928, game2_book_old_nox_xxx_bookClickSpell_45B1F0);
	}
	dword_5d4594_1046868 = 3;
	game2_book_old_nox_book_anim_start(dword_5d4594_1046928);
	nox_xxx_clientPlaySoundSpecial_452D80(788, 100);
	return 1;
}

static int game2_book_old_nox_xxx_bookChildWndProcMB_45B360(uint32_t* a1, unsigned int a2) {
	int v2; // esi
	nox_playerInfo* player = (nox_playerInfo*)dword_5d4594_1047516;

	v2 = 0;
	if (dword_5d4594_1047520 == 1 ||
		(game2_book_old_nox_client_localPlayerHasFlags30(2) && !sub_478030() && !sub_47A260())) {
		return 1;
	}
	if (a2 != 5) {
		if (a2 <= 5 || a2 > 7) {
			return 0;
		}
		return 1;
	}
	if (*a1 == 1310) {
		if (dword_5d4594_1046872) {
			dword_5d4594_1046868 = 0;
			dword_5d4594_1046872 = 0;
			if (!game2_book_old_nox_xxx_guiSpellSortList_45ADF0(player->info.playerClass)) {
				dword_5d4594_1046868 = 1;
				dword_5d4594_1046872 = 1;
				game2_book_old_nox_xxx_guiSpellSortList_45ADF0(player->info.playerClass);
				nox_xxx_clientPlaySoundSpecial_452D80(925, 100);
				return 1;
			}
			v2 = 1;
		}
		if (!nox_xxx_aNox_cfg_0_587000_132132 || dword_5d4594_1046936 || v2 == 1) {
			nox_xxx_aNox_cfg_0_587000_132132 = 1;
			dword_5d4594_1046936 = 0;
			nox_xxx_clientPlaySoundSpecial_452D80(788, 100);
			dword_5d4594_1046868 = 3;
			game2_book_old_nox_book_anim_start(dword_5d4594_1046928);
			game2_book_old_nox_book_anim_set_end(dword_5d4594_1046928, game2_book_old_nox_xxx_bookClickSpell_45B1F0);
		nox_window_set_hidden(dword_5d4594_1046952, 1);
		nox_window_set_hidden(dword_5d4594_1046948, 0);
			nox_xxx_wndClearCaptureMain_46ADE0(nox_win_unk1);
			nox_xxx_bookSpellDnDclear_477660();
		}
		return 1;
	}
	if (*a1 != 1320) {
		return 1;
	}
	if (dword_5d4594_1046872 != 1) {
		dword_5d4594_1046868 = 1;
		dword_5d4594_1046872 = 1;
		if (!game2_book_old_nox_xxx_guiSpellSortList_45ADF0(player->info.playerClass)) {
			dword_5d4594_1046868 = 0;
			dword_5d4594_1046872 = 0;
			game2_book_old_nox_xxx_guiSpellSortList_45ADF0(player->info.playerClass);
			nox_xxx_clientPlaySoundSpecial_452D80(925, 100);
			return 1;
		}
		v2 = 1;
	}
	if (nox_xxx_aNox_cfg_0_587000_132132 && !dword_5d4594_1046936 && v2 != 1) {
		return 1;
	}
	nox_xxx_aNox_cfg_0_587000_132132 = 1;
	dword_5d4594_1046936 = 0;
	nox_xxx_clientPlaySoundSpecial_452D80(788, 100);
	if (v2 == 1) {
		dword_5d4594_1046868 = 2;
		game2_book_old_nox_book_anim_start(dword_5d4594_1046924);
		game2_book_old_nox_book_anim_set_end(dword_5d4594_1046924, game2_book_old_nox_xxx_bookClickCreature_45B200);
	} else {
		dword_5d4594_1046868 = 3;
		game2_book_old_nox_book_anim_start(dword_5d4594_1046928);
		game2_book_old_nox_book_anim_set_end(dword_5d4594_1046928, game2_book_old_nox_xxx_bookClickCreature_45B200);
	}
	nox_window_set_hidden(dword_5d4594_1046952, 1);
	nox_window_set_hidden(dword_5d4594_1046948, 0);
	nox_xxx_wndClearCaptureMain_46ADE0(nox_win_unk1);
	nox_xxx_bookSpellDnDclear_477660();
	return 1;
}

static int game2_book_old_nox_xxx_bookListWndProc_45B5F0(nox_window* a1, unsigned int a2, unsigned int a3) {
	int v3;        // ebp
	int v4;        // edi
	int v5;        // esi
	int result;    // eax
	bool v7;       // sf
	signed int v8; // ebp
	int v9;        // eax
	int v10;       // ecx
	char v11;      // al
	int v12;       // eax
	int v13;       // eax
	int v14;       // [esp+10h] [ebp-4h]
	nox_playerInfo* player = (nox_playerInfo*)dword_5d4594_1047516;

	v3 = a3 >> 16;
	v4 = (unsigned short)a3;
	if (dword_5d4594_1047520 == 1 ||
		(game2_book_old_nox_client_localPlayerHasFlags30(2) && !sub_478030() && !sub_47A260())) {
		return 1;
	}
	if (!nox_xxx_aNox_cfg_0_587000_132132) {
		if (a2 >= 5) {
			if (a2 <= 8) {
				return 1;
			}
			if (a2 == 11) {
				game2_book_old_nox_client_toggleSpellbook_45AC70();
				return 1;
			}
		}
		return 0;
	}
	dword_5d4594_1046656 = nox_xxx_guiFontHeightMB_43F320(0) + 2;
	v5 = 141 / dword_5d4594_1046656 - 1;
	nox_gui_getWindowOffs_46AA20(a1, (unsigned int*)&v14, &a3);
	switch (a2) {
	case 5u:
		if (game2_book_old_nox_client_localPlayerHasFlags30(2)) {
			return 1;
		}
		dword_5d4594_1047536 = v3;
		v7 = (int)(-19 - a3 + v3) < 0;
		v8 = -19 - a3 + v3;
		dword_5d4594_1047532 = v4;
		if (v7) {
			nox_xxx_aNox_cfg_0_587000_132136 = -1;
			return 1;
		}
		v9 = v8 / *(int*)&dword_5d4594_1046656;
		if (v8 / *(int*)&dword_5d4594_1046656 >= v5) {
			nox_xxx_aNox_cfg_0_587000_132136 = -1;
			return 1;
		}
		if (v4 - v14 > 145) {
			v9 += v5;
		}
		nox_xxx_aNox_cfg_0_587000_132136 = 2 * v5 * dword_5d4594_1046936 + v9;
		nox_xxx_wndSetCaptureMain_46ADC0(a1);
		v10 = nox_xxx_aNox_cfg_0_587000_132136;
		if (nox_xxx_aNox_cfg_0_587000_132136 >= *getMemIntPtr(0x5D4594, 1047508) - dword_5d4594_1047512) {
			nox_xxx_aNox_cfg_0_587000_132136 = -1;
			return 1;
		}
		v11 = player->info.playerClass;
		if (!v11 && !dword_5d4594_1046868) {
			dword_5d4594_1047528 = *getMemU32Ptr(0x5D4594, 1046960 + 4 * nox_xxx_aNox_cfg_0_587000_132136);
			nox_xxx_bookSaveSpellForDragDrop_477640(*(int*)&dword_5d4594_1047528, 1);
			nox_xxx_clientPlaySoundSpecial_452D80(793, 100);
			return 1;
		}
		if (v11 != 2 || dword_5d4594_1046868 != 1) {
			if (dword_5d4594_1046868) {
				return 1;
			}
			dword_5d4594_1047528 = *getMemU32Ptr(0x5D4594, 1046960 + 4 * nox_xxx_aNox_cfg_0_587000_132136);
			nox_xxx_bookSaveSpellForDragDrop_477640(*(int*)&dword_5d4594_1047528, 1);
			nox_xxx_clientPlaySoundSpecial_452D80(793, 100);
			result = 1;
		} else {
			if (!player->spell_lvl[134]) {
				if (!nox_common_gameFlags_check_40A5C0(0x2000) || nox_common_gameFlags_check_40A5C0(4096)) {
					nox_xxx_wndClearCaptureMain_46ADE0(a1);
					game2_book_old_nox_xxx_bookMoveToPage_45B930(*(int*)&nox_xxx_aNox_cfg_0_587000_132136);
					dword_5d4594_1047528 = 0;
					nox_xxx_aNox_cfg_0_587000_132136 = -1;
					return 1;
				}
				v10 = nox_xxx_aNox_cfg_0_587000_132136;
			}
			dword_5d4594_1047528 = *getMemU32Ptr(0x5D4594, 1046960 + 4 * v10) + 74;
			nox_xxx_bookSaveSpellForDragDrop_477640(*(int*)&dword_5d4594_1047528, 1);
			nox_xxx_clientPlaySoundSpecial_452D80(793, 100);
			result = 1;
		}
		break;
	case 6u:
	case 7u:
		if ((int)nox_xxx_aNox_cfg_0_587000_132136 >= 0) {
			v12 = dword_5d4594_1047532 - v4;
			if (*(int*)&dword_5d4594_1047532 - v4 < 0) {
				v12 = v4 - dword_5d4594_1047532;
			}
			if (v12 >= 5) {
				nox_xxx_bookSpellDrop_45DCA0(*(int*)&dword_5d4594_1047528, 0, v4, v3, 0);
			} else {
				v13 = dword_5d4594_1047536 - v3;
				if (*(int*)&dword_5d4594_1047536 - v3 < 0) {
					v13 = v3 - dword_5d4594_1047536;
				}
				if (v13 >= 5) {
					nox_xxx_bookSpellDrop_45DCA0(*(int*)&dword_5d4594_1047528, 0, v4, v3, 0);
				} else {
					game2_book_old_nox_xxx_bookMoveToPage_45B930(*(int*)&nox_xxx_aNox_cfg_0_587000_132136);
				}
			}
		}
		if (nox_xxx_wndGetCaptureMain_46AE00()) {
			nox_xxx_wndClearCaptureMain_46ADE0(a1);
		}
		nox_xxx_bookSpellDnDclear_477660();
		return 1;
	case 8u:
		return 1;
	case 0xBu:
		game2_book_old_nox_client_toggleSpellbook_45AC70();
		return 1;
	default:
		return 0;
	}
	return result;
}

static int game2_book_old_nox_xxx_bookMoveToPage_45B930(int a1) {
	nox_xxx_aNox_cfg_0_587000_132132 = 0;
	dword_5d4594_1046932 = a1;
	dword_5d4594_1046936 = 99;
	if (dword_5d4594_1046872) {
		game2_book_old_nox_book_anim_set_end(dword_5d4594_1046924, game2_book_old_nox_xxx_bookClickCreature_45B200);
	} else {
		game2_book_old_nox_book_anim_set_end(dword_5d4594_1046924, game2_book_old_nox_xxx_bookClickSpell_45B1F0);
	}
	dword_5d4594_1046868 = 2;
	game2_book_old_nox_book_anim_start(dword_5d4594_1046924);
	nox_xxx_clientPlaySoundSpecial_452D80(788, 100);
	nox_window_set_hidden(dword_5d4594_1046952, 0);
	nox_window_set_hidden(dword_5d4594_1046944, 0);
	return nox_window_set_hidden(dword_5d4594_1046948, 0);
}

static int game2_book_old_nox_xxx_book_45BD30(int a1, int a2) { return 1; }

static int game2_book_old_nox_xxx_bookInit_45B9D0() {
	nox_window* v1;
	nox_window* v2;
	nox_video_bag_image_t* v3;
	nox_video_bag_image_t* v4;
	nox_video_bag_image_t* image;

	dword_5d4594_1047516 = dword_8531A0_2576;
	image = nox_xxx_gLoadImg_42F970("ArrowNW");
	nox_book_arrow_images[0] = image;
	if (!image) {
		return 0;
	}

	image = nox_xxx_gLoadImg_42F970("ArrowN");
	nox_book_arrow_images[1] = image;
	if (!image) {
		return 0;
	}

	image = nox_xxx_gLoadImg_42F970("ArrowNE");
	nox_book_arrow_images[2] = image;
	if (!image) {
		return 0;
	}

	image = nox_xxx_gLoadImg_42F970("ArrowW");
	nox_book_arrow_images[3] = image;
	if (!image) {
		return 0;
	}

	image = nox_xxx_gLoadImg_42F970("ArrowE");
	nox_book_arrow_images[4] = image;
	if (!image) {
		return 0;
	}

	image = nox_xxx_gLoadImg_42F970("ArrowSW");
	nox_book_arrow_images[5] = image;
	if (!image) {
		return 0;
	}

	image = nox_xxx_gLoadImg_42F970("ArrowS");
	nox_book_arrow_images[6] = image;
	if (!image) {
		return 0;
	}

	image = nox_xxx_gLoadImg_42F970("ArrowSE");
	nox_book_arrow_images[7] = image;
	if (!image) {
		return 0;
	}

	nox_book_background_image = nox_xxx_gLoadImg_42F970("BookOfKnowledge");
	if (!nox_book_background_image) {
		return 0;
	}

	nox_book_guide_tab_lit_image = nox_xxx_gLoadImg_42F970("GuideTabLit");
	if (!nox_book_guide_tab_lit_image) {
		return 0;
	}

	nox_book_spell_tab_lit_image = nox_xxx_gLoadImg_42F970("SpellTabLit");
	if (!nox_book_spell_tab_lit_image) {
		return 0;
	}

	dword_5d4594_1046924 = nox_xxx_gLoadAnim_42FA20("BookPageForward");
	if (!dword_5d4594_1046924) {
		return 0;
	}

	game2_book_old_nox_book_anim_set_end(dword_5d4594_1046924, game2_book_old_nox_xxx_bookClickSpell_45B1F0);
	dword_5d4594_1046928 = nox_xxx_gLoadAnim_42FA20("BookPageBackward");
	if (!dword_5d4594_1046928) {
		return 0;
	}

	game2_book_old_nox_book_anim_set_end(dword_5d4594_1046928, game2_book_old_nox_xxx_bookClickSpell_45B1F0);
	nox_win_unk1 = nox_window_new(0, 1196, 5, nox_win_height - 323, 285, 168, 0);
	nox_window_set_all_funcs(nox_win_unk1, game2_book_old_nox_xxx_bookListWndProc_45B5F0, nox_xxx_bookDrawList_45BD40, 0);
	nox_window_set_hidden(nox_win_unk1, 1);
	v1 = nox_window_new(nox_win_unk1, 8, 257, 15, 27, 40, 0);
	if (!v1) {
		return 0;
	}

	nox_window_set_all_funcs(v1, game2_book_old_nox_xxx_bookChildWndProcMB_45B360, game2_book_old_nox_xxx_book_45BD30,
								 nox_xxx_book_45CF00);
	v1->id = 1320;
	v2 = nox_window_new(nox_win_unk1, 8, 253, 61, 27, 40, 0);
	if (!v2) {
		return 0;
	}

	nox_window_set_all_funcs(v2, game2_book_old_nox_xxx_bookChildWndProcMB_45B360, game2_book_old_nox_xxx_book_45BD30,
								 nox_xxx_book_45CF00);
	v2->id = 1310;
	dword_5d4594_1046944 = nox_window_new(nox_win_unk1, 136, 24, 138, 20, 20, 0);
	nox_window_set_all_funcs(dword_5d4594_1046944, game2_book_old_nox_xxx_book_45B210, 0, 0);
	v3 = nox_xxx_gLoadImg_42F970("ArrowW");
	nox_xxx_wndSetIcon_46AE60(dword_5d4594_1046944, v3);
	dword_5d4594_1046948 = nox_window_new(nox_win_unk1, 136, 233, 138, 20, 20, 0);
	nox_window_set_all_funcs(dword_5d4594_1046948, game2_book_old_nox_xxx_bookWndProc_45B070, 0, 0);
	v4 = nox_xxx_gLoadImg_42F970("ArrowE");
	nox_xxx_wndSetIcon_46AE60(dword_5d4594_1046948, v4);
	dword_5d4594_1046952 = nox_window_new(nox_win_unk1, 8, 63, 19, 30, 30, 0);
	nox_window_set_all_funcs(dword_5d4594_1046952, game2_book_old_nox_xxx_bookWndFn_45CC10,
								 game2_book_old_nox_xxx_bookDrawIconFn_45CB30, 0);
	nox_window_set_hidden(dword_5d4594_1046952, 1);
	dword_5d4594_1046956 = nox_window_new(0, 40, 0, 0, 30, 30, 0);
	nox_window_set_all_funcs(dword_5d4594_1046956, 0, nox_xxx_bookDrawFn_45C7D0, 0);
	nox_window_set_hidden(dword_5d4594_1046956, 1);
	dword_5d4594_1046868 = 0;
	dword_5d4594_1046872 = 0;
	nox_xxx_aNox_cfg_0_587000_132132 = 1;
	dword_5d4594_1046936 = 0;

	return 1;
}

static int game2_book_old_nox_xxx_bookDrawIconFn_45CB30(uint32_t* a1) {
	nox_video_bag_image_t* v1; // eax
	nox_video_bag_image_t* v2; // esi
	int v4; // [esp+4h] [ebp-8h]
	int v5; // [esp+8h] [ebp-4h]
	nox_playerInfo* player = (nox_playerInfo*)dword_5d4594_1047516;

	if (dword_5d4594_1046868) {
		if (dword_5d4594_1046868 != 1 || player->info.playerClass != 2 ||
			!player->spell_lvl[134] &&
				(!nox_common_gameFlags_check_40A5C0(0x2000) || nox_common_gameFlags_check_40A5C0(4096))) {
			return 1;
		}
		v1 = nox_xxx_spellIcon_424A90(*getMemU32Ptr(0x5D4594, 1046960 + 4 * dword_5d4594_1046932) + 74);
	} else if (player->info.playerClass) {
		v1 = nox_xxx_spellIcon_424A90(*getMemU32Ptr(0x5D4594, 1046960 + 4 * dword_5d4594_1046932));
	} else {
		v1 = nox_xxx_spellGetAbilityIcon_425310(*getMemU32Ptr(0x5D4594, 1046960 + 4 * dword_5d4594_1046932), 0);
	}
	v2 = v1;
	if (v1) {
		nox_client_wndGetPosition_46AA60(a1, (unsigned int*)&v5, (unsigned int*)&v4);
		nox_client_drawImageAt_47D2C0(v2, v5, v4);
	}
	return 1;
}

static int game2_book_old_nox_xxx_bookWndFn_45CC10(uint32_t* a1, int a2, unsigned int a3) {
	int result; // eax
	int v4;     // ebp
	char v5;    // al
	char v6;    // al
	nox_playerInfo* player = (nox_playerInfo*)dword_5d4594_1047516;

	if (game2_book_old_nox_client_localPlayerHasFlags30(2)) {
		return 1;
	}
	v4 = a3 >> 16;
	if (dword_5d4594_1046868 == 1) {
		v5 = player->info.playerClass;
		if (v5 == 1 || !v5 ||
			!nox_common_gameFlags_check_40A5C0(0x2000) && !nox_common_gameFlags_check_40A5C0(4096) &&
				!player->spell_lvl[134]) {
			return 0;
		}
	}
	switch (a2) {
	case 5:
		if (nox_xxx_wndGetCaptureMain_46AE00() || dword_5d4594_1047540) {
			return 1;
		}
		v6 = player->info.playerClass;
		if (!v6 && !dword_5d4594_1046868) {
			dword_5d4594_1047540 = *getMemU32Ptr(0x5D4594, 1046960 + 4 * dword_5d4594_1046932);
			nox_xxx_bookSaveSpellForDragDrop_477640(*(int*)&dword_5d4594_1047540, 1);
			nox_xxx_wndSetCaptureMain_46ADC0((nox_window*)a1);
			nox_xxx_clientPlaySoundSpecial_452D80(793, 100);
			return 1;
		}
		if (v6 == 2 && dword_5d4594_1046868 == 1) {
			dword_5d4594_1047540 = *getMemU32Ptr(0x5D4594, 1046960 + 4 * dword_5d4594_1046932) + 74;
			nox_xxx_bookSaveSpellForDragDrop_477640(*(int*)&dword_5d4594_1047540, 1);
			nox_xxx_wndSetCaptureMain_46ADC0((nox_window*)a1);
			nox_xxx_clientPlaySoundSpecial_452D80(793, 100);
			return 1;
		}
		if (dword_5d4594_1046868) {
			return 1;
		}
		dword_5d4594_1047540 = *getMemU32Ptr(0x5D4594, 1046960 + 4 * dword_5d4594_1046932);
		if (!nox_xxx_spellHasFlags_424A50(*(int*)&dword_5d4594_1047540, 0x15000)) {
			nox_xxx_bookSaveSpellForDragDrop_477640(*(int*)&dword_5d4594_1047540, 1);
			nox_xxx_wndSetCaptureMain_46ADC0((nox_window*)a1);
			nox_xxx_clientPlaySoundSpecial_452D80(793, 100);
			return 1;
		}
		dword_5d4594_1047540 = 0;
		return 1;
	case 6:
	case 7:
		if (!dword_5d4594_1047540) {
			return 1;
		}
		if (!nox_xxx_wndPointInWnd_46AAB0(a1, (unsigned short)a3, v4)) {
			if (nox_xxx_guiSpell_460650()) {
				game2_book_old_nox_xxx_guiSpellTargetClickSet_45D9D0(*(int*)&dword_5d4594_1047540);
			} else if (sub_4611A0()) {
				sub_4611B0();
			} else {
				nox_xxx_bookSpellDrop_45DCA0(*(int*)&dword_5d4594_1047540, 0, (unsigned short)a3, v4, 0);
			}
			dword_5d4594_1047540 = 0;
			nox_xxx_wndClearCaptureMain_46ADE0((nox_window*)a1);
			nox_xxx_bookSpellDnDclear_477660();
			return 1;
		}
		if (player->info.playerClass) {
			if (nox_xxx_guiSpell_460650()) {
				nox_xxx_guiSpellSetCursor_45DF60(0, 0);
				dword_5d4594_1047540 = 0;
				nox_xxx_wndClearCaptureMain_46ADE0((nox_window*)a1);
				nox_xxx_bookSpellDnDclear_477660();
				return 1;
			}
			if (!nox_xxx_spellHasFlags_424A50(*(int*)&dword_5d4594_1047540, 0x600)) {
				nox_xxx_guiSpellSetCursor_45DF60(*(int*)&dword_5d4594_1047540, 0);
				nox_xxx_bookSpellDnDclear_477660();
				return 1;
			}
			if (nox_xxx_guiSpellSetCursor_45DF60(*(int*)&dword_5d4594_1047540, 1)) {
				dword_5d4594_1047540 = 0;
				nox_xxx_wndClearCaptureMain_46ADE0((nox_window*)a1);
				nox_xxx_bookSpellDnDclear_477660();
				return 1;
			}
		} else {
			if (sub_4611A0()) {
				sub_45DFC0(0);
				dword_5d4594_1047540 = 0;
				nox_xxx_wndClearCaptureMain_46ADE0((nox_window*)a1);
				nox_xxx_bookSpellDnDclear_477660();
				return 1;
			}
			if (sub_45DFC0(*(int*)&dword_5d4594_1047540)) {
				dword_5d4594_1047540 = 0;
				nox_xxx_wndClearCaptureMain_46ADE0((nox_window*)a1);
				nox_xxx_bookSpellDnDclear_477660();
				return 1;
			}
		}
		nox_xxx_bookSpellDnDclear_477660();
		return 1;
	case 8:
		return 1;
	default:
		return 0;
	}
	return result;
}

static int game2_book_old_sub_45CFC0() { return ((unsigned char)~nox_xxx_wndGetFlags_46ADA0(nox_win_unk1) >> 4) & 1; }

static void game2_book_old_nox_xxx_netSpellRewardCli_45CFE0(int a1, int a2, int a3, int a4) {
	int v5;             // ebp
	int v6;             // esi
	uint32_t* v7;       // edi
	int v8;             // eax
	int v9;             // esi
	uint32_t* v10;      // eax
	int v11;            // [esp+14h] [ebp+4h]

	nox_playerInfo* player = (nox_playerInfo*)dword_8531A0_2576;
	if (!player) {
		return;
	}
	v5 = a1;
	if (nox_xxx_playerCheckSpellClass_57AEA0(player->info.playerClass, a1) == 9) {
		return;
	}
	player->spell_lvl[a1] = a2;
	if (nox_xxx_spellHasFlags_424A50(a1, 0x1000)) {
		v11 = 0x2000;
		goto LABEL_9;
	} else if (nox_xxx_spellHasFlags_424A50(a1, 0x4000)) {
		v11 = 0x8000;
		goto LABEL_9;
	} else if (nox_xxx_spellHasFlags_424A50(a1, 0x10000)) {
		v11 = 0x20000;
		goto LABEL_9;
	} else {
		goto LABEL_A;
	}
LABEL_9:
	v6 = 1;
	v7 = &player->spell_lvl[1];
	do {
		if (nox_xxx_spellHasFlags_424A50(v6, v11) && nox_xxx_spellIsValid_424B50(v6)) {
			*v7 = a2;
		}
		++v6;
		++v7;
	} while (v6 < 137);
LABEL_A:
	dword_5d4594_1046868 = 0;
	dword_5d4594_1046872 = 0;
	game2_book_old_nox_xxx_guiSpellSortList_45ADF0(player->info.playerClass);
	if (v5 == 34) {
		v8 = 1;
		if (player->spell_lvl[34] != 1 || !a3) {
			v8 = 0;
		}
		nox_xxx_quickbarAddTrap_460EC0(v8);
	} else if (a3) {
		v9 = 0;
		v10 = getMemU32Ptr(0x5D4594, 1046960);
		do {
			if (*v10 == v5) {
				break;
			}
			++v10;
			++v9;
		} while (v10 < getMemU32Ptr(0x5D4594, 1047508));
		if (v9 != 137) {
			game2_book_old_nox_xxx_bookHideMB_45ACA0(0);
			game2_book_old_nox_xxx_bookMoveToPage_45B930(v9);
			game2_book_old_nox_xxx_book_45B010(0);
			nox_xxx_bookRewardCli_499CF0((int*)2, v5, a4);
		}
	}
}

static int* game2_book_old_nox_xxx_bookSetForward_45D200(int* a1, int a2, int2* a3) {
	int* result; // eax
	int v4;      // ecx
	int v5;      // ecx

	nox_xxx_bookShowMB_45AD70(1);
	nox_window_setPos_46A9B0(nox_win_unk1, a3->field_0, a3->field_4);
	result = a1;
	if (a1 == (int*)2 || a1 == (int*)4) {
		v5 = 0;
		result = getMemIntPtr(0x5D4594, 1046960);
		do {
			if (*result == a2) {
				break;
			}
			++result;
			++v5;
		} while (result < getMemIntPtr(0x5D4594, 1047508));
		if (v5 != 137) {
			result = (int*)game2_book_old_nox_xxx_bookMoveToPage_45B930(v5);
		}
	} else if (a1 == (int*)3) {
		v4 = 0;
		result = getMemIntPtr(0x5D4594, 1046960);
		do {
			if (*result == a2) {
				break;
			}
			++result;
			++v4;
		} while (result < getMemIntPtr(0x5D4594, 1046984));
		if (v4 != 6) {
			result = (int*)game2_book_old_nox_xxx_bookMoveToPage_45B930(v4);
		}
	}
	return result;
}

static void game2_book_old_nox_xxx_abilityReward_45D290(int a1, char* a2, int a3) {
	nox_playerInfo* player; // eax
	int v4;                // esi
	int32_t* entry;         // eax

	player = nox_common_playerInfoGetByID_417040(nox_player_netCode_85319C);
	if (player) {
		dword_5d4594_1046868 = 0;
		dword_5d4594_1046872 = 0;
		game2_book_old_nox_xxx_guiSpellSortList_45ADF0(player->info.playerClass);
		if (a2) {
			v4 = 0;
			entry = getMemI32Ptr(0x5D4594, 1046960);
			do {
				if (*entry == a1) {
					break;
				}
				++entry;
				++v4;
			} while (entry < getMemI32Ptr(0x5D4594, 1046984));
			if (v4 != 6) {
				game2_book_old_nox_xxx_bookHideMB_45ACA0(0);
				game2_book_old_nox_xxx_bookMoveToPage_45B930(v4);
				game2_book_old_nox_xxx_book_45B010(0);
				if (a3) {
					nox_xxx_bookRewardCli_499CF0((int*)3, a1, a3);
				}
			}
		}
	}
}

static int game2_book_old_sub_45D320(int a1) {
	nox_playerInfo* player; // ebx
	int result;            // eax
	int v3;                // ebp
	int v4;                // esi
	uint32_t* v5;          // edi

	player = (nox_playerInfo*)dword_8531A0_2576;
	result = game2_book_old_nox_xxx_bookHideMB_45ACA0(1);
	if (player) {
		player->spell_lvl[a1] = 0;
		sub_461360(a1);
		if (nox_xxx_spellHasFlags_424A50(a1, 0x1000)) {
			v3 = 0x2000;
		} else if (nox_xxx_spellHasFlags_424A50(a1, 0x4000)) {
			v3 = 0x8000;
		} else {
			if (!nox_xxx_spellHasFlags_424A50(a1, 0x10000)) {
				return game2_book_old_nox_xxx_guiSpellSortList_45ADF0(player->info.playerClass);
			}
			v3 = 0x20000;
		}
		v4 = 1;
		v5 = &player->spell_lvl[1];
		do {
			if (nox_xxx_spellHasFlags_424A50(v4, v3)) {
				if (nox_xxx_spellIsValid_424B50(v4)) {
					*v5 = 0;
					sub_461360(v4);
				}
			}
			++v4;
			++v5;
		} while (v4 < 137);
		return game2_book_old_nox_xxx_guiSpellSortList_45ADF0(player->info.playerClass);
	}
	return result;
}

static int game2_book_old_sub_45D400(int a1) {
	nox_playerInfo* player; // edi
	int result;            // eax

	player = (nox_playerInfo*)dword_8531A0_2576;
	result = game2_book_old_nox_xxx_bookHideMB_45ACA0(1);
	if (player) {
		player->beast_scroll_lvl[a1] = 0;
		sub_461360(a1 + 74);
		for (int group_index = 0; group_index < 41; ++group_index) {
			uint32_t* group = getMemPtr(0x587000, 132124 + 4 * group_index);
			if (!group) {
				break;
			}
			if (a1 == group[0]) {
				for (uint32_t* member = &group[1]; *member; ++member) {
					player->beast_scroll_lvl[*member] = 0;
					sub_461360(*member + 74);
				}
			}
		}
		result = game2_book_old_nox_xxx_guiSpellSortList_45ADF0(player->info.playerClass);
	}
	return result;
}

static char* game2_book_old_nox_xxx_clientQuestDisableAbility_45D4A0(int a1) {
	char* result; // eax

	game2_book_old_nox_xxx_bookHideMB_45ACA0(1);
	nox_xxx_netAbilityRewardCli_4611E0(a1, 0, 0);
	nox_playerInfo* local_player = (nox_playerInfo*)dword_5d4594_1047516;
	if (local_player) {
		local_player->spell_lvl[a1] = 0;
	}
	sub_461360(a1);
	nox_playerInfo* player = nox_common_playerInfoGetByID_417040(nox_player_netCode_85319C);
	result = (char*)player;
	if (player) {
		result = (char*)(intptr_t)game2_book_old_nox_xxx_guiSpellSortList_45ADF0(player->info.playerClass);
	}
	return result;
}

static int game2_book_old_sub_45D500(int a1) {
	int result; // eax

	result = dword_5d4594_1046864;
	if (a1) {
		if (dword_5d4594_1046864) {
			result = nox_window_set_hidden(nox_win_unk1, 0);
		}
	} else if (dword_5d4594_1046864) {
		result = sub_47A260();
		if (!result) {
			result = nox_window_set_hidden(nox_win_unk1, 1);
		}
	}
	return result;
}

static uint32_t* game2_book_old_sub_45D550(uint32_t* a1) {
	uint32_t* result; // eax

	result = a1;
	if (a1) {
		result = (uint32_t*)nox_client_wndGetPosition_46AA60(dword_5d4594_1046952, a1, a1 + 1);
	}
	return result;
}

static void game2_book_old_nox_xxx_bookFillAll_45D570(int a1, int a2) {
	int result; // eax
	double v3;  // st7
	int2 a4;    // [esp+Ch] [ebp-10h]
	int2 a3;    // [esp+14h] [ebp-8h]

	result = dword_5d4594_1047520;
	if (dword_5d4594_1047520 != 1) {
		nox_xxx_quickBarClose_4606B0();
		*getMemU32Ptr(0x5D4594, 1046612) = nox_xxx_buttonsGetSelectedRow_45E180();
		if (a1 != 2 || (result = nox_xxx_spellHasFlags_424A50(a2, 0x15000)) == 0) {
			result = nox_xxx_buttonHaveSpellInBarMB_4612D0(a2);
			if (result != 1) {
				if (a1 == 2 || a1 == 4) {
					result = nox_xxx_buttonFindFirstEmptySlot_461250();
				} else {
					if (a1 != 3) {
						return;
					}
					result = sub_4612A0();
				}
				dword_5d4594_1046852 = result;
				if (result != -1) {
					dword_5d4594_1047520 = 1;
					if (a1 == 2 || a1 == 3) {
						dword_5d4594_1046868 = 0;
						dword_5d4594_1046872 = 0;
					} else {
						dword_5d4594_1046868 = 1;
						dword_5d4594_1046872 = 1;
					}
					game2_book_old_sub_45D550(getMemAt(0x5D4594, 1046844));
					*(float*)&dword_5d4594_1046636 = (double)*getMemIntPtr(0x5D4594, 1046844);
					*(float*)&dword_5d4594_1046640 = (double)*getMemIntPtr(0x5D4594, 1046848);
					nox_window_setPos_46A9B0(dword_5d4594_1046956, *getMemIntPtr(0x5D4594, 1046844),
											 *getMemIntPtr(0x5D4594, 1046848));
					nox_xxx_buttonSetImgMB_461320(*(int*)&dword_5d4594_1046852, getMemAt(0x5D4594, 1046668));
					a3.field_0 = 400;
					a4.field_4 = 400;
					dword_5d4594_1047524 = a2;
					dword_5d4594_1046652 = a1 == 3;
					*getMemU32Ptr(0x5D4594, 1046676) = a1;
					nox_client_setRenderGUI(1);
					a3.field_4 = -500;
					a4.field_0 = 350;
					*getMemU32Ptr(0x5D4594, 1046680) = 0;
					sub_4BEDE0((int2*)getMemAt(0x5D4594, 1046844), (int2*)getMemAt(0x5D4594, 1046668), &a3, &a4, 19,
							   0.0, (uintptr_t)game2_book_old_sub_45D7D0, NULL);
					*getMemU32Ptr(0x5D4594, 1046628) = 0;
					obj_5d4594_1046620.field_0 =
						*getMemFloatPtr(0x5D4594, 1046692) - *getMemFloatPtr(0x5D4594, 1046684);
					obj_5d4594_1046620.field_4 =
						*getMemFloatPtr(0x5D4594, 1046696) - *getMemFloatPtr(0x5D4594, 1046688);
					nox_xxx_utilNormalizeVector_509F20(&obj_5d4594_1046620);
					if (nox_win_width < 1000) {
						if (nox_win_width < 750) {
							v3 = 6.0;
						} else {
							v3 = 8.0;
						}
					} else {
						v3 = 10.0;
					}
					obj_5d4594_1046620.field_0 = obj_5d4594_1046620.field_0 * v3;
					obj_5d4594_1046620.field_4 = obj_5d4594_1046620.field_4 * v3;
					nox_window_set_hidden(dword_5d4594_1046956, 0);
					nox_xxx_wndShowModalMB_46A8C0(dword_5d4594_1046956);
					if (nox_common_gameFlags_check_40A5C0(2048)) {
						sub_57AF30(0, a1);
					}
					dword_5d4594_1046648 = nox_xxx_bookGet_430B40_get_mouse_prev_seq();
					if (!nox_common_gameFlags_check_40A5C0(2048) || nox_gui_xxx_check_446360() == 1 ||
						(result = nox_xxx_gameGet_4DB1B0(), result == 1)) {
						sub_45D870();
					}
				}
			}
		}
	}
}

static int game2_book_old_sub_45D7D0(int* a1, int* a2) {
	int result; // eax
	double v3;  // st7

	result = *getMemU32Ptr(0x5D4594, 1046680);
	if (*getMemIntPtr(0x5D4594, 1046680) < 20) {
		result = *getMemU32Ptr(0x5D4594, 1046680) + 1;
		*getMemFloatPtr(0x5D4594, 1046676 + 8 * result) = (double)*a1;
		*getMemFloatPtr(0x5D4594, 1046680 + 8 * result) = (double)a1[1];
		*getMemFloatPtr(0x5D4594, 1046684 + 8 * result) = (double)*a2;
		v3 = (double)a2[1];
		*getMemU32Ptr(0x5D4594, 1046680) = result;
		*getMemFloatPtr(0x5D4594, 1046688 + 8 * result) = v3;
	}
	return result;
}

static void game2_book_old_sub_45D810() {
	if (dword_5d4594_1047520) {
		dword_5d4594_1047520 = 0;
		nox_window_set_hidden(dword_5d4594_1046956, 1);
		dword_5d4594_1046648 = 0;
		nox_xxx_clientUpdateButtonRow_45E110(*getMemIntPtr(0x5D4594, 1046612));
		if (nox_common_gameFlags_check_40A5C0(2048)) {
			sub_57B0A0();
			sub_413A00(0);
		}
	}
}

static int game2_book_old_sub_45D9B0() { return dword_5d4594_1047520; }

static int game2_book_old_nox_xxx_guiSpellTargetClickSet_45D9D0(int a1) {
	if (*getMemU32Ptr(0x5D4594, 1047916)) {
		return 0;
	}
	*getMemU32Ptr(0x5D4594, 1047916) = a1;
	*getMemU8Ptr(0x5D4594, 1047920) = 0;
	nox_xxx_setKeybTimeout_4160D0(5);
	*getMemU32Ptr(0x5D4594, 1047928) = 0;
	*getMemU32Ptr(0x5D4594, 1047924) = 0;
	return 1;
}

static int game2_book_old_nox_xxx_guiSpell_45DA10(int a1) {
	int result; // eax

	if (*getMemU32Ptr(0x5D4594, 1047916)) {
		return 0;
	}
	*getMemU32Ptr(0x5D4594, 1047916) = a1;
	*getMemU8Ptr(0x5D4594, 1047920) = 0;
	nox_xxx_setKeybTimeout_4160D0(5);
	result = 1;
	*getMemU32Ptr(0x5D4594, 1047924) = 1;
	return result;
}

static void game2_book_old_nox_client_invokeSpellSlot_45DA50(int a1) {
	nox_playerInfo* player = (nox_playerInfo*)dword_8531A0_2576;
	if (!player || nox_xxx_playerAnimCheck_4372B0()) {
		return;
	}
	unsigned char* slot = game2_book_old_nox_quickbar_selected_slot(nox_xxx_aClosewoodengat_587000_133480, a1);
	if (!slot) {
		return;
	}
	if (player->info.playerClass) {
		game2_book_old_nox_xxx_clientSendSpell_45DB20((char*)slot, 1, slot[4] & 1);
	} else {
		game2_book_old_nox_xxx_clientSendAbil_45DAF0(*(uint32_t*)slot);
	}
	game2_book_old_nox_xxx_clientStoreLastButton_45DAD0(a1);
}

static void game2_book_old_nox_xxx_clientStoreLastButton_45DAD0(int a1) {
	*getMemU32Ptr(0x587000, 133484) = a1;
	*getMemU32Ptr(0x5D4594, 1049540) = gameFrame();
}

static int game2_book_old_nox_xxx_clientSendAbil_45DAF0(int a1) {
	int result; // eax
	short v3;   // [esp+0h] [ebp-2h]

	result = nox_xxx_guiCursor_477600();
	if (!result) {
		result = a1;
		if (a1) {
			HIBYTE(v3) = a1;
			LOBYTE(v3) = 122;
			result = nox_netlist_addToMsgListCli_40EBC0(31, 0, &v3, 2);
		}
	}
	return result;
}

static int game2_book_old_nox_xxx_clientSendSpell_45DB20(char* a1, int a2, char a3) {
	int result;  // eax
	int v4;      // ecx
	char* v5;    // eax
	char v6[22]; // [esp+0h] [ebp-18h]

	result = nox_xxx_guiCursor_477600();
	if (!result && *(uint32_t*)a1) {
		v4 = 0;
		v6[0] = 121;
		v5 = &v6[1];
		do {
			if (v4 >= a2) {
				*(uint32_t*)v5 = 0;
			} else {
				*(uint32_t*)v5 = *(uint32_t*)&v5[a1 - &v6[1]];
			}
			++v4;
			v5 += 4;
		} while (v4 < 5);
		v6[21] = a3;
		result = nox_netlist_addToMsgListCli_40EBC0(31, 0, (unsigned char*)v6, 22);
	}
	return result;
}

static int game2_book_old_sub_45DB90() {
	int result; // eax

	result = 0;
	memset(getMemAt(0x5D4594, 1049544), 0, 0x88u);
	*getMemU8Ptr(0x5D4594, 1049680) = 0;
	return result;
}

static void game2_book_old_nox_xxx_guiSpellTargetClickCheckSend_45DBB0() {
	if (*getMemU32Ptr(0x5D4594, 1047916)) {
		game2_book_old_nox_xxx_clientSendSpell_45DB20((char*)getMemAt(0x5D4594, 1047916), 1, getMemByte(0x5D4594, 1047924));
		*getMemU32Ptr(0x5D4594, 1047916) = 0;
	}
}

static unsigned char* game2_book_old_nox_quickbar_selected_row(void* base) {
	if (!base) {
		return 0;
	}
	unsigned char* data = base;
	unsigned int row = data[200];
	if (row >= 5) {
		return 0;
	}
	return data + 40 * row;
}

static unsigned char* game2_book_old_nox_quickbar_trap_selected_row(void) {
	unsigned int row = getMemByte(0x5D4594, 1048140);
	if (row >= 3) {
		return 0;
	}
	return getMemAt(0x5D4594, 1047940 + 40 * row);
}

static unsigned char* game2_book_old_nox_quickbar_selected_slot(void* base, int slot) {
	if (slot < 0 || slot >= 5) {
		return 0;
	}
	unsigned char* row = game2_book_old_nox_quickbar_selected_row(base);
	return row ? row + 8 * slot : 0;
}

#undef nox_xxx_guiCursor_477600

#undef nox_netlist_addToMsgListCli_40EBC0

#endif
